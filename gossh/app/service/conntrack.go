package service

import (
	"bufio"
	"os"
	"sort"
	"strconv"
	"strings"

	"gossh/app/utils"
	"gossh/gin"
)

// ConnTrackSource 单个源 IP 的连接占用（含设备名称）
type ConnTrackSource struct {
	IP    string `json:"ip"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ConnTrackData /api/sys/conntrack 返回结构
type ConnTrackData struct {
	Count        int                `json:"count"`          // 当前连接表条目数
	Max          int                `json:"max"`            // 连接表上限
	UsagePercent int                `json:"usage_percent"`  // 占用率(0-100)
	Protocols    map[string]int     `json:"protocols"`      // 协议分布 (udp/tcp/icmp...)
	TopSources   []ConnTrackSource  `json:"top_sources"`    // 占用最高的 LAN 设备(已解析设备名)
	NonLanCount  int                `json:"non_lan_count"`  // 占用表但不在 LAN 设备列表的连接数(CPE 自身/外网/链路本地)
}

// ConntrackHandler 读取内核 conntrack 连接表状态，并返回协议分布与占用最高的源设备。
// 设备名称来自 ubus zwrt_router.api router_lan_access_list（即后台 station_info 页面数据源）。
func ConntrackHandler(c *gin.Context) {
	data := ConnTrackData{Protocols: map[string]int{}}

	// 1) 连接表当前条目数与上限
	if b, err := os.ReadFile("/proc/sys/net/netfilter/nf_conntrack_count"); err == nil {
		data.Count, _ = strconv.Atoi(strings.TrimSpace(string(b)))
	}
	if b, err := os.ReadFile("/proc/sys/net/netfilter/nf_conntrack_max"); err == nil {
		data.Max, _ = strconv.Atoi(strings.TrimSpace(string(b)))
	}
	if data.Max > 0 {
		data.UsagePercent = data.Count * 100 / data.Max
	}

	// 2) 解析 /proc/net/nf_conntrack：协议分布 + 每个源 IP 的连接数
	//    每行格式: "ipv4  2 tcp 6 75 TIME_WAIT src=192.168.0.228 dst=... ..."
	//    第 3 字段为 L4 协议名；原始方向源地址为行内第一个 src=（LAN 设备 IP）。
	ipCounts := map[string]int{}
	protoCounts := map[string]int{}
	if f, err := os.Open("/proc/net/nf_conntrack"); err == nil {
		scanner := bufio.NewScanner(f)
		// conntrack 行可能很长（含多地址/标签），放大缓冲到 4MB
		scanner.Buffer(make([]byte, 0, 1024*1024), 4*1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				protoCounts[fields[2]]++
			}
			if idx := strings.Index(line, "src="); idx >= 0 {
				rest := line[idx+4:]
				if end := strings.IndexAny(rest, " \t"); end >= 0 {
					rest = rest[:end]
				}
				if rest != "" {
					ipCounts[rest]++
				}
			}
		}
		f.Close()
	}

	// 3) 取 LAN 设备名称映射 (IP -> hostname) — ubus 当 stations 还没刷新时返回空，
	//    这种情况下所有 src 都将被识别为非 LAN（前端显示"暂无可识别设备"），不会乱 fallback 成 IP。
	nameMap := map[string]string{}
	if res, err := utils.GetDataFromUbus("zwrt_router.api", "router_lan_access_list", map[string]interface{}{}); err == nil {
		if arr, ok := res["lan_access_list_info"].([]interface{}); ok {
			for _, item := range arr {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				ip, _ := m["ip_address"].(string)
				name, _ := m["hostname"].(string)
				if ip != "" {
					nameMap[ip] = name
				}
			}
		}
	}

	// 4) 取 TOP 源：只列"命中 LAN 设备"的 src。
	//    CPE 上行 IPv6 出向 / 运营商 IPv6 / 链路本地 (fe80::) 等不在 LAN 表中，归入 non_lan_count，
	//    避免出现"设备名称空，回退显示裸 IP"的体验问题。
	type kv struct {
		IP    string
		Count int
	}
	list := make([]kv, 0, len(ipCounts))
	nonLan := 0
	for ip, cnt := range ipCounts {
		if _, ok := nameMap[ip]; ok {
			list = append(list, kv{ip, cnt})
		} else {
			nonLan += cnt
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Count > list[j].Count })

	topN := 3
	if len(list) < topN {
		topN = len(list)
	}
	top := make([]ConnTrackSource, 0, topN)
	for i := 0; i < topN; i++ {
		ip := list[i].IP
		name := nameMap[ip]
		if name == "" {
			// 几乎不可能走到（前面已经按 nameMap 命中筛过），作兜底
			name = ip
		}
		top = append(top, ConnTrackSource{IP: ip, Name: name, Count: list[i].Count})
	}

	data.Protocols = protoCounts
	data.TopSources = top
	data.NonLanCount = nonLan

	c.JSON(200, gin.H{"code": 0, "data": data})
}

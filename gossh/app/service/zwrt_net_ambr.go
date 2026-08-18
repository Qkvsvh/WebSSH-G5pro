package service

import (
	"fmt"
	"log/slog"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"gossh/gin"
)

// NetAmbrGetHandler 返回 G5 Pro 的 QCI 与上下行协商速率(AMBR)
// 数据源: UCI zwrt_data_tmp.wwaniface1 (dl_ambr/ul_ambr 单位为 kbps, qci 为整型)
// 注意: U60 Pro 的 /data/logfs/key.log 在 G5 Pro 上不存在, 故改为读取 UCI。
func NetAmbrGetHandler(c *gin.Context) {
	qci, dlKbps, ulKbps, err := getAmbrFromUci()

	// ICCID / SIM卡号 与 AMBR 同源(设备后台"设备信息"区), 独立读取, 即使 AMBR 取失败也尽量返回
	iccid := getSimInfoField("iccid")
	simNumber := getSimInfoField("msisdn")

	if err != nil {
		// 优雅降级: 返回全 0, 前端卡片显示 N/A 而非报错
		slog.Warn("[API] /api/net/ambr/get 未取到 AMBR(可能设备不支持):", "err", err.Error())
		c.JSON(200, gin.H{
			"code": 0,
			"data": gin.H{
				"dl":       gin.H{"value": 0, "unit": "", "unit_num": 0, "unit_raw": ""},
				"ul":       gin.H{"value": 0, "unit": "", "unit_num": 0, "unit_raw": ""},
				"qci1":     0,
				"qci2":     0,
				"raw_ambr": "",
				"raw_qci":  "",
				"iccid":    iccid,
				"sim_number": simNumber,
			},
		})
		return
	}

	// kbps -> Mbps, 四舍五入到整百 Mbps
	// 与设备后台 #network_info 设备信息展示一致: 下行约 2000M / 上行约 200M
	dlMbps := math.Round(float64(dlKbps)/1000.0/100) * 100
	ulMbps := math.Round(float64(ulKbps)/1000.0/100) * 100

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"dl": gin.H{
				"value":    int(dlMbps),
				"unit":     "Mbps",
				"unit_num": 0,
				"unit_raw": fmt.Sprintf("%d kbps", dlKbps),
			},
			"ul": gin.H{
				"value":    int(ulMbps),
				"unit":     "Mbps",
				"unit_num": 0,
				"unit_raw": fmt.Sprintf("%d kbps", ulKbps),
			},
			"qci1":     qci,
			"qci2":     qci,
			"raw_ambr": fmt.Sprintf("dl=%d ul=%d kbps", dlKbps, ulKbps),
			"raw_qci":  strconv.Itoa(qci),
			"iccid":    iccid,
			"sim_number": simNumber,
		},
	})
}

// getSimInfoField 从 UCI zwrt_zte_mdm.sim_info.<key> 读取 SIM 相关字段
//   iccid  = SIM 卡 ICCID(设备实测 898601**********5440，已脱敏)
//   msisdn = SIM 卡号/手机号(设备实测 861******6400，已脱敏)
// 与设备后台"设备信息"区展示一致, 与 zwrt_zte_mdm.api get_sim_info 同源同值
func getSimInfoField(key string) string {
	out, e := exec.Command("sh", "-c",
		fmt.Sprintf("uci get zwrt_zte_mdm.sim_info.%s 2>/dev/null", key)).CombinedOutput()
	if e != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getAmbrFromUci 从 UCI zwrt_data_tmp.wwaniface1 读取 qci / dl_ambr / ul_ambr
func getAmbrFromUci() (qci int, dlKbps int, ulKbps int, err error) {
	get := func(key string) (int, error) {
		out, e := exec.Command("sh", "-c",
			fmt.Sprintf("uci get zwrt_data_tmp.wwaniface1.%s 2>/dev/null", key)).CombinedOutput()
		if e != nil {
			return 0, fmt.Errorf("uci get %s: %v (%s)", key, e, strings.TrimSpace(string(out)))
		}
		s := strings.TrimSpace(string(out))
		if s == "" {
			return 0, fmt.Errorf("uci %s 为空", key)
		}
		v, e2 := strconv.Atoi(s)
		if e2 != nil {
			return 0, fmt.Errorf("uci %s 非整数: %q", key, s)
		}
		return v, nil
	}

	qci, e1 := get("qci")
	dlKbps, e2 := get("dl_ambr")
	ulKbps, e3 := get("ul_ambr")
	if e1 != nil || e2 != nil || e3 != nil {
		return 0, 0, 0, fmt.Errorf("qci=%v dl=%v ul=%v", e1, e2, e3)
	}
	return qci, dlKbps, ulKbps, nil
}

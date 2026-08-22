package service

import (
	"log/slog"
	"net/http"
	"sync"

	"gossh/app/utils"
	"gossh/gin"
)

// =============================================================================
// 整机防火墙开关（zwrt_router.api router_set/get_firewall_switch）
// =============================================================================
//
// 背景：原本这两组路由在 main.go 中被整段注释（"已下线实现"），导致前端
//       /api/sys/firewall/status 永远拿不到状态，firewallStatus 恒为 null，
//       两个防火墙按钮（开启/关闭）因此一直 disabled（灰色）。
//
// 修复：补齐 status / set 两个 handler，并恢复路由注册。
//   - 设置：ubus call zwrt_router.api router_set_firewall_switch '{"enable":0|1}'
//   - 查询：ubus call zwrt_router.api router_get_firewall_switch（set 的自然对应方法）
//
// 健壮性设计：
//   - 查询返回形态兼容 {"enable":1} / {"enabled":true} / {"enable":"1"} 等
//     （JSON 数值解出来是 float64，字符串/布尔也都支持）。
//   - 用内存变量缓存"最近一次成功 set 的状态"，当 ubus get 在设备上不可用
//     （方法名差异/版本不同）时回退到缓存值，避免前端因拿不到状态而重新置灰。

const fwUbusService = "zwrt_router.api"

var (
	fwMu        sync.RWMutex
	fwLastKnown *bool // 最近一次成功 set 后的状态；nil 表示从未成功设置过
)

// FirewallSetHandler 开关整机防火墙。
// 请求: POST /api/sys/firewall/set  body: {"enable": true|false}
func FirewallSetHandler(c *gin.Context) {
	var req struct {
		Enable bool `json:"enable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}

	enableInt := 0
	if req.Enable {
		enableInt = 1
	}
	slog.Info("[API] /api/sys/firewall/set 调用开始", "enable", req.Enable)

	res, err := utils.GetDataFromUbus(fwUbusService, "router_set_firewall_switch",
		map[string]interface{}{"enable": enableInt})
	if err != nil {
		slog.Error("[API] router_set_firewall_switch 失败", "err", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "防火墙设置失败: " + err.Error(),
		})
		return
	}

	// 缓存最近一次成功状态（get 不可用时的兜底来源）
	fwMu.Lock()
	v := req.Enable
	fwLastKnown = &v
	fwMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"code":   0,
		"data":   gin.H{"known": true, "enabled": req.Enable},
		"msg":    "防火墙已" + ternary(req.Enable, "开启", "关闭"),
		"ubus":   res,
	})
}

// FirewallStatusHandler 查询整机防火墙开关状态。
// 请求: GET /api/sys/firewall/status
// 返回: {"code":0,"data":{"known":true,"enabled":bool}} 或 {"code":0,"data":{"known":false}}
func FirewallStatusHandler(c *gin.Context) {
	enabled, ok, err := queryFirewallEnabled()
	if err != nil {
		slog.Warn("[API] /api/sys/firewall/status 获取失败，尝试回退缓存", "err", err.Error())
	}
	if ok {
		// 查询成功，更新缓存
		fwMu.Lock()
		v := enabled
		fwLastKnown = &v
		fwMu.Unlock()
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{"known": true, "enabled": enabled},
		})
		return
	}

	// ubus get 失败/不可用：回退到缓存的最近状态（至少一次成功 set 后才可能有）
	fwMu.RLock()
	cached := fwLastKnown
	fwMu.RUnlock()
	if cached != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{"known": true, "enabled": *cached, "cached": true},
		})
		return
	}

	// 既查不到也没有任何缓存：如实告知未知（前端此时两个按钮保持禁用）
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"known": false},
	})
}

// queryFirewallEnabled 调用 ubus router_get_firewall_switch 并解析 enabled 状态。
func queryFirewallEnabled() (enabled bool, ok bool, err error) {
	res, e := utils.GetDataFromUbus(fwUbusService, "router_get_firewall_switch", nil)
	if e != nil {
		return false, false, e
	}
	raw, present := pickEnabledField(res)
	if !present {
		// 返回里没有 enable/enabled 字段，视为无法判定
		return false, false, nil
	}
	return parseBoolLike(raw), true, nil
}

// pickEnabledField 兼容 enabled / enable 两种键名。
func pickEnabledField(m map[string]interface{}) (interface{}, bool) {
	for _, key := range []string{"enabled", "enable"} {
		if v, ok := m[key]; ok {
			return v, true
		}
	}
	return nil, false
}

// parseBoolLike 把各种"像布尔"的值归一为 bool。
func parseBoolLike(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64: // JSON 数值默认解成 float64
		return t != 0
	case int:
		return t != 0
	case int64:
		return t != 0
	case string:
		return t == "1" || t == "true" || t == "on" || t == "yes"
	default:
		return false
	}
}

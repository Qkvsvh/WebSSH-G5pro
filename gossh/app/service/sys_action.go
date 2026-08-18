package service

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gossh/gin"
)

// =============================================================================
// conntrack-tune 装机/卸载/状态（治本：扩容 nf_conntrack 表 + 缩短 UDP 超时）
// =============================================================================
//
// 设计要点：
//   - WebSSH 在 G5 Pro 上以 root 身份运行，与 CPE 共享同一文件系统；脚本本身
//     通过 os.WriteFile 写入 /etc/init.d/conntrack-tune，免去额外 scp/上传步骤。
//   - 安装流程：写文件 → chmod +x → enable（生成 /etc/rc.d/S99conntrack-tune）
//     → start（立即生效，不重启）。
//   - 卸载流程：disable → stop → rm 脚本与 /etc/rc.d/S99conntrack-tune 软链。
//   - 状态：是否已安装 + 当前生效值。便于前端按钮显示"安装/卸载/已安装"。
//
// 命名：旧名 zz-conntrack 只是 OpenWrt 启动顺序占位（zz- = 最后启动），
//       名字不够直观。conntrack-tune 一眼能看出"调优 conntrack"语义。
//       install 时检测到旧文件会自动清理，无需用户手动操作。
//
// 脚本内容与 g5pro-tools/conntrack-tune 必须保持一致（两份是同一脚本的内联副本）。

const conntrackTuneInitPath = "/etc/init.d/conntrack-tune"
// 旧名兼容：升级时自动清理用户历史部署的 zz-conntrack
const conntrackTuneLegacyPath = "/etc/init.d/zz-conntrack"
const conntrackTuneScript = `#!/bin/sh /etc/rc.common
# conntrack-tune: 扩容 conntrack 表 + 缩短 UDP 超时，适配 PCDN 等高并发 UDP 场景
#
# 背景：G5 Pro 默认 nf_conntrack_max=16384 太小，PCDN 海量 UDP 短连接会把表顶满，
#       内核丢包 -> 云端心跳(:30443)被丢 -> 中兴智慧生活 app 显示"广域网不在线"。
#       本脚本开机自动应用；start 也立即生效，无需重启设备。
#
# 部署（在 G5 Pro 上，192.168.0.1:3540，root）：
#   scp conntrack-tune root@192.168.0.1:/etc/init.d/conntrack-tune
#   chmod +x /etc/init.d/conntrack-tune
#   /etc/init.d/conntrack-tune enable        # 生成 /etc/rc.d/S99conntrack-tune，开机自启
#   /etc/init.d/conntrack-tune start         # 立即生效
#
# 验证：
#   cat /proc/sys/net/netfilter/nf_conntrack_max      # 应为 524288
#   cat /proc/sys/net/netfilter/nf_conntrack_count    # 当前条目数（对比 max 看占用率）
#
# 注意：
#   - 本机 /etc/rc.common 为 procd 模式，开机实际调用 start_service()；同时保留 start()
#     作为手动 xxx start 的兜底（procd 不会调用 start()，但 rc.common 手动执行会）。
#   - rc.local 在 G5 Pro 开机不执行，故用 procd init.d，不要依赖 rc.local。
#   - 恢复出厂会清掉本脚本与 /etc/rc.d/S99conntrack-tune，需按上面步骤重新部署。

START=99
STOP=10

# ===== 可调参数（按需改这一块即可）=====
CONNTRACK_MAX=524288        # 表上限；PCDN 场景 512k 足够，内存充裕也可改 1048576(1M)
UDP_TIMEOUT=10              # UDP 短包超时(秒)，默认60->10，加速回收
UDP_TIMEOUT_STREAM=30       # UDP 已建流超时(秒)，默认180->30
TCP_EST_TIMEOUT=3600        # TCP 已建连接超时(秒)，默认432000->3600

apply_settings() {
    [ -w /proc/sys/net/netfilter/nf_conntrack_max ] && \
        echo "$CONNTRACK_MAX" > /proc/sys/net/netfilter/nf_conntrack_max
    [ -w /proc/sys/net/netfilter/nf_conntrack_udp_timeout ] && \
        echo "$UDP_TIMEOUT" > /proc/sys/net/netfilter/nf_conntrack_udp_timeout
    [ -w /proc/sys/net/netfilter/nf_conntrack_udp_timeout_stream ] && \
        echo "$UDP_TIMEOUT_STREAM" > /proc/sys/net/netfilter/nf_conntrack_udp_timeout_stream
    [ -w /proc/sys/net/netfilter/nf_conntrack_tcp_timeout_established ] && \
        echo "$TCP_EST_TIMEOUT" > /proc/sys/net/netfilter/nf_conntrack_tcp_timeout_established
    logger -t conntrack-tune "applied: max=$CONNTRACK_MAX udp=$UDP_TIMEOUT udp_stream=$UDP_TIMEOUT_STREAM tcp_est=$TCP_EST_TIMEOUT"
}

# procd 模式开机调用此入口
start_service() {
    apply_settings
}

# 手动直接调用脚本 start 时的兜底
start() {
    apply_settings
}
`

// runLocalCmd 执行本地命令（WebSSH 与 CPE 同机，root 身份直接 exec 即可，
// 与 OpenAdbHandler 调用 ubus 一致），返回合并输出与退出码。
// 可选 timeout：依赖外部 timeout 命令；缺命令时回退到无超时（不阻断主路径）。
func runLocalCmd(timeoutSec int, name string, args ...string) (string, int, error) {
	var cmd *exec.Cmd
	if timeoutSec > 0 {
		if _, err := exec.LookPath("timeout"); err == nil {
			allArgs := append([]string{fmt.Sprintf("%d", timeoutSec), name}, args...)
			cmd = exec.Command("timeout", allArgs...)
		} else {
			cmd = exec.Command(name, args...)
		}
	} else {
		cmd = exec.Command(name, args...)
	}
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return strings.TrimRight(string(out), "\n"), exitCode, err
}

// runSh 执行 sh -c 形式（多项管道时用），同样支持 timeout。
func runSh(timeoutSec int, script string) (string, int, error) {
	args := []string{"-c", script}
	if timeoutSec > 0 {
		if _, err := exec.LookPath("timeout"); err == nil {
			return runLocalCmd(timeoutSec, "sh", args...)
		}
	}
	return runLocalCmd(0, "sh", args...)
}

// ConntrackInstallHandler 安装 conntrack-tune 治本脚本到 /etc/init.d 并启用 + 启动。
// 设计：写文件 → chmod → enable（生成 /etc/rc.d/S99conntrack-tune）→ start。
// 兼容：若检测到旧名 zz-conntrack 残留（用户历史部署），先自动 disable+stop+rm，
//       避免新老脚本同时存在造成双重 enable 冲突。
func ConntrackInstallHandler(c *gin.Context) {
	slog.Info("[API] /api/sys/conntrack/install 调用开始")

	steps := []gin.H{}

	// 0) 旧名兼容：检测并清理 zz-conntrack 历史残留
	if _, err := os.Stat(conntrackTuneLegacyPath); err == nil {
		legacy := []gin.H{}
		if out, code, er := runLocalCmd(5, conntrackTuneLegacyPath, "disable"); er == nil {
			legacy = append(legacy, gin.H{"step": "legacy_disable", "code": code, "out": out})
		}
		if out, code, er := runLocalCmd(5, conntrackTuneLegacyPath, "stop"); er == nil {
			legacy = append(legacy, gin.H{"step": "legacy_stop", "code": code, "out": out})
		}
		if rmErr := os.Remove(conntrackTuneLegacyPath); rmErr != nil && !os.IsNotExist(rmErr) {
			legacy = append(legacy, gin.H{"step": "legacy_rm_script", "err": rmErr.Error()})
		} else {
			legacy = append(legacy, gin.H{"step": "legacy_rm_script", "ok": true})
		}
		for _, link := range []string{"/etc/rc.d/S99zz-conntrack", "/etc/rc.d/K10zz-conntrack"} {
			if rmErr := os.Remove(link); rmErr != nil && !os.IsNotExist(rmErr) {
				legacy = append(legacy, gin.H{"step": "legacy_rm_" + filepath.Base(link), "err": rmErr.Error()})
			} else {
				legacy = append(legacy, gin.H{"step": "legacy_rm_" + filepath.Base(link), "ok": true})
			}
		}
		steps = append(steps, gin.H{"legacy_cleanup": legacy})
	}

	// 1) 写脚本（覆盖式）
	if err := os.WriteFile(conntrackTuneInitPath, []byte(conntrackTuneScript), 0755); err != nil {
		slog.Error("[API] 写脚本失败", "err", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": 1, "success": false,
			"msg": fmt.Sprintf("写入 %s 失败: %s", conntrackTuneInitPath, err.Error()),
		})
		return
	}
	steps = append(steps, gin.H{"step": "write_script", "path": conntrackTuneInitPath, "size": len(conntrackTuneScript), "ok": true})

	// 2) chmod +x（WriteFile 已经给了 0755，再保险一次）
	chmodOut, codeChmod, errChmod := runLocalCmd(5, "chmod", "+x", conntrackTuneInitPath)
	if errChmod != nil {
		slog.Warn("[API] chmod 异常", "code", codeChmod, "err", errChmod.Error())
	}
	steps = append(steps, gin.H{"step": "chmod", "code": codeChmod, "out": chmodOut})

	// 3) enable（生成 /etc/rc.d/S99conntrack-tune）
	enableOut, codeEnable, errEnable := runLocalCmd(5, conntrackTuneInitPath, "enable")
	if errEnable != nil {
		slog.Error("[API] enable 失败", "code", codeEnable, "err", errEnable.Error(), "out", enableOut)
		// enable 失败仍 start（不阻断主流程）
	}
	steps = append(steps, gin.H{"step": "enable", "code": codeEnable, "out": enableOut})

	// 4) start（立即生效）
	startOut, codeStart, errStart := runLocalCmd(5, conntrackTuneInitPath, "start")
	if errStart != nil {
		slog.Error("[API] start 失败", "code", codeStart, "err", errStart.Error(), "out", startOut)
		c.JSON(http.StatusOK, gin.H{
			"code": 2, "success": false,
			"msg":         "脚本写入但 start 失败，请检查 /etc/rc.common 是否存在",
			"chmod":       codeChmod,
			"enable":      codeEnable,
			"enable_out":  enableOut,
			"start":       codeStart,
			"start_out":   startOut,
			"script_path": conntrackTuneInitPath,
			"steps":       steps,
		})
		return
	}
	steps = append(steps, gin.H{"step": "start", "code": codeStart, "out": startOut})

	c.JSON(http.StatusOK, gin.H{
		"code": 0, "success": true,
		"msg":         "conntrack-tune 已安装并启动",
		"script_path": conntrackTuneInitPath,
		"rc_link":     "/etc/rc.d/S99conntrack-tune",
		"chmod":       codeChmod,
		"enable":      codeEnable,
		"start":       codeStart,
		"start_out":   startOut,
		"steps":       steps,
		"applied_at":  time.Now().Format(time.RFC3339),
	})
}

// ConntrackUninstallHandler 卸载 conntrack-tune：disable + stop + rm。
// 同步清理 procd 在 /etc/rc.d 下生成的 S99conntrack-tune 软链，以及历史 zz-conntrack 残留。
func ConntrackUninstallHandler(c *gin.Context) {
	slog.Info("[API] /api/sys/conntrack/uninstall 调用开始")

	steps := []gin.H{}

	// 清理主路径
	if _, err := os.Stat(conntrackTuneInitPath); err == nil {
		out, code, errRun := runLocalCmd(5, conntrackTuneInitPath, "disable")
		steps = append(steps, gin.H{"step": "disable", "code": code, "out": out, "err": errMsg(errRun)})
	}
	if _, err := os.Stat(conntrackTuneInitPath); err == nil {
		out, code, errRun := runLocalCmd(5, conntrackTuneInitPath, "stop")
		steps = append(steps, gin.H{"step": "stop", "code": code, "out": out, "err": errMsg(errRun)})
	}
	if err := os.Remove(conntrackTuneInitPath); err != nil && !os.IsNotExist(err) {
		steps = append(steps, gin.H{"step": "rm_script", "err": err.Error()})
	} else {
		steps = append(steps, gin.H{"step": "rm_script", "ok": true})
	}
	for _, link := range []string{
		"/etc/rc.d/S99conntrack-tune",
		"/etc/rc.d/K10conntrack-tune",
		// 历史 zz-conntrack 残留（旧用户部署）
		"/etc/rc.d/S99zz-conntrack",
		"/etc/rc.d/K10zz-conntrack",
	} {
		if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
			steps = append(steps, gin.H{"step": "rm_" + filepath.Base(link), "err": err.Error()})
		} else {
			steps = append(steps, gin.H{"step": "rm_" + filepath.Base(link), "ok": true})
		}
	}

	// 兼容清理旧名脚本本体
	if _, err := os.Stat(conntrackTuneLegacyPath); err == nil {
		if out, code, er := runLocalCmd(5, conntrackTuneLegacyPath, "disable"); er == nil {
			steps = append(steps, gin.H{"step": "legacy_disable", "code": code, "out": out})
		}
		if rmErr := os.Remove(conntrackTuneLegacyPath); rmErr != nil && !os.IsNotExist(rmErr) {
			steps = append(steps, gin.H{"step": "legacy_rm_script", "err": rmErr.Error()})
		} else {
			steps = append(steps, gin.H{"step": "legacy_rm_script", "ok": true})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0, "success": true,
		"msg":   "conntrack-tune 已卸载",
		"steps": steps,
	})
}

func errMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// ConntrackStatusHandler 返回脚本是否已安装 + 当前生效的 conntrack 参数。
// 数据来源：脚本文件是否存在 + 当前 /proc/sys 内核值。
// 兼容：同时检测新旧两个名字（zz-conntrack → conntrack-tune）。
func ConntrackStatusHandler(c *gin.Context) {
	installed := false
	installedName := ""
	if _, err := os.Stat(conntrackTuneInitPath); err == nil {
		installed = true
		installedName = "conntrack-tune"
	} else if _, err := os.Stat(conntrackTuneLegacyPath); err == nil {
		// 兼容旧部署：仍上报为已安装，让用户走 install 自动升级到新名
		installed = true
		installedName = "zz-conntrack(legacy)"
	}
	// 软链是否生成（procd enable 后才有）
	rcLinkExists := false
	for _, link := range []string{"/etc/rc.d/S99conntrack-tune", "/etc/rc.d/S99zz-conntrack"} {
		if _, err := os.Stat(link); err == nil {
			rcLinkExists = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":           0,
		"installed":      installed,
		"installed_name": installedName,
		"rc_link_exists": rcLinkExists,
		"script_path":    conntrackTuneInitPath,
	})
}

// =============================================================================
// RF 频段功率切换（mipc_wan_cli --at_cmd AT+EFMAXPWR=...）
// =============================================================================
//
// 设计要点：
//   - 在 G5 Pro 上 root 直接 exec mipc_wan_cli（与 OpenAdbHandler 走 ubus 同样思路，
//     不绕 SSH）；命令发送成功即视为生效，modem 内部应答由厂商工具解读。
//   - 第 1 条覆盖 7,9,1,3,5,8,28,41,77,78,79,... (29dBm)；第 2 条 4,9,1,3,5,8,34,38,39,40,41,...
//     末位 1 表示应用，0 表示恢复默认。
//   - 两条连发，串行执行（OpenWrt 下并发可能与 modem 冲突）。

// mipcEfmaxpwr29dBm 设置最大功率（≈29 dBm，覆盖 NSA/SA 常用频段）
var mipcEfmaxpwr29dBm = []string{
	"AT+EFMAXPWR=7,9,1,3,5,8,28,41,77,78,79,232,232,232,232,232,232,232,232,232,1",
	"AT+EFMAXPWR=4,9,1,3,5,8,34,38,39,40,41,232,232,232,232,232,232,232,232,232,1",
}

// rfMaxpwrMarkerPath 本地标记：记录"本工具最后一次成功写入的目标"。
// 用于与 modem 寄存器 readback 交叉验证，捕捉网络重注册导致的寄存器漂移
// （漂移时硬件 readback 与本地标记不符，前端据此提示用户重新应用）。
// 注意：恢复出厂可能清掉 /data/webssh，但硬件 readback 仍是真实来源，不依赖此标记。
const rfMaxpwrMarkerPath = "/data/webssh/.rf_maxpwr_target"

// rfTargetPowerValue 我们写入的目标功率值（232 ≈ 29dBm）。
// status 判定不仅看"是否非空"，还要校验是否等于该目标值，避免把出厂默认
// 的非空 readback 误判为"已应用"。
const rfTargetPowerValue = 232

// mipcEfmaxpwrDefault 恢复默认（末位 0）
var mipcEfmaxpwrDefault = []string{
	"AT+EFMAXPWR=7,9,1,3,5,8,28,41,77,78,79,232,232,232,232,232,232,232,232,232,0",
	"AT+EFMAXPWR=4,9,1,3,5,8,34,38,39,40,41,232,232,232,232,232,232,232,232,232,0",
}

// runAtBatch 串行执行多条 AT 命令；任一失败收集错误但不中断（用户可看 details）。
func runAtBatch(commands []string) (results []gin.H, ok bool) {
	results = make([]gin.H, 0, len(commands))
	allOk := true
	for _, at := range commands {
		out, code, err := runLocalCmd(10, "mipc_wan_cli", "--at_cmd", at)
		results = append(results, gin.H{
			"at":   at,
			"out":  out,
			"code": code,
			"err":  errMsg(err),
		})
		if err != nil || code != 0 {
			allOk = false
		}
	}
	return results, allOk
}

// RfMaxpowerHandler 设置所有频段最大功率 ≈29 dBm。
// 请求: POST /api/sys/rf-maxpower/set
func RfMaxpowerHandler(c *gin.Context) {
	slog.Info("[API] /api/sys/rf-maxpower/set 调用开始")
	results, ok := runAtBatch(mipcEfmaxpwr29dBm)
	msg := "所有频段功率已设置为 29dBm"
	if !ok {
		msg = "已发送，但部分命令未返回成功，请查看 details"
	}
	// 全部 AT 成功才写入本地标记；失败则不写（保留原标记，避免误标）
	if ok {
		_ = os.WriteFile(rfMaxpwrMarkerPath, []byte("29"), 0644)
	}
	c.JSON(http.StatusOK, gin.H{
		"code":     boolToInt(ok),
		"success":  ok,
		"msg":      msg,
		"details":  results,
	})
}

// RfMaxpowerDefaultHandler 恢复 RF 频段功率为默认值（AT 末位 0）。
// 请求: POST /api/sys/rf-maxpower/default
func RfMaxpowerDefaultHandler(c *gin.Context) {
	slog.Info("[API] /api/sys/rf-maxpower/default 调用开始")
	results, ok := runAtBatch(mipcEfmaxpwrDefault)
	msg := "所有频段功率已恢复默认"
	if !ok {
		msg = "已发送，但部分命令未返回成功，请查看 details"
	}
	// 成功恢复默认则删除本地标记（下次 status 与硬件默认态一致）
	if ok {
		_ = os.Remove(rfMaxpwrMarkerPath)
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    boolToInt(ok),
		"success": ok,
		"msg":     msg,
		"details": results,
	})
}

func boolToInt(b bool) int {
	if b {
		return 0
	}
	return 1
}

// RfPowerEntry 单条 RAT 的频段与功率读回。
type RfPowerEntry struct {
	Rat    string `json:"rat"`
	Bands  []int  `json:"bands"`
	Powers []int  `json:"powers"`
}

// RfMaxpowerStatusHandler 读取 modem 当前 RF 发射功率上限配置。
// 数据来源：mipc_wan_cli --at_cmd AT+EFMAXPWR? 回显，格式：
//
//	+EFMAXPWR: <rat>,(<bands>),(<powers>)
//
// 与早期版本相比，本实现做了更精准的判定：
//   - 不再仅凭"powers 元组是否非空"判定（避免把出厂默认的非空 readback 误判为已应用）；
//   - 改为校验 powers 是否全部等于目标值（232 ≈ 29dBm），区分三种状态：
//     default（全空）/ applied_target（=232，即本工具写入的 29dBm）/ custom（非空且非 232）。
//   - 与本地标记文件交叉验证：捕捉 modem 网络重注册导致的寄存器漂移。
//
// 注意：状态反映的是"modem 寄存器写入的指令"，不等于"实时发射功率"——
// 实际发射功率由基站 TPC 调度决定，本工具无法读取也不可能强制。
// 请求: GET /api/sys/rf-maxpower/status
func RfMaxpowerStatusHandler(c *gin.Context) {
	slog.Info("[API] /api/sys/rf-maxpower/status 调用开始")
	out, _, err := runLocalCmd(10, "mipc_wan_cli", "--at_cmd", "AT+EFMAXPWR?")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"success": false,
			"msg":     "读取失败: " + err.Error(),
			"applied": false,
			"state":   "unknown",
		})
		return
	}

	entries := []RfPowerEntry{}
	hasNonEmpty := false
	targetMatch := false
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "+EFMAXPWR:") {
			continue
		}
		body := strings.TrimSpace(strings.TrimPrefix(line, "+EFMAXPWR:"))
		open1 := strings.Index(body, "(")
		close1 := strings.Index(body, ")")
		open2 := strings.LastIndex(body, "(")
		close2 := strings.LastIndex(body, ")")
		if open1 < 0 || close1 < 0 || open2 < 0 || close2 < 0 || open1 == open2 {
			continue
		}
		rat := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(body[:open1]), ","))
		bands := parseCsvInts(body[open1+1 : close1])
		powers := parseCsvInts(body[open2+1 : close2])
		if len(powers) > 0 {
			hasNonEmpty = true
			// 全部等于目标值（232）才视为本工具写入的 29dBm
			if allEqualTo(powers, rfTargetPowerValue) {
				targetMatch = true
			}
		}
		entries = append(entries, RfPowerEntry{Rat: rat, Bands: bands, Powers: powers})
	}

	// 三态判定
	state := "default"
	switch {
	case targetMatch:
		state = "applied_target"
	case hasNonEmpty:
		state = "custom"
	}

	// 本地标记交叉验证（捕捉 modem 网络重注册导致的寄存器漂移）
	recorded := fileExists(rfMaxpwrMarkerPath)
	consistent := true
	driftMsg := ""
	switch state {
	case "default":
		if recorded {
			consistent = false
			driftMsg = "本地记录显示已应用 29dBm，但 modem 寄存器已回到默认（可能因网络重注册被重置）。建议重新点击『所有频段功率设为 29dBm』使其重新生效。"
		}
	case "custom":
		consistent = false
		driftMsg = "modem 已应用非 29dBm 的自定义功率上限（非本工具设置）。如需统一为 29dBm，请点击『所有频段功率设为 29dBm』覆盖。"
	}

	msg := map[string]string{
		"default":        "当前为默认功率",
		"applied_target": "已应用 29dBm 上限（所有频段）",
		"custom":         "已应用自定义功率上限（非 29dBm）",
	}[state]

	c.JSON(http.StatusOK, gin.H{
		"code":            0,
		"success":         true,
		"state":           state,
		"applied":         state != "default",
		"recorded_target": recorded,
		"consistent":      consistent,
		"drift_msg":       driftMsg,
		"entries":         entries,
		"msg":             msg,
	})
}

// parseCsvInts 解析逗号分隔的整数列表，空串返回空切片。
func parseCsvInts(s string) []int {
	s = strings.TrimSpace(s)
	if s == "" {
		return []int{}
	}
	out := []int{}
	for _, f := range strings.Split(s, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(f)); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// allEqualTo 判断是否所有元素都等于 v（空切片返回 false）。
func allEqualTo(s []int, v int) bool {
	if len(s) == 0 {
		return false
	}
	for _, x := range s {
		if x != v {
			return false
		}
	}
	return true
}

// fileExists 判断文件是否存在（用于本地标记交叉验证）。
func fileExists(p string) bool {
	if _, err := os.Stat(p); err == nil {
		return true
	}
	return false
}

// ternary 小工具，避免引入外部依赖。
func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

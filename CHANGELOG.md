# WebSSH for G5 Pro — 相对原版(o8oo8o/WebSSH)更新日志

> 本分支在 [o8oo8o/WebSSH](https://github.com/o8oo8o/WebSSH)（原作针对中兴 U60 Pro / ZTE MU5120 二改）基础上，全面适配 **中兴 G5 Pro CPE**（MT6990 / OpenWrt 21.02.7），并重构了前端面板与移动端体验。
> 本文档已脱敏：不含任何设备账号、密码、内网 IP 等敏感信息；设备默认管理地址按"设备默认管理 IP"表述。

---

## 〇、近期修复（2026-08-22）

- **修复防火墙开关置灰**：原 `main.go` 中整机防火墙的两条路由
  `GET /api/sys/firewall/status` 与 `POST /api/sys/firewall/set` 被整段注释
  （标注"已下线实现"），导致前端永远拿不到 `firewallStatus`，两个防火墙按钮
  恒为 `disabled`（灰色）。新增 `gossh/app/service/firewall.go` 实现两个 handler
  （设置走 `ubus call zwrt_router.api router_set_firewall_switch '{"enable":0|1}'`，
  查询走对应的 `router_get_firewall_switch`，并兼容 `enable`/`enabled` 键名与多种
  取值形态），并在 `main.go` 恢复路由注册。另用内存缓存兜底：当设备上的 get 方法
  不可用时，回退到最近一次成功 set 的状态，避免按钮再次置灰。
- **README 文档修正**：① 卸载说明与脚本行为矛盾——`g5pro-install.sh remove` 实际执行 `rm -rf /data/webssh`（连同配置 `.GoWebSSH` 一并删除），原文"保留配置目录"已改为如实描述；② 功能特性补全"整机防火墙一键开关"，并澄清安装时"放行防火墙"（仅新增一条放行 `8899` 端口的 UCI 规则）与面板内"关闭整个防火墙"是两回事；③ 仓库 owner 统一为规范名 `Qkvsvh`（GitHub 对 owner 大小写不敏感，原 `qkvsvh` 也可用），并同步修正 `g5pro-install.sh` 与 `build.sh` 中的同名引用。

## 一、项目定位与适配目标

- 后端 Go(gin) 单二进制，内嵌前端（`//go:embed webroot`），监听 Web `:8899` 与内嵌 SSH 服务端 `:3540`。
- 前端 Vue3 + Vite + Element Plus + TypeScript + xterm.js。
- 目标设备：中兴 G5 Pro 随身 WiFi / CPE，平台 MT6990、OpenWrt 21.02.7；WebSSH 以 root 运行，本地 `ubus` 调用透明。

## 二、真实数据源接入（G5 Pro 专用）

- **AMBR / QCI**：改用 UCI `zwrt_data_tmp.wwaniface1`（`qci` / `dl_ambr` / `ul_ambr`，单位 kbps），替代原作的占位/估算数据。
- **温度**：改用 sysfs `thermal_zone*`（soc_max / cpu / md / 主板 NTC），后端新增 `GET /api/sys/thermal`；不再依赖只返回单点温度的 `zwrt_bsp.thermal`。
- **SIM 信息**：改用 `zwrt_zte_mdm.api get_sim_info`（`sim_iccid` / `msisdn` / `imsi` / `Operator`），并严格区分 **ICCID（物理卡序列号）** 与 **SIM 卡号(手机号, msisdn)**，前端不再混用/误显。
- **OpenADB 开关**：
  - 开启：`ubus call zwrt_bsp.usb set '{"mode":"debug"}'`（按 debug 重配 USB gadget，挂接 ffs.adb）。
  - 关闭：`uci set usb.mode.mode=user; uci commit usb; /etc/init.d/usb.init restart`（**必须用 `user` 而非 `normal`**，否则 init 脚本 `case` 不含 `normal` 会导致 USB 异常）。
- **接口状态 / 信号 / 小区**：对接 `zte_nwinfo_api` / `zwrt_zte_mdm` / `zwrt_data` 等 ubus 接口，前端已适配 G5 Pro 接口名。
- **AMBR 接口**同时回传 `iccid` 与 `sim_number` 字段，前端 hero 区只展示 SIM 卡号，避免与 ICCID 混淆。

## 三、前端面板重构（Main.vue 为主）

### 3.1 信号强度区
- 信号强度 tile 增加 **SINR** 副指标。
- 信号强度增加**实时 dBm 数值**（与信号格图标同组展示），不再只有图标。

### 3.2 网络综合评分 tile
- 评分科学化：纳入 **RSRP / RSRQ / SINR / RSSI** 四维，按 3GPP 阈值映射为综合分。
- UI 扁平化：标题 + 分数，副指标 2×2 网格，QCI 业务等级副标题。
- 后续按需求简化：仅保留**标题 + 分数**（去掉 4 个子指标），QCI=9 的"默认数据"副标题隐藏。

### 3.3 流量统计
- 恢复原始 `repeat(auto-fit, minmax(150px, 1fr))` 等大布局（8 个 tile 本就等大），撤销曾误加的固定列网格。
- 流量 label / value 字号字重统一。

### 3.4 温度区
- CPU 温度 / 基带温度**两个圈等大**（`aspect-ratio: 1/1` 锁定）。

### 3.5 AMBR 顶部 2×2
- 第 4 格改为"网络综合评分"。
- `ambr-hero` 等高热力图调整：最终恢复最初的 `1fr 1fr` + 默认行高（避免强制等高留白）。

### 3.6 设备信息卡
- 内部 **CPU 负载 / 温度状态 / 内存使用** 三个子卡等高（`flex: 1 1 auto`）。
- 整卡满宽，视觉宽度等于顶部两卡之和。

### 3.7 网络信息卡
- `info-grid` 双列布局，新增 3 个分组小标题（**SIM/网络、频段/频道、设备标识**），用虚线分隔，信息有序易读。
- 信号强度行横向（图标 + dBm）。
- 窄屏 `.info-item` 改为横向（标签左、值右），并去掉 `.value` 的 `word-break: break-all`，避免"-83 dBm"等被挤压成竖排字符。

### 3.8 字号 / 字重体系统一
- 建立 5 档字号层级（卡片标题 / section 标题 / 主数字 / 次数字 / 标签值），统一字重与字间距，消除原有多处字号字重不一致、互相遮挡的问题。

## 四、导航与交互（Home.vue）

- 去掉顶部居中的"展开导航栏"按钮。
- 导航唤出改为**屏幕右边缘中间（屏幕高度 20%~80%，避开系统手势区）向左滑 ≥50px** 触发（早期曾用左缘，后改为右缘）。
- 桌面端保留右缘透明 hover 感应条作为兜底。
- 移除折叠态的浮动导航按钮。
- 新增**首次进入引导动画**：`localStorage` 标记 `webssh_nav_guide_v1`，未看过则弹出遮罩，含"右→左"滑动手势循环演示 + 说明文案 + "知道了"按钮；点击后即写入标记不再出现。

## 五、响应式与移动 / 平板适配

- 桌面（>1200px）：顶部 2 列、设备信息 3 列、底部 3 列等宽布局。
- **网络信息 / 流量统计 / 接口状态三卡**：用 `.bottom-cards` 包裹为 3 等列网格；**≤1200px 竖排单列**（覆盖手机与 iPad 横屏），桌面等宽。
- 设备信息三子卡、底部三卡均设 `align-items: stretch` 等高。
- 断点：≤1100px 设备与底部堆叠；≤900px 顶部堆叠；≤768px 流量 2 列、信息网格双列、`.info-item` 横向；≤1200px 三卡竖排单列。
- 修复关键 CSS 级联顺序 bug：`.bottom-cards` 的 1200px 单列 media 曾写在基础定义之前，被后定义的 `repeat(3,…)` 覆盖导致移动端一直并排；已移至基础定义之后使单列规则胜出。

## 六、构建 / 部署与安装

- 前端 `npm run build` → `cp -R webssh/dist/. gossh/webroot/` → Go `//go:embed` 内嵌，交叉编译为 arm64 静态二进制部署到目标设备。
- 提供设备端安装 / 启停脚本（含默认管理地址占位，可由使用者按实际环境修改），由 procd 托管自启与保活。

---

## 版本说明（v1.0.0 · 统一版本，2026-08-18）

本仓库此前为开发期多提交积累（累计 36 个本地提交）。为便于发布与回溯，已将所有改动**整理为单一统一版本**，历史提交已合并。以下按功能域汇总全部变更：

### 1. 项目基础
- fork 自 o8oo8o/WebSSH（原作针对中兴 U60 Pro / ZTE MU5120），全面适配**中兴 G5 Pro CPE**（MT6990 / OpenWrt 21.02.7）。
- 后端 Go(gin) 单二进制内嵌前端（`//go:embed webroot`），监听 Web `:8899` 与内嵌 SSH 服务端 `:3540`；前端 Vue3 + Vite + Element Plus + TypeScript + xterm.js。
- WebSSH 以 root 运行，本地 `ubus` 调用透明。

### 2. 真实数据源接入（G5 Pro 专用）
- **AMBR / QCI**：UCI `zwrt_data_tmp.wwaniface1`（`qci` / `dl_ambr` / `ul_ambr`，kbps）。
- **温度**：sysfs `thermal_zone*`（soc_max / cpu / md / 主板 NTC），后端 `GET /api/sys/thermal`。
- **SIM 信息**：`zwrt_zte_mdm.api get_sim_info`，严格区分 ICCID 与 msisdn（手机号）。
- **OpenADB 开关**：开 `ubus call zwrt_bsp.usb set '{"mode":"debug"}'`；关 `uci set usb.mode.mode=user` + init 重启（必须用 `user` 非 `normal`）。
- **信号 / 小区 / 接口状态**：对接 `zte_nwinfo_api` / `zwrt_zte_mdm` / `zwrt_data`。

### 3. 前端面板重构（Main.vue）
- 信号强度 tile 加 **SINR** 副指标 + 实时 **dBm** 数值。
- 网络综合评分：纳入 RSRP/RSRQ/SINR/RSSI 四维科学化，后简化为「标题 + 分数」。
- 流量统计恢复 `auto-fit` 等大 8 tile；温度 CPU/基带两圈等大；AMBR 顶部 2×2（第 4 格综合评分）。
- 设备信息卡三子卡等高满宽；网络信息卡 `info-grid` 双列 + 分组小标题 + 横向 info-item。
- 字号/字重体系统一为 5 档层级。

### 4. 导航与交互（Home.vue）
- 导航改为**屏幕右边缘中间左滑 ≥50px** 唤出（避开系统手势区）；桌面保留右缘感应条。
- 首次进入引导动画（`localStorage` 标记，看过即不再显示）。

### 5. 响应式与移动 / 平板适配
- 网络信息 / 流量统计 / 接口状态三卡 ≤1200px **竖排单列**（覆盖手机与 iPad 横屏）。
- 修复关键 CSS 级联顺序 bug（media 单列规则须写在 `.bottom-cards` 基础定义之后）。
- 各类卡片 `align-items: stretch` 等高。

### 6. 构建 / 部署 / 安装
- 前端 `npm run build` → `cp -R webssh/dist/. gossh/webroot/` → Go `//go:embed` 内嵌 → 交叉编译 arm64 静态二进制。
- procd init.d（`START=99`）托管自启与保活（去除对 rc.local 的依赖）。
- `g5pro-tools/hotswap-webssh.sh` 安全热更（subshell + nohup 避免 kill 自己 SSH 锁门）。

### 7. conntrack 治本（解决 PCDN 高并发丢包 → app 误判"广域网不在线"）
- `g5pro-tools/conntrack-tune`（原名 `zz-conntrack`，已更名语义化）：扩容 `nf_conntrack_max` + 缩短 UDP 超时，procd 开机自启。
- 后端 `GET/POST /api/sys/conntrack`：连接数 / 占用率 / 协议分布 / TOP 源设备名（仅 LAN 设备入榜，其余归 `non_lan_count`）。
- 前端「安装 / 卸载 conntrack 优化」按钮；旧名 `zz-conntrack` 残留自动兼容清理。

### 8. RF 频段功率控制（CPE modem RF 发射功率上限）
- `AT+EFMAXPWR` 写 NSA/SA 各频段最大功率（约 29dBm），末位 1/0 切换。
- 状态检测三态（default / applied_target / custom）+ 本地标记与硬件 readback 交叉校验（漂移告警）。
- 前端「所有频段功率设为 29dBm」「恢复默认」按钮 + 状态 tag + 漂移警示。

### 9. UI 布局优化与操作反馈
- 流量统计卡：上传/下载用量单行紧凑，8 tile 等高。
- 网络信息卡：4 tile 等高对齐；RF 区与连接状态区间距收紧、横向排版。
- 点击 conntrack / RF 按钮**弹窗显示具体做了什么**（步骤 / AT 命令逐条结果 / 元数据），含「复制原始输出」。

---

*本文档已脱敏：不含任何设备账号、密码、内网 IP 等敏感信息；本地工作记忆目录（.workbuddy）未纳入版本控制。*

# g5pro-tools —— 中兴 G5 Pro 设备侧运维脚本

本目录存放**直接跑在 G5 Pro CPE 上**的运维/加固脚本，与 WebSSH 本体（Go 二进制 + 前端）分离。
WebSSH 负责"网页 SSH 管理设备"，这里的脚本负责"设备本身的网络健壮性"与"WebSSH 自身生命周期"。

> 关键背景（一次说清）：G5 Pro 的 `/etc/rc.local` **默认开机不执行**，持久化开机动作只能靠
> **procd init.d 脚本**（`/etc/rc.common` 为 procd 模式，入口必须叫 `start_service()`，只写 `start()` 会被跳过）。
> 本目录所有自启脚本均走 procd，已全部在设备实测生效。

## 1. conntrack-tune —— conntrack 表扩容治本

G5 Pro conntrack 连接跟踪表扩容 + UDP 超时加速回收的**治本脚本**（procd init.d）。

- **问题**：默认 `nf_conntrack_max=16384`，跑 PCDN / 内网有设备狂刷 UDP 时表被撑爆 → 内核丢包 →
  设备到云端 30443 心跳丢失 → 中兴智慧生活 app 误判"广域网不在线"。
- **方案**：内核层面扩容 + 缩短 UDP 超时，**不杀进程、不限速、不影响正常使用**。
- **参数**（脚本顶部可调）：`CONNTRACK_MAX=524288`、`UDP_TIMEOUT=10`、
  `UDP_STREAM_TIMEOUT=30`、`TCP_EST_TIMEOUT=3600`。

部署（在 G5 Pro 上，192.168.0.1:3540，root）：

```sh
scp conntrack-tune root@192.168.0.1:/etc/init.d/conntrack-tune
chmod +x /etc/init.d/conntrack-tune
/etc/init.d/conntrack-tune enable   # 生成 /etc/rc.d/S99conntrack-tune，开机自启
/etc/init.d/conntrack-tune start    # 立即生效，无需重启
```

验证：`cat /proc/sys/net/netfilter/nf_conntrack_max`（应 524288）；
`cat /proc/sys/net/netfilter/nf_conntrack_count`（当前条目数，对比 max 看占用率）。

> 注意：**恢复出厂会清掉本脚本和 `/etc/rc.d/S99conntrack-tune`**，max 退回 16384、老病灶复发，需重新部署。
>
> **历史**：本脚本曾用名 `zz-conntrack`（`zz-` 只是 OpenWrt 启动顺序占位符，
> 含义"最后启动"，但用户看不出脚本干啥）。rename 到 `conntrack-tune` 后
> 语义清晰。旧名残留会在 WebSSH 装机时自动清理。

## 2. webssh-init —— WebSSH 自身 procd 自启

WebSSH 本体的 **procd init.d 自启脚本**（`-ConfigDir /data/webssh` 拉起 + `respawn` 崩溃自愈）。

**正常情况下你不需要手动部署它**——`webssh.sh` / `g5pro-install.sh` 的安装流程已内置
`add_procd_init()`，自动写 `/etc/init.d/webssh` 并 `enable`（生成 `/etc/rc.d/S90webssh`）。

仅当需要在已安装的设备上单独重置自启时，才手动跑：

```sh
scp webssh-init root@192.168.0.1:/etc/init.d/webssh
chmod 755 /etc/init.d/webssh
/etc/init.d/webssh enable && /etc/init.d/webssh start
```

## 3. hotswap-webssh.sh —— 离线安全热更新（应急）

专门解决"G5 Pro 上唯一 SSH 入口(3540) 就是 WebSSH 自己"的**锁门痛点**：直接 `service.sh restart`
会在 SSH 会话内部 kill 旧进程 → 连当前 SSH 一起掐断 → start 来不及跑 → 把自己锁在门外。

本脚本把"杀旧 → mv → 启新"塞进 `(... ) </dev/null >/tmp/swap.log 2>&1 &` 子 shell，父 SSH 退出时
子进程已被 init 收养（nohup 抗 SIGHUP，已在设备实测）。**不依赖 `setsid`**（BusyBox 不带）。

用法（在 G5 Pro 上执行）：

```sh
# 1. 先把新二进制 scp 到设备 /tmp/webssh_new（可用 /tmp/scp_put.exp 封装）
# 2. 上传本脚本并执行（SSH 会中途断开，约 5s 后 3540 由新进程接管）
sh /tmp/hotswap-webssh.sh
```

特性：旧进程先 TERM（3s 宽限）→ 还活着 KILL，平稳替换；旧二进制自动备份到 `/tmp/webssh_old_bak`；
启动参数固定 `-ConfigDir /data/webssh`。

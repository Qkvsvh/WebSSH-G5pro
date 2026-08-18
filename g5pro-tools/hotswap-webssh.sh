#!/bin/sh
# 安全热更新 WebSSH（不依赖 setsid，适用于 BusyBox only 的 G5 Pro）
#
# 用法（在设备本地）：
#   sh /tmp/hotswap_webssh.sh
#
# 行为：
#   1. 备份当前 /data/webssh/webssh → /tmp/webssh_old_bak
#   2. pgrep 取旧 PID，TERM → 等 3s → 还在则 KILL
#   3. mv /tmp/webssh_new → /data/webssh/webssh
#   4. 启新：优先 /etc/init.d/webssh start（procd 托管，respawn 自愈），否则 nohup 直启
#      （父 shell 退出不影响，已验证 nohup 抗 SSH 断开）
#
# 前提：
#   - 新二进制已 scp 到设备的 /tmp/webssh_new
#   - 旧二进制用的是 -ConfigDir /data/webssh（启动命令一致）
#
# 为什么不用 service.sh restart：
#   G5 Pro 上 3540 既是 WebSSH 的 SSH 端口，也是唯一的 SSH 入口；
#   service.sh restart 会在 SSH 会话内部 kill 旧进程 → 同时杀掉当前 SSH → start 子 shell 来不及拉新进程。
#   本脚本用 ( subshell ) + </dev/null + nohup + &，把启动动作脱离父 SSH session，
#   子进程立即被 init 收养，SSH 关掉不影响新进程接管 3540。

set -e

echo "=== HOTSWAP WebSSH ==="
OLDPID="$(pgrep -x webssh 2>/dev/null | head -1)"
if [ -z "$OLDPID" ]; then
  # pgrep -x 兜底（comm 是 webssh，但 argv[0] 在重 exec 后可能改了；再按 cmdline 头试一次）
  OLDPID="$(pgrep -af 'webssh -ConfigDir' | grep -v '$$' | head -1 | awk '{print $1}')"
fi
echo "OLD_PID=$OLDPID"

cd /data/webssh

# 子 shell 内完成"杀旧 → mv → 启新"完整动作，与父 SSH session 完全解耦
(
  # 备份
  cp /data/webssh/webssh /tmp/webssh_old_bak 2>/dev/null || true

  # 杀旧
  if [ -n "$OLDPID" ]; then kill -TERM "$OLDPID" 2>/dev/null || true; fi
  sleep 3
  if [ -n "$OLDPID" ] && kill -0 "$OLDPID" 2>/dev/null; then
    kill -KILL "$OLDPID" 2>/dev/null || true
    sleep 1
  fi

  # 替换
  mv -f /tmp/webssh_new /data/webssh/webssh
  chmod +x /data/webssh/webssh

  # 启新：若已注册 procd 自启脚本，交给它拉起（respawn 自愈，且与开机自启一致）；
  # 否则 nohup 直启（无 procd 的旧场景兜底）。
  if [ -x /etc/init.d/webssh ]; then
    /etc/init.d/webssh start
  else
    nohup ./webssh -ConfigDir /data/webssh >/tmp/webssh.log 2>&1 &
  fi
  echo "NEW_PID=$!"
) </dev/null >/tmp/swap.log 2>&1 &

SWAPPER=$!
disown $SWAPPER 2>/dev/null || true
echo "SWAPPER_PID=$SWAPPER"

# 给启动 5s 时间
sleep 5
NEW_PID="$(pgrep -x webssh 2>/dev/null | head -1)"
if [ -z "$NEW_PID" ]; then
  NEW_PID="$(pgrep -af 'webssh -ConfigDir' | grep -v '$$' | head -1 | awk '{print $1}')"
fi
echo "AFTER_PID=$NEW_PID"
if [ -n "$NEW_PID" ]; then
  echo "STATUS=OK"
  cat /tmp/swap.log
else
  echo "STATUS=FAIL"
  cat /tmp/swap.log
  exit 1
fi

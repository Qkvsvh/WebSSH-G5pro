#!/bin/sh
# ============================================================================
# WebSSH for G5pro —— 中兴 G5 Pro CPE 管理 / 自动更新脚本 (OpenWrt / arm64)
#
# 改造自原版 webssh.sh：保留原版「菜单实时检测更新 + 一键更新」的交互，
# 更新地址指向你自己的 GitHub 仓库；开机自启采用 procd init.d 方式
# （/etc/init.d/webssh + enable → /etc/rc.d/S90webssh，替代不可靠的 rc.local）。
#
# 在 G5 Pro 的 SSH 终端里运行（billy 即 root）：
#   sh webssh.sh                       进入互动菜单（自动显示是否有更新）
#   sh webssh.sh install               在线安装（GitHub Release 最新版）
#   sh webssh.sh update                在线更新（检测新版本并下载替换，会询问）
#   sh webssh.sh autoupdate            静默更新（供定时任务调用，不询问）
#   sh webssh.sh force                 强制安装（跳过版本比对，重新下载）
#   sh webssh.sh remove                卸载
#   sh webssh.sh start|stop|restart|status
#
# 自定义仓库：  REPO=你的名/仓库 sh webssh.sh
# ============================================================================

# ====================== 可配置项 ======================
# ↓↓↓ 你的 GitHub 仓库（原版是 cdwangtao/WebSSH-u60pro，已改为你的）↓↓↓
REPO="${REPO:-Qkvsvh/WebSSH-G5pro}"
Module_dir="/data/webssh"
Port="8899"
LanIP="192.168.0.1"
VERSION_FILE="$Module_dir/VERSION.txt"
BIN="$Module_dir/webssh"
# 发布物命名（沿用原版约定）：version.txt + webssh_<版本号> 二进制
VERSION_URL="https://github.com/$REPO/releases/latest/download/version.txt"
WEBSH_URL_PREFIX="https://github.com/$REPO/releases/latest/download/webssh_"
# =====================================================

# gh-proxy 加速（国内访问 GitHub 更稳，直连失败自动回退）
PROXIES="
https://ghproxy.net/
https://gh-proxy.org/
https://ghfast.top/
https://mirror.ghproxy.com/
"

fetch_url() {
    _original_url="$1"
    for _proxy in $PROXIES ""; do
        _url="${_proxy}${_original_url}"
        _result=$(curl -fsSL --connect-timeout 10 "$_url" 2>/dev/null)
        if [ $? -eq 0 ] && [ -n "$_result" ]; then
            echo "$_result"
            return 0
        fi
    done
    return 1
}

download_file() {
    _original_url="$1"
    _output="$2"
    _show_progress="${3:-0}"
    for _proxy in $PROXIES ""; do
        _url="${_proxy}${_original_url}"
        if [ "$_show_progress" = "1" ]; then
            curl -fSL --connect-timeout 10 -# "$_url" --output "$_output" && return 0
        else
            curl -fSL --connect-timeout 10 "$_url" --output "$_output" 2>/dev/null && return 0
        fi
    done
    return 1
}

# 下载二进制：先按原版版本号命名 webssh_<版本>，失败再试架构命名 webssh_<arch>
# （两种命名都在 G5 Pro 仓库里可用，自动兼容）
download_bin() {
    _version="$1"
    _out="$2"
    _arch=$(uname -m 2>/dev/null | sed 's/aarch64/arm64/; s/x86_64/amd64/')
    if download_file "${WEBSH_URL_PREFIX}${_version}" "$_out" 1; then
        return 0
    fi
    if download_file "https://github.com/$REPO/releases/latest/download/webssh_${_arch}" "$_out" 1; then
        return 0
    fi
    return 1
}

# ---- service.sh（薄封装，生命周期统一交给 procd /etc/init.d/webssh）----
# 不再独立 nohup 拉起，避免与 procd 托管冲突；存在 /etc/init.d/webssh 时直接委托它。
gen_service_sh() {
    cat > "$Module_dir/service.sh" << SEOF
#!/bin/sh
# 薄封装：生命周期统一交给 procd (/etc/init.d/webssh)
INIT="/etc/init.d/webssh"
BIN="$BIN"
Module_dir="$Module_dir"
start() {
    if [ -x "\$INIT" ]; then
        "\$INIT" start
    elif ! pgrep -f "\$BIN" >/dev/null 2>&1; then
        cd "\$Module_dir"
        nohup "\$BIN" -ConfigDir "\$Module_dir" >/dev/null 2>&1 &
        echo "WebSSH 已启动（无 procd，直启）"
    else
        echo "WebSSH 正在运行中，不重复启动"
    fi
}
stop() {
    if [ -x "\$INIT" ]; then
        "\$INIT" stop
    else
        for p in \$(pgrep -f "\$BIN"); do kill "\$p" 2>/dev/null; done
        echo "WebSSH 已关闭"
    fi
}
restart() {
    stop
    sleep 1
    start
}
case "\$1" in
    start) start ;;
    stop) stop ;;
    restart) restart ;;
    *) echo "用法: sh \$0 start|stop|restart" ;;
esac
SEOF
    chmod 755 "$Module_dir/service.sh"
}

# ---- 开机自启：procd init.d 方式（替代不可靠的 rc.local）----
# 写 /etc/init.d/webssh 并 enable，生成 /etc/rc.d/S90webssh 开机启动。
# 说明：G5 Pro 的 /etc/rc.local 默认开机不执行，procd init.d 脚本开机必跑
#       （同目录 g5pro-tools/zz-conntrack 已验证），用它托管 webssh 最稳。
add_procd_init() {
    INIT_SCRIPT="/etc/init.d/webssh"
    cat > "$INIT_SCRIPT" << SEOF
#!/bin/sh /etc/rc.common
USE_PROCD=1
START=90
STOP=10
start_service() {
    procd_open_instance webssh
    procd_set_param command $BIN -ConfigDir $Module_dir
    procd_set_param respawn
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}
SEOF
    chmod 755 "$INIT_SCRIPT"
    "$INIT_SCRIPT" enable 2>/dev/null
    echo "  已注册 procd 开机自启 (/etc/rc.d/S90webssh)"
}

add_alias() {
    ALIAS_CMD="alias webssh=\"sh $Module_dir/webssh.sh\""
    for f in /etc/profile /etc/shinit; do
        [ -f "$f" ] || touch "$f"
        grep -F "$ALIAS_CMD" "$f" >/dev/null 2>&1 || echo "$ALIAS_CMD" >> "$f"
    done
}

# ---- 运行环境守卫 ----
guard() {
    if [ "$(id -u 2>/dev/null)" != "0" ]; then
        echo "请用 root 运行（G5 Pro 的 billy 即为 root）"
        exit 1
    fi
    if [ ! -d /etc/init.d ]; then
        echo "未检测到 /etc/init.d，本脚本需在 OpenWrt(G5 Pro) 上运行。"
        exit 1
    fi
}

# ---- 安装（在线）----
do_install() {
    mkdir -p "$Module_dir"
    echo "检查版本信息..."
    if ! REMOTE_VERSION=$(fetch_url "$VERSION_URL" 2>/dev/null); then
        echo "  获取远程版本失败，请检查网络连接"
        return 1
    fi
    REMOTE_VERSION=$(echo "$REMOTE_VERSION" | tr -d '\r\n')
    echo "  最新版本: $REMOTE_VERSION"

    if [ -f "$BIN" ]; then
        echo "  检测到已安装，转更新流程"
        do_update
        return $?
    fi

    echo "下载 WebSSH 主程序..."
    if ! download_bin "$REMOTE_VERSION" "$BIN"; then
        echo "  下载失败，请检查网络或仓库发布物（需要 version.txt + webssh_<版本> 或 webssh_arm64）"
        return 1
    fi
    echo "  下载完成"
    chmod 755 "$BIN"
    echo "$REMOTE_VERSION" > "$VERSION_FILE"

    gen_service_sh
    add_procd_init
    sh "$Module_dir/service.sh" start
    open_firewall

    cp "$0" "$Module_dir/webssh.sh" 2>/dev/null
    add_alias

    echo ""
    echo "======================================"
    echo "       WebSSH 安装完成"
    echo "--------------------------------------"
    echo "  版本   : $REMOTE_VERSION"
    echo "  目录   : $Module_dir"
    echo "  访问   : http://$LanIP:$Port/app"
    echo "  快捷键 : webssh (重开终端生效)"
    echo "======================================"
}

# ---- 强制安装（跳过版本比对）----
do_force() {
    mkdir -p "$Module_dir"
    echo "获取版本信息..."
    REMOTE_VERSION=$(fetch_url "$VERSION_URL" 2>/dev/null | tr -d '\r\n')
    REMOTE_VERSION=${REMOTE_VERSION:-unknown}
    echo "  最新版本: $REMOTE_VERSION (强制安装，跳过版本比对)"
    echo "停止旧服务..."
    sh "$Module_dir/service.sh" stop 2>/dev/null
    echo "下载 WebSSH 主程序..."
    if ! download_bin "$REMOTE_VERSION" "$BIN"; then
        echo "  下载失败，请检查网络或仓库发布物"
        sh "$Module_dir/service.sh" start 2>/dev/null
        return 1
    fi
    chmod 755 "$BIN"
    echo "$REMOTE_VERSION" > "$VERSION_FILE"
    gen_service_sh
    add_procd_init
    sh "$Module_dir/service.sh" restart
    open_firewall
    cp "$0" "$Module_dir/webssh.sh" 2>/dev/null
    add_alias
    echo "强制安装完成，访问 http://$LanIP:$Port/app"
}

# ---- 更新（原版自动更新核心逻辑）----
# 不传参=交互（菜单调用，会询问）；传 auto=静默（定时任务调用）
do_update() {
    _auto="${1:-}"
    echo "检查更新..."
    if ! REMOTE_VERSION=$(fetch_url "$VERSION_URL" 2>/dev/null); then
        echo "  获取远程版本失败，请检查网络连接"
        return 1
    fi
    REMOTE_VERSION=$(echo "$REMOTE_VERSION" | tr -d '\r\n')
    LOCAL_VERSION=$(cat "$VERSION_FILE" 2>/dev/null | tr -d '\r\n')
    echo "  最新版本: $REMOTE_VERSION"
    echo "  当前版本: ${LOCAL_VERSION:-未安装}"

    if [ "$LOCAL_VERSION" = "$REMOTE_VERSION" ]; then
        echo "已是最新版本，无需更新"
        return 0
    fi

    if [ -z "$_auto" ]; then
        read -rp "发现新版本 $REMOTE_VERSION，是否更新？[y/N]: " u </dev/tty
        if [ "$(echo "$u" | tr '[:upper:]' '[:lower:]')" != "y" ]; then
            echo "已取消更新"
            return 0
        fi
    fi

    echo "停止 WebSSH 服务..."
    sh "$Module_dir/service.sh" stop 2>/dev/null
    echo "下载最新版本..."
    if ! download_bin "$REMOTE_VERSION" "$BIN.new"; then
        echo "  下载失败，请检查网络连接"
        echo "  尝试启动旧版本..."
        sh "$Module_dir/service.sh" start 2>/dev/null
        return 1
    fi
    chmod 755 "$BIN.new"
    mv -f "$BIN.new" "$BIN"
    echo "$REMOTE_VERSION" > "$VERSION_FILE"
    echo "重启 WebSSH 服务..."
    sh "$Module_dir/service.sh" restart
    echo ""
    echo "======================================"
    echo "       WebSSH 更新完成"
    echo "--------------------------------------"
    echo "  版本   : ${LOCAL_VERSION:-未知} -> $REMOTE_VERSION"
    echo "  访问   : http://$LanIP:$Port/app"
    echo "======================================"
}

# ---- 放行防火墙（OpenWrt uci，便于局域网访问 :8899）----
open_firewall() {
    if command -v uci >/dev/null 2>&1; then
        if ! uci show firewall 2>/dev/null | grep -q "webssh"; then
            uci add firewall rule >/dev/null 2>&1
            uci set firewall.@rule[-1].name='webssh'
            uci set firewall.@rule[-1].target='ACCEPT'
            uci set firewall.@rule[-1].src='lan'
            uci set firewall.@rule[-1].proto='tcp'
            uci set firewall.@rule[-1].dest_port="$Port"
            uci commit firewall
            /etc/init.d/firewall restart 2>/dev/null
        fi
    fi
}

# ---- 卸载 ----
remove() {
    sh "$Module_dir/service.sh" stop 2>/dev/null
    rm -f /etc/init.d/webssh /etc/rc.d/S*webssh
    FILE="/etc/rc.local"
    [ -f "$FILE" ] && sed -i "\|$Module_dir/service.sh|d" "$FILE"
    rm -rf "$Module_dir"
    ALIAS_CMD="alias webssh=\"sh $Module_dir/webssh.sh\""
    for f in /etc/profile /etc/shinit; do
        [ -f "$f" ] && sed -i "\|$ALIAS_CMD|d" "$f"
    done
    echo "卸载完成"
}

# ---- 运行状态查看 ----
status() {
    echo "===== WebSSH 状态 ====="
    if [ -f "$VERSION_FILE" ]; then
        echo "  版本     : $(cat "$VERSION_FILE" 2>/dev/null)"
    else
        echo "  版本     : 未安装"
    fi
    if [ -x "$BIN" ]; then
        echo "  二进制   : 存在 ($BIN)"
    else
        echo "  二进制   : 不存在"
    fi
    if ls /etc/rc.d/S*webssh >/dev/null 2>&1; then
        echo "  开机自启 : 已启用 (procd)"
    else
        echo "  开机自启 : 未启用"
    fi
    if pgrep -f "$BIN" >/dev/null 2>&1; then
        echo "  运行状态 : 运行中 (PID $(pgrep -f "$BIN" 2>/dev/null | head -1))"
    elif ps 2>/dev/null | grep -q "[w]ebssh"; then
        echo "  运行状态 : 运行中"
    else
        echo "  运行状态 : 未运行"
    fi
    echo "  配置目录 : $Module_dir"
    echo "  访问地址 : http://$LanIP:$Port/app"
}

# ---- 互动菜单（原版风格：自动检测更新并提示）----
pause() {
    echo
    printf "按回车键返回菜单..."
    read -r _dummy </dev/tty
    echo
}

show_menu() {
    while true; do
        clear
        _remote=$(fetch_url "$VERSION_URL" 2>/dev/null | tr -d '\r\n')
        _remote=${_remote:-未知}
        _local=$(cat "$VERSION_FILE" 2>/dev/null | tr -d '\r\n')
        if [ -z "$_local" ]; then
            _installed=0
        else
            _installed=1
        fi
        if [ "$_installed" = "1" ] && [ "$_local" != "$_remote" ] && [ "$_remote" != "未知" ]; then
            _has_update=1
        else
            _has_update=0
        fi

        _idx=1
        _install_idx=""; _update_idx=""; _force_idx=""; _remove_idx=""
        _start_idx=""; _stop_idx=""; _restart_idx=""; _status_idx=""
        echo "======================================"
        echo "       WebSSH for G5pro 管理脚本"
        echo "--------------------------------------"
        echo "  程序作者 : 基于原版 WebSSH-u60pro 二改 (适配 G5 Pro)"
        if [ "$_installed" = "1" ]; then
            if [ "$_has_update" = "1" ]; then
                echo "  当前版本: $_local (有更新可用)"
            else
                echo "  当前版本: $_local"
            fi
        fi
        if [ "$_has_update" = "1" ]; then
            echo "  最新版本: $_remote"
        fi
        echo "--------------------------------------"
        if [ "$_installed" = "0" ]; then
            _install_idx=$_idx; echo "  $_idx) 安装 (install)"; _idx=$((_idx + 1))
        elif [ "$_has_update" = "1" ]; then
            _update_idx=$_idx; echo "  $_idx) 更新 (update)"; _idx=$((_idx + 1))
        fi
        _force_idx=$_idx; echo "  $_idx) 强制安装 (force install)"; _idx=$((_idx + 1))
        _remove_idx=$_idx; echo "  $_idx) 卸载 (remove)"; _idx=$((_idx + 1))
        if [ "$_installed" = "1" ]; then
            _start_idx=$_idx; echo "  $_idx) 启动 (start)"; _idx=$((_idx + 1))
            _stop_idx=$_idx; echo "  $_idx) 停止 (stop)"; _idx=$((_idx + 1))
            _restart_idx=$_idx; echo "  $_idx) 重启 (restart)"; _idx=$((_idx + 1))
        fi
        _status_idx=$_idx; echo "  $_idx) 状态 (status)"; _idx=$((_idx + 1))
        echo "  0) 退出 (exit)"
        echo "======================================"
        echo
        printf "请输入选择: "
        read -r choice </dev/tty

        case "$choice" in
            $_install_idx)
                [ -n "$_install_idx" ] && { guard; do_install; pause; } ;;
            $_update_idx)
                [ -n "$_update_idx" ] && { guard; do_update; pause; } ;;
            $_force_idx)
                guard; do_force; pause ;;
            $_remove_idx)
                guard; remove; pause ;;
            $_start_idx)
                guard; sh "$Module_dir/service.sh" start; pause ;;
            $_stop_idx)
                guard; sh "$Module_dir/service.sh" stop; pause ;;
            $_restart_idx)
                guard; sh "$Module_dir/service.sh" restart; pause ;;
            $_status_idx)
                guard; status; pause ;;
            0)
                echo "已退出。"
                exit 0 ;;
            *)
                echo "无效的选择，请重新输入。"
                sleep 1 ;;
        esac
    done
}

case "$1" in
    install)            guard; do_install ;;
    update)             guard; do_update ;;
    autoupdate)         guard; do_update auto ;;
    force)              guard; do_force ;;
    remove|uninstall)   guard; remove ;;
    start)              guard; sh "$Module_dir/service.sh" start ;;
    stop)               guard; sh "$Module_dir/service.sh" stop ;;
    restart)            guard; sh "$Module_dir/service.sh" restart ;;
    status)             guard; status ;;
    menu|"")            show_menu ;;
    *) echo "用法: sh $0 [install|update|autoupdate|force|remove|start|stop|restart|status|menu]"; exit 1 ;;
esac

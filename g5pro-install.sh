#!/bin/sh
# ============================================================================
# WebSSH for G5pro —— 中兴 G5 Pro CPE 一键安装/管理脚本 (OpenWrt / arm64)
#
# 在 G5 Pro 的 SSH 终端里运行（billy 即 root）。
#
# 用法：
#   不传参数            => 进入【互动菜单】（安装/卸载/启动/重启/停止/状态）
#   install            用当前目录里已编译好的 webssh / webssh_arm64 本地安装
#   install-url        从 GitHub Release 下载安装（需联网）
#   update             重新下载最新版并重启
#   remove|uninstall   卸载（停服务 + 删目录 + 去防火墙规则）
#   start|stop|restart 服务控制
#   status             查看运行状态
#   menu               进入互动菜单
#
# 仓库无关：把下面 REPO 改成你自己的 GitHub 仓库即可。
#   也可在运行时覆盖：  REPO=你的名/仓库 sh g5pro-install.sh install-url
# ============================================================================

# ====================== 可配置项 ======================
# ↓↓↓ 改成你上传后的 GitHub 仓库（当前：Qkvsvh/WebSSH-G5pro）↓↓↓
REPO="${REPO:-Qkvsvh/WebSSH-G5pro}"
# 安装目录（OpenWrt 需可写，/data 已持久化）
Module_dir="/data/webssh"
BIN="$Module_dir/webssh"
Port="8899"
LanIP="192.168.0.1"          # G5 Pro 默认管理 IP，若不同请改
# =====================================================

# gh-proxy 加速（国内访问 GitHub 更稳，直连失败自动回退）
PROXIES="
https://ghproxy.net/
https://gh-proxy.org/
https://ghfast.top/
https://mirror.ghproxy.com/
"

# ---- 架构 → release 资产名 ----
detect_arch() {
    m=$(uname -m 2>/dev/null)
    case "$m" in
        aarch64|arm64) echo arm64 ;;
        x86_64|amd64)   echo amd64 ;;
        *) echo "$m" ;;
    esac
}
ARCH=$(detect_arch)
ASSET="webssh_${ARCH}"

# ---- 下载/抓取（curl 优先，wget 兜底；逐个代理尝试）----
_fetch() {
    _url="$1"
    for _p in $PROXIES ""; do
        _u="${_p}${_url}"
        if command -v curl >/dev/null 2>&1; then
            _r=$(curl -fsSL --connect-timeout 12 "$_u" 2>/dev/null) && [ -n "$_r" ] && { echo "$_r"; return 0; }
        fi
        if command -v wget >/dev/null 2>&1; then
            _r=$(wget -q -O - --timeout=12 "$_u" 2>/dev/null) && [ -n "$_r" ] && { echo "$_r"; return 0; }
        fi
    done
    return 1
}

_dl() {
    _url="$1"; _out="$2"
    for _p in $PROXIES ""; do
        _u="${_p}${_url}"
        if command -v curl >/dev/null 2>&1; then
            curl -fSL --connect-timeout 12 "$_u" --output "$_out" 2>/dev/null && return 0
        fi
        if command -v wget >/dev/null 2>&1; then
            wget -q -O "$_out" --timeout=12 "$_u" 2>/dev/null && return 0
        fi
    done
    return 1
}

# ---- service.sh（薄封装，生命周期统一交给 procd /etc/init.d/webssh）----
# 不再独立 nohup 拉起，避免与 procd 托管冲突；存在 /etc/init.d/webssh 时直接委托它。
gen_service_sh() {
    cat > "$Module_dir/service.sh" << SEOF
#!/bin/sh
# 薄封装：生命周期统一交给 procd (/etc/init.d/webssh)
INIT="/etc/init.d/webssh"
BIN="$Module_dir/webssh"
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

# ---- 放行防火墙（OpenWrt uci）----
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

# ---- 取最新版本号（仅用于显示 / 写版本文件）----
latest_tag() {
    _t=$(_fetch "https://api.github.com/repos/$REPO/releases/latest" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    echo "${_t:-latest}"
}

install_url() {
    TAG=$(latest_tag)
    echo "最新版本: $TAG"
    mkdir -p "$Module_dir"
    # 兼容两种发布物命名：webssh_<版本号> 或 webssh_<架构>
    if ! _dl "https://github.com/$REPO/releases/latest/download/webssh_${TAG}" "$Module_dir/webssh.new"; then
        if ! _dl "https://github.com/$REPO/releases/latest/download/$ASSET" "$Module_dir/webssh.new"; then
            echo "下载失败：请确认 $REPO 的 Release 中存在 webssh_${TAG} 或 $ASSET，或检查网络"
            return 1
        fi
    fi
    echo "下载完成"
    chmod 755 "$Module_dir/webssh.new"
    mv -f "$Module_dir/webssh.new" "$Module_dir/webssh"
    echo "$TAG" > "$Module_dir/VERSION.txt"
    gen_service_sh
    add_procd_init
    open_firewall
    sh "$Module_dir/service.sh" restart
    echo "安装完成，访问 http://$LanIP:$Port/app"
}

install_local() {
    _bin=""
    for _c in "./webssh" "./$ASSET" "./webssh_arm64" "./webssh_amd64"; do
        [ -f "$_c" ] && _bin="$_c" && break
    done
    if [ -z "$_bin" ]; then
        echo "当前目录未找到二进制（webssh / $ASSET），请先放好再执行"
        return 1
    fi
    mkdir -p "$Module_dir"
    cp "$_bin" "$Module_dir/webssh"
    chmod 755 "$Module_dir/webssh"
    echo "local" > "$Module_dir/VERSION.txt"
    gen_service_sh
    add_procd_init
    open_firewall
    sh "$Module_dir/service.sh" restart
    echo "安装完成，访问 http://$LanIP:$Port/app"
}

remove() {
    sh "$Module_dir/service.sh" stop 2>/dev/null
    rm -f /etc/init.d/webssh /etc/rc.d/S*webssh
    [ -f /etc/rc.local ] && sed -i "\|$Module_dir/service.sh|d" /etc/rc.local
    if command -v uci >/dev/null 2>&1; then
        _line=$(uci show firewall 2>/dev/null | grep -n "webssh" | head -1 | cut -d: -f1)
        if [ -n "$_line" ]; then
            uci delete firewall.@rule[$_line] 2>/dev/null
            uci commit firewall 2>/dev/null
        fi
    fi
    rm -rf "$Module_dir"
    echo "已卸载"
}

# ---- 运行状态查看 ----
status() {
    echo "===== WebSSH 状态 ====="
    if [ -f "$Module_dir/VERSION.txt" ]; then
        echo "  版本     : $(cat "$Module_dir/VERSION.txt" 2>/dev/null)"
    else
        echo "  版本     : 未安装"
    fi
    if [ -x "$Module_dir/webssh" ]; then
        echo "  二进制   : 存在 ($Module_dir/webssh)"
    else
        echo "  二进制   : 不存在"
    fi
    if ls /etc/rc.d/S*webssh >/dev/null 2>&1; then
        echo "  开机自启 : 已启用 (procd)"
    else
        echo "  开机自启 : 未启用"
    fi
    if pgrep -f "$Module_dir/webssh" >/dev/null 2>&1; then
        echo "  运行状态 : 运行中 (PID $(pgrep -f "$Module_dir/webssh" 2>/dev/null | head -1))"
    elif ps 2>/dev/null | grep -q "[w]ebssh"; then
        echo "  运行状态 : 运行中"
    else
        echo "  运行状态 : 未运行"
    fi
    echo "  配置目录 : $Module_dir"
    echo "  访问地址 : http://$LanIP:$Port/app"
}

# ---- 运行环境守卫 ----
guard() {
    if [ "$(id -u 2>/dev/null)" != "0" ]; then
        echo "请用 root 运行（G5 Pro 的 billy 即为 root）"
        exit 1
    fi
    if [ ! -d /etc/init.d ]; then
        echo "未检测到 /etc/init.d，本脚本需在 OpenWrt (G5 Pro) 上运行。"
        echo "本地交叉编译请用仓库里的 build.sh，再 scp 到设备执行本脚本的 install。"
        exit 1
    fi
    if [ "$ARCH" != "arm64" ] && [ "$ARCH" != "amd64" ]; then
        echo "警告：当前架构 $ARCH 非预期（G5 Pro 为 arm64），仍继续。"
    fi
}

# ---- 互动菜单 ----
pause() {
    echo
    printf "按回车键返回菜单..."
    read -r _dummy
    echo
}

show_menu() {
    while true; do
        echo "=========================================="
        echo "        WebSSH for G5pro 管理脚本"
        echo "------------------------------------------"
        if [ -f "$Module_dir/VERSION.txt" ]; then
            echo "  当前版本 : $(cat "$Module_dir/VERSION.txt" 2>/dev/null)"
        else
            echo "  当前状态 : 未安装"
        fi
        echo "------------------------------------------"
        echo "  1) 安装        (本地，使用当前目录的二进制)"
        echo "  2) 安装/更新   (在线，从 GitHub Release 拉取)"
        echo "  3) 卸载"
        echo "  4) 启动"
        echo "  5) 重启"
        echo "  6) 停止"
        echo "  7) 状态"
        echo "  0) 退出"
        echo "=========================================="
        echo
        printf "请输入选择 [0-7]: "
        read -r choice
        case "$choice" in
            1) guard; install_local; pause ;;
            2) guard; install_url; pause ;;
            3) guard; remove; pause ;;
            4) guard; sh "$Module_dir/service.sh" start; pause ;;
            5) guard; sh "$Module_dir/service.sh" restart; pause ;;
            6) guard; sh "$Module_dir/service.sh" stop; pause ;;
            7) guard; status; pause ;;
            0) echo "已退出。"; exit 0 ;;
            *) echo "无效选择，请输入 0-7。"; sleep 1 ;;
        esac
    done
}

case "$1" in
    install)            guard; install_local ;;
    install-url|update) guard; install_url ;;
    remove|uninstall)   guard; remove ;;
    start)              guard; sh "$Module_dir/service.sh" start ;;
    stop)               guard; sh "$Module_dir/service.sh" stop ;;
    restart)            guard; sh "$Module_dir/service.sh" restart ;;
    status)             guard; status ;;
    menu|"")            show_menu ;;
    *) echo "用法: sh $0 [install|install-url|update|remove|uninstall|start|stop|restart|status|menu]"; exit 1 ;;
esac

#!/bin/sh
# ============================================================================
# WebSSH for G5pro —— 交叉编译脚本（在 macOS / Linux x86_64 开发机上运行）
#
# 产出: gossh/webssh_<arch>   （默认 arm64，适配中兴 G5 Pro）
# 发布: 把 gossh/webssh_arm64 作为 GitHub Release 资产上传，
#       资产名必须保持 webssh_arm64（安装脚本 g5pro-install.sh 按此名下载）
#
# 用法:
#   sh build.sh                # 编译 arm64
#   GOARCH=amd64 sh build.sh   # 编译 amd64
# ============================================================================
set -e
cd "$(dirname "$0")"

ARCH="${GOARCH:-arm64}"
OS="${GOOS:-linux}"
BIN="webssh_${ARCH}"

# ---- 1. 前端构建 ----
echo "==> 构建前端 (npm run buildOnly)"
cd webssh
[ -d node_modules ] || npm install
npm run buildOnly
cd ..

# ---- 2. 拷贝前端到 gossh/webroot（Go //go:embed 会打包进去）----
echo "==> 拷贝前端到 gossh/webroot"
rm -rf gossh/webroot/assets gossh/webroot/index.html
cp -r webssh/dist/. gossh/webroot/

# ---- 3. Go 交叉编译 ----
echo "==> 交叉编译 Go ($OS/$ARCH)"
export CGO_ENABLED=0
export GOOS="$OS"
export GOARCH="$ARCH"
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
export GOFLAGS="${GOFLAGS:--mod=mod}"
export GOPATH="${GOPATH:-$HOME/go}"
GO="${GO:-go}"

cd gossh
"$GO" build -trimpath -ldflags="-s -w" -o "$BIN" .
cd ..

echo "==> 完成: gossh/$BIN"
echo "    发布步骤：在 GitHub 创建 Release，上传 gossh/$BIN（保持文件名 $BIN）"
echo "    用户在 G5 Pro 上执行："
echo "      sh -c \"\$(curl -fsSL https://raw.githubusercontent.com/Qkvsvh/WebSSH-G5pro/main/g5pro-install.sh)\""

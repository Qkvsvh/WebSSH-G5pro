# WebSSH for G5pro

网页版 SSH 终端 + SFTP 文件管理 + 系统管理面板，专为 **中兴 G5 Pro CPE（MT6990 / OpenWrt 21.02.7，arm64）** 适配。

后端 Go(gin) 单二进制内嵌 Vue3 前端，监听 `8899`；同时内嵌一个 SSH 服务端（默认 `3540`）。  
单文件部署，procd init.d 开机自启（崩溃自愈 + 开机自启动），一条命令安装。

> 本仓库由维护者持续更新。版本变更见 [CHANGELOG.md](CHANGELOG.md)；  
> 推送 `main` 分支时 GitHub Actions 会自动交叉编译并发布到 Releases（资产名固定 `webssh_arm64`）。

---

## 基于 / 致谢 (Credits)

本项目修改自以下开源项目，遵循其 MIT 许可证并保留原作者版权声明（见 [LICENSE](LICENSE)）：

- **WebSSH** by [o8oo8o](https://github.com/o8oo8o/WebSSH) —— 原始项目（MIT, © 2021 o8oo8o）
- **WebSSH-u60pro** by [cdwangtao](https://github.com/cdwangtao/WebSSH-u60pro) —— 针对中兴 U60 Pro / G5 Pro 的适配分支

本次适配目标设备为 **中兴 G5 Pro CPE**（MT6990 / OpenWrt 21.02.7）。

## 功能特性

- 网页 SSH 终端（xterm.js）/ SFTP 文件管理
- 实时设备仪表盘：5G/4G 信号、流量统计、AMBR/QCI、温度（CPU / 基带 / 主板）、SIM 卡号
- OpenADB 一键开启 / 关闭
- 主机 / 连接 / 用户 / 命令收藏 / 审计 等管理功能
- 单文件二进制 + procd init.d 开机自启（崩溃自愈），一条命令安装 / 升级 / 卸载

---

## 使用教程（给最终用户）

适用于已经开启 G5 Pro SSH 的用户（设备默认管理地址 `192.168.0.1`，SSH 端口 `3540`）。

### 1. 一键安装

在 G5 Pro 的 SSH 终端里执行下面一行：

```sh
sh -c "$(curl -fsSL https://raw.githubusercontent.com/qkvsvh/WebSSH-G5pro/main/g5pro-install.sh)"
```

脚本会自动完成：

1. 下载最新 Release 的 `webssh_arm64`
2. 安装到 `/data/webssh`
3. 注册 procd 开机自启（`/etc/init.d/webssh` + `enable` → `/etc/rc.d/S90webssh`）
4. 放行防火墙
5. 启动服务

### 2. 打开管理面板

浏览器访问 **<http://192.168.0.1:8899/app>** ，首次打开需初始化一个网页管理员账号。

### 3. 升级到新版本

重新执行上面的安装命令即可，或显式升级：

```sh
sh g5pro-install.sh update
```

### 4. 启停 / 卸载

```sh
sh g5pro-install.sh start|stop|restart
sh g5pro-install.sh remove      # 卸载（保留配置目录 /data/webssh/.GoWebSSH）
```

---

## 开发者：本地编译与发布

### 环境要求

- Go 1.21+（交叉编译目标 `linux/arm64`）
- Node 20+（构建前端）

### 一键交叉编译

```sh
sh build.sh                 # 产出 gossh/webssh_arm64（默认 arm64）
GOARCH=amd64 sh build.sh    # 也可编 amd64
```

编译流程：构建前端 → 拷贝到 `gossh/webroot`（Go `//go:embed` 会打包进去）→  
`CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w"`。

### 手动步骤（等价于 build.sh）

```sh
# 前端
cd webssh && npm install && npm run buildOnly && cd ..

# 拷贝前端到后端 webroot
rm -rf gossh/webroot/assets gossh/webroot/index.html
cp -r webssh/dist/. gossh/webroot/

# 后端交叉编译
cd gossh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o webssh_arm64 .
cd ..
```

> 可选 UPX 压缩：`upx --best --lzma gossh/webssh_arm64`（CI 中已自动压缩）。

### 发布新版本

- **推荐**：直接推送 `main`，GitHub Actions 会自动编译并在 Releases 创建带时间戳的版本，  
  同时上传稳定名资产 `webssh_arm64`（安装脚本按此名下载 `releases/latest/download/webssh_arm64`）。
- **手动**：在 GitHub 创建 Release，把 `gossh/webssh_arm64` 作为资产上传，  
  **文件名必须保持 `webssh_arm64`**，否则安装脚本下载会 404。

### 发布前注意

安装脚本 `g5pro-install.sh` 顶部的 `REPO` 变量已设为本仓库 `qkvsvh/WebSSH-G5pro`。
如果你 fork 到自己的账号，把它改成你的 `用户名/仓库名` 即可：

```sh
REPO="${REPO:-qkvsvh/WebSSH-G5pro}"
```

---

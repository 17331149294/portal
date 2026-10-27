Portal
======

HuaWei Portal Middleware（华为 Portal 认证中间件），基于 Syler 改造。

一个支持华为 Portal 协议的网络管理中间件，支持：

* 华为 Portal 1.0 协议
* 华为 Portal 2.0 协议

本版本在 Syler 基础上新增 **MAC 缓存免认证**功能（mac.txt 持久化、可配置有效期），
并且已改造为 **零外部依赖、可离线编译**（内置最小 TOML 解析器，radius 包已内嵌到 `internal/radius`）。

## 已测试设备

### 华为
S3700, AR 1220

### 华三
WX5000

## 特性

* 华为 Portal 1.0 / 2.0 协议
* HTTP 登录/注销页面（`/login`、`/logout`）
* RADIUS 认证 / 计费服务
* MAC 免认证（`mac.txt` 缓存，默认有效期 `basic.mac_cache_timeout`，单位秒）

## 编译

本模块无任何外部 Go 依赖，**不需要联网拉取依赖**即可编译：

```bash
# 本机编译（当前系统）
go build -o portal .

# 交叉编译示例（Linux amd64）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o portal_linux_amd64 .
```

### 云端自动编译 + Release

仓库已包含 `.github/workflows/build.yml`，推送 `main` 分支会触发 8 个目标平台的交叉编译
（windows/linux × 386/amd64/arm/arm64），并上传构建产物（Actions Artifact）。

**打 tag 发布 Release：**

```bash
git tag v1.0.0
git push origin v1.0.0
```

推送 `v*` 标签后，workflow 会自动把所有平台的二进制 + SHA256 校验和上传到 GitHub Release。

## 配置

配置文件默认 `./portal.conf`（可用 `-config` 参数指定），TOML 格式，例如：

```toml
[radius]
    port = 1812
    acc_port = 1813
    secret = "huawei123"
    enabled = true

[http]
    port = 8081

[huawei]
    port = 50100
    version = 2
    secret = "huawei123"
    nas_port = 2000
    domain = "dongjianwei.cn"
    timeout = 15

[basic]
    callback_logout = "http://10.10.100.68"
    logfile = "./debug.log"
    remote_ip_as_user_ip = true
    nas_ip = "10.10.100.253"
    mac_cache_timeout = 3600   # MAC 缓存有效期(秒)
```

## 注意

本开源包仅供非商业使用。

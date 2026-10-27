Portal
======

HuaWei Portal Middleware（华为 Portal 认证中间件），基于 Syler 改造。

一个支持华为 Portal 协议的网络管理中间件，支持：

* 华为 Portal 1.0 协议
* 华为 Portal 2.0 协议

本版本在 Syler 基础上新增 **MAC 缓存免认证**功能（`mac.txt` 持久化、可配置有效期），
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
* 踢用户下线（HTTP `/logout` 向 NAS 发送 REQ_LOGOUT）
* 零外部依赖，支持 8 平台交叉编译，可一键云端发版

---

## 使用办法（运行与部署）

### 1. 准备配置文件

配置文件默认 `./portal.conf`（可用 `-config` 参数指定），TOML 格式。完整配置项见下方
[配置项参考](#配置项参考)。一个最小可用的示例：

```toml
[radius]
    port = 1812        # RADIUS 认证端口
    acc_port = 1813    # RADIUS 计费端口
    secret = "huawei123"
    enabled = true     # 是否启用 RADIUS 认证/计费服务

[http]
    port = 8081        # 门户回调 HTTP 服务端口
    # white_list = "10.0.0.0/8,192.168.1.100"   # 可选：请求来源 IP 白名单，空=放行全部

[huawei]
    port = 50100       # 与 NAS 通信的 UDP 端口
    version = 2        # Portal 协议版本：1 或 2
    secret = "huawei123"   # 与 NAS 上配置的共享密钥一致
    nas_port = 2000
    domain = "example.com" # 认证用户名拼的域名后缀
    timeout = 15       # 请求 NAS 的超时秒数

[basic]
    callback_logout = "http://10.10.100.68"
    logfile = "./debug.log"   # 留空则输出到标准输出
    remote_ip_as_user_ip = true   # 用发起登录的客户端 IP 作为用户 IP
    nas_ip = "10.10.100.253"      # 指定 NAS IP（可覆盖 /login 的 nasip 参数）
    mac_cache_timeout = 3600      # MAC 缓存有效期（秒）
```

### 2. 启动服务

```bash
./portal                     # 默认读取 ./portal.conf
./portal -config /path/portal.conf   # 指定配置文件
./portal -h                  # 查看命令行帮助
```

启动后自动拉起以下服务（与配置有关）：

| 服务 | 端口 | 说明 |
| --- | --- | --- |
| HTTP（门户回调） | `http.port` | `/login`、`/logout`、`/` |
| 华为 Portal（UDP） | `huawei.port` | 与 NAS 通信，接收/响应挑战与认证请求 |
| RADIUS 认证 | `radius.port` | 仅在 `radius.enabled=true` 时启动 |
| RADIUS 计费 | `radius.acc_port` | 仅在 `radius.enabled=true` 时启动 |

### 3. HTTP 认证接口（对接门户）

门户页面认证成功后回调用 `/login`（GET，参数放在 query 或 form）：

```
GET /login?userip=<用户IP>&username=<用户名>&userpwd=<密码>&timeout=<时长秒>&nasip=<NAS IP>
```

* 用户名会拼上 `@<huawei.domain>` 作为完整账号提交认证。
* 认证成功后返回该终端的 **MAC 地址**，并自动把 MAC 写入 `mac.txt` 缓存。
* 认证失败返回 HTTP 400 和错误信息。
* 若未传 `username` 且 `basic.random_user=true`，会生成随机账号完成认证。
* 用户 IP 取值优先级：`basic.remote_ip_as_user_ip=true` 时用发起请求的客户端 IP →
  否则用 `userip` 参数 → 否则解析 `Referer`。

踢用户下线（向 NAS 发送 REQ_LOGOUT）：

```
GET /logout?userip=<用户IP>&nasip=<NAS IP>
```

根路径 `/`：提供健康/信息页。

### 4. MAC 免认证

* 终端首次认证成功后，其 MAC 自动写入 `mac.txt`，默认有效期 `basic.mac_cache_timeout`（秒）。
* 同 MAC 在有效期内再次认证时直接放行，无需再次校验密码。
* `mac.txt` 每行格式（4 个字段，逗号分隔）：

  ```
  MAC地址,过期时间(RFC3339),用户名,IP
  ```

* 程序启动时自动加载 `mac.txt`；删除该文件可清空全部 MAC 缓存。
* 缓存写入/加载目录为程序当前工作目录。

### 5. 常见部署拓扑

```
门户/交换机 → HTTP /login /logout → 本中间件 →(华为 Portal UDP)→ NAS
                                        ↓ 认证时
                                   RADIUS 认证/计费（可选）
```

---

## 编译方法

本模块无任何外部 Go 依赖，**不需要联网拉取依赖**即可编译（`go.mod` 为零依赖，
radius 与 TOML 解析均为内置实现）。

### 本机编译（当前系统）

```bash
go build -o portal .
```

### 交叉编译 8 平台

统一加 `CGO_ENABLED=0` 可离线交叉编译（Go 需为支持对应 GOARCH 的版本）：

```bash
# Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o portal_linux_amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=386   go build -trimpath -ldflags "-s -w" -o portal_linux_386 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm   go build -trimpath -ldflags "-s -w" -o portal_linux_arm .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o portal_linux_arm64 .

# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o portal_windows_amd64.exe .
CGO_ENABLED=0 GOOS=windows GOARCH=386   go build -trimpath -ldflags "-s -w" -o portal_windows_386.exe .
CGO_ENABLED=0 GOOS=windows GOARCH=arm   go build -trimpath -ldflags "-s -w" -o portal_windows_arm.exe .
CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o portal_windows_arm64.exe .
```

### 云端自动编译 + Release

仓库已包含 `.github/workflows/build.yml`：

* 推送 `main` 分支 → 自动交叉编译 **windows/linux × 386/amd64/arm/arm64** 共 8 个
  二进制，并上传为 Actions Artifact。
* 推送 **`v*` 标签** → 除上述构建外，还会把所有平台的二进制 + `SHA256SUMS.txt`
  自动发布到 GitHub Release。

打 tag 发版：

```bash
git tag v1.0.1
git push origin v1.0.1
```

发版后可到 Release 页面下载，或命令行下载：

```bash
# 需要安装 gh 并登录
gh release download v1.0.1 --repo <owner>/<repo> --pattern 'portal_linux_arm64'
```

### 验证与冒烟

```bash
go vet ./...        # 静态检查
go test ./...       # 运行全部单元测试
./portal -h         # 确认可执行文件可用
```

---

## 配置项参考

| 配置项 | 默认值 | 说明 |
| --- | --- | --- |
| `radius.enabled` | `false` | 是否启用 RADIUS 认证/计费服务 |
| `radius.port` | `1812` | RADIUS 认证端口 |
| `radius.acc_port` | `1813` | RADIUS 计费端口 |
| `radius.secret` | `huawei123` | RADIUS 共享密钥 |
| `radius.auth_type` | `random` | RADIUS 认证方式 |
| `http.port` | `8080` | HTTP 服务端口 |
| `http.white_list` | 空 | 允许请求的来源 IP 白名单（逗号分隔）；空=放行全部 |
| `huawei.port` | `50100` | 与 NAS 通信的 UDP 端口 |
| `huawei.version` | `1` | Portal 协议版本（1 或 2） |
| `huawei.secret` | `testing123` | 与 NAS 配置一致的共享密钥 |
| `huawei.nas_port` | `2000` | NAS 端口 |
| `huawei.domain` | `huawei.com` | 认证用户名拼的域名后缀 |
| `huawei.timeout` | `15` | 请求 NAS 超时秒数 |
| `basic.random_user` | `true` | 无用户名时是否生成随机账号认证 |
| `basic.logfile` | 空 | 日志文件路径；空=输出到标准输出 |
| `basic.callback_logout` | 空 | 注销回调地址 |
| `basic.remote_ip_as_user_ip` | `false` | 是否用发起请求的客户端 IP 作为用户 IP |
| `basic.login_page` | `./login.html` | 登录页模板路径 |
| `basic.nas_ip` | 空 | 指定 NAS IP（可覆盖 `/login` 的 `nasip` 参数） |
| `basic.default_timeout` | `0` | 默认认证时长（秒），`0` 表示不设 |
| `basic.mac_cache_timeout` | `3600` | MAC 免认证缓存有效期（秒） |

## 注意

本开源包仅供非商业使用。

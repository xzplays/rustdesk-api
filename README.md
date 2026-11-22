# RustDesk API Server（xzplays 维护）

本仓库是 [lejianwen/rustdesk-api](https://github.com/lejianwen/rustdesk-api) 的分支，**上游已停止维护**；自 fork 起由 xzplays 独立维护，`master` 为唯一主干。

> 感谢原作者 [lejianwen](https://github.com/lejianwen) 及所有贡献者：本项目的数据模型、Gin 路由骨架、Web 客户端资源与大量业务逻辑都来自上游。[lejianwen/rustdesk-api-web](https://github.com/lejianwen/rustdesk-api-web) 的后台界面同样出自其手（本仓库配套的前端已重写，见下文）。

## 这套服务做什么

为自建 RustDesk 服务端提供账号中心、地址簿、审计与后台管理：

- 账号：登录、注册（可开关）、改密、启用/停用、用户组
- 设备：设备列表、别名、分组（设备由客户端上报后自动绑定到登录账号）
- 地址簿：个人地址簿、收藏夹、共享规则、标签
- 审计：登录审计（含 hbbs 令牌校验结果）、连接审计、文件传输审计、登录日志、令牌
- 其他：OAuth（OIDC）单点登录、分享记录、系统配置读取
- 后台 UI：`/_admin/`（构建产物嵌入 `resources/admin`）

## 与上游的差异

| 能力 | 说明 |
| --- | --- |
| hbbs 令牌校验 | `POST /api/server/verify-token`，配套 `audit_login` 表；hbbs 侧用 `X-Rdtest-Secret` 调用，用于"未登录不允许建立连接" |
| 自助接口 | `GET /api/my/peers`、`POST /api/my/peer/update`、`POST /api/my/peer/unbind`、`GET /api/my/address_book`，让普通用户管理自己的设备与地址簿 |
| 开放注册 | `app.register` 开关 + `/api/admin/login-options` 暴露注册状态，前端据此显示注册入口 |
| 后台登录审计页 | `/api/admin/audit_login/list`（管理员界面见 api-web 仓库） |
| 若干修复 | `user/update` 的 `status=0` 校验、GORM 零值不落库、`verify-secret` 兼容旧配置文件（环境变量兜底）、数据库版本 266 |

协议与接口保持与官方客户端兼容：前端 `access_token` 走 `Authorization: Bearer <token>`，管理端走 `api-token` 头。

## 构建

需要 Go 1.22+，SQLite 驱动依赖 CGO：

```bash
# 前端产物（可选，仓库内已带一份）
# 见 https://github.com/xzplays/rustdesk-api-web
CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o apimain ./cmd/apimain.go
```

静态二进制（musl）：

```bash
CGO_ENABLED=1 CC=musl-gcc CGO_LDFLAGS=-static go build -trimpath -ldflags "-s -w" -o apimain ./cmd/apimain.go
```

运行目录需包含 `data/`、`resources/`（`resources/admin` 为后台 UI，`resources/templates` 为模板）。首次启动会自动迁移数据库并创建 `admin` 账号，初始密码打印在日志里，可用 `./apimain reset-admin-pwd <新密码>` 重置。

## 配置

配置读取优先级：`conf/config.yaml` → 环境变量 `RUSTDESK_API_*`（`.`/`-` 转 `_`），例如：

```bash
RUSTDESK_API_GIN_API_ADDR=127.0.0.1:21114
RUSTDESK_API_RUSTDESK_ID_SERVER=example.com:21116
RUSTDESK_API_RUSTDESK_RELAY_SERVER=example.com:21117
RUSTDESK_API_RUSTDESK_API_SERVER=https://example.com
RUSTDESK_API_RUSTDESK_KEY_FILE=/opt/rustdesk/id_ed25519.pub
RUSTDESK_API_JWT_KEY=<随机串>
RUSTDESK_API_RUSTDESK_VERIFY_SECRET=<与 hbbs 共享的随机串>
RUSTDESK_API_APP_REGISTER=true
```

`rustdesk.verify-secret` 为空时 `/api/server/verify-token` 直接返回错误（不会放行），请务必与 hbbs 的 `API_VERIFY_SECRET` 保持一致。

## 部署提示

- 建议只监听 `127.0.0.1`，由 nginx 反代并终止 TLS。
- SQLite 路径在代码中固定为 `./data/rustdeskapi.db`（相对工作目录），systemd 里显式设置 `WorkingDirectory`。
- 生产环境请把 `runtime/log.txt`（或 `RUSTDESK_API_LOGGER_PATH`）指向可写目录。

## 许可证

MIT，见 [LICENSE](LICENSE)。原始版权归 lejianwen 及贡献者所有；本分支的修改同样以 MIT 发布。

# Issue 06：改造启动参数、限流器和案例注册

## 背景

迁移后不再使用旧 TCP 参数。启动参数应围绕 Mr_Robot WebSocket、心跳和限流配置；登录协议字段在案例中直接传入。

## 目标

清理旧参数，新增 Mr_Robot 非协议参数，注册新案例，并确保限流器默认可用。

## 删除或废弃参数

- `--Tcp-Svr`
- 只服务旧登录协议的自定义参数。

## 新增参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `--ws` | `ws://127.0.0.1:21000` | WebSocket 网关地址。 |
| `--heartbeat-interval` | `10` | 心跳间隔，单位秒。 |
| `--login-rate` | `1` | 登录流程限流阈值。 |

## 新案例名

- `LoginOnline`
- `LoginOnce`

## 建议文件

- `global/globalConf.go`
- `main.go`
- 可选新增：`startup/main_args.go`
- `CaseDefine.go`
- `gamecase/login.go`

## 验收标准

- 不传 Mr_Robot 参数时可使用默认值启动本地网关。
- 启动日志打印 WebSocket 地址、心跳间隔和登录限流阈值。
- 未配置限流时不发生空指针 panic。
- 旧 `LoginOnline` / `LoginOnce` 不再误导用户运行旧协议。

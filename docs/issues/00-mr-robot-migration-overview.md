# Issue 00：Mr_Robot 迁移总览与实施顺序

## 背景

Mr_Robot 已迁移到当前仓库：`Mr_Robot/`。

原始 Protobuf 文件已拿到，后续需要统一放入当前项目的 `gameproto/` 目录，不再维护额外的协议源目录。

因此后续不再从 `Mr_Robot/proto/*_pb.js` 反推协议，也不需要让 Go 直接调用 JS 版 Protobuf。`gameproto/` 是正式协议源和生成代码目录，Go 压测项目应基于 `gameproto/*.proto` 生成 Go 版本协议代码。

## 总体目标

将当前 `demo-server-perf` 改造为 Mr_Robot 项目的 Go 压测机器人：

- 使用 Mr_Robot WebSocket 网关协议。
- 删除旧 TCP 网络连接和旧协议链路。
- 基于 `gameproto/*.proto` 生成 Go 协议代码。
- 实现 Salt 握手、CRC16、XOR、16 字节包头编解码。
- 使用 `sys + msgId` 两级路由。
- 完成登录、创角或选角、角色登录、属性初始化、进入场景、心跳保活。
- 后续逐步接入移动、战斗、宠物、装备、竞技场等行为。

## 明确删除或替换范围

后续实现时可以删除或替换以下旧能力，不需要兼容：

- `module.Robot.CreateConn()` 的旧 TCP 连接路径。
- `msgpackager/` 旧 TCP 包协议。
- `crypto/` 旧 AES CTR 加密路径，如果新协议不使用。
- `module.OnSendPb(pb, msgId uint32)`。
- `module.OnReadPb(msg []byte, protocolID uint32)`。
- `module/login.go` 中旧 `CReqLogin`、`CReqAuth`、`CReqUserInfo`。
- `gamecase/login.go` 中旧登录案例的协议调用逻辑。
- `global.TcpSvr`、`--Tcp-Svr` 等旧 TCP 参数。

应保留或迁移的通用能力：

- `jasmine` 调度器、集合点、限流器。
- `logger` 日志封装。
- `tools` 中账号、路径、环境变量等通用工具。
- `Robot` 的账号身份、异步等待和关闭生命周期模型，但字段可以重构。

## Issue 拆分

建议按以下顺序推进：

1. `01-generate-go-protobuf.md`：将原始 `.proto` 整理到 `gameproto/` 并生成 Go 协议代码。
2. `02-replace-network-codec.md`：实现 Mr_Robot WebSocket codec，替换旧 TCP codec。
3. `03-replace-robot-session-routing.md`：替换 Robot 会话、收发和 `sys + msgId` 路由。
4. `04-mr-robot-login-flow.md`：实现 Mr_Robot 登录状态机。
5. `05-heartbeat-lifecycle.md`：实现心跳和连接生命周期。
6. `06-startup-case-params.md`：改造启动参数、限流器和案例注册。
7. `07-post-login-systems.md`：登录后子系统行为拆分。
8. `08-tests-validation-cleanup.md`：测试、验证和旧代码清理验收。

## 协议依据

- `gameproto/*.proto`：正式协议源文件和 Go 生成代码目录。
- `Mr_Robot/current-project.md`：协议说明。
- `Mr_Robot/standalone_login.js`：JS 登录流程参考。
- `Mr_Robot/lib/Encrypt.js`：Salt、Key、XOR、CRC16 调用参考。
- `Mr_Robot/lib/CRC16.js`：CRC16 实现参考。
- `Mr_Robot/lib/DataPacket.js`：小端读写参考。

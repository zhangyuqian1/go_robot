# Issue 05：实现心跳与连接生命周期管理

## 背景

心跳协议来自 `gameproto/0.default.proto`：

- `c_heartbeat_2 C→S: 0.default msgId=2`
- `s_heartbeat_10 S→C: 0.default msgId=10`

登录完成后需要定时发送心跳，保持 WebSocket 在线。

## 目标

实现登录完成后的心跳、回包处理和机器人退出清理。

## 工作内容

1. 登录完成后启动心跳 goroutine。
2. 心跳包填写 `tick` 字段。
3. 周期默认 `10s`，支持启动参数覆盖。
4. 收到 `s_heartbeat_10` 后可记录服务端 tick 或刷新在线状态。
5. `ctx.Done()`、WebSocket 关闭、机器人 `Close()` 时停止心跳。
6. 心跳发送失败时记录错误并关闭机器人或触发重连策略。

## 建议文件

- 新增：`module/heartbeat.go`
- 修改：`module/robot.go`
- 修改：`gamecase/login.go`

## 验收标准

- 登录完成前不会发送心跳。
- 登录完成后按周期发送 `c_heartbeat_2`。
- context 取消后心跳 goroutine 退出。
- WebSocket 关闭后不会继续写消息。

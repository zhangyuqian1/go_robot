# Issue 03：替换 Robot WebSocket 会话与 sys/msgId 路由

## 背景

当前 `module/robot.go` 包含旧 TCP 连接、旧 WebSocket 发送逻辑和 `uint32 protocolID` 路由。Mr_Robot 应统一使用 WebSocket 和 `sys + msgId` 两级路由。

## 目标

替换 Robot 网络会话，实现 Mr_Robot 握手、收发包、响应分发和异步等待。

## 工作内容

1. 删除或替换 `CreateConn()` TCP 路径。
2. 改造 `CreateWebSocketConn()` 为 Mr_Robot 专用连接入口。
3. WebSocket 打开后发送 ClientSalt。
4. 收到 ServerSalt 后发送 CRC16(Key)，进入正常收发状态。
5. 新增 `sendGatewayRequest`，替代旧 `OnSendPb`。
6. 新增收包循环，调用 `gateway.UnmarshalResponse`。
7. 新增 `routeGatewayResponse(sys, msgId, body)`。
8. 删除或废弃旧 `OnReadPb`。

## 建议接口

```go
func (r *Robot) sendGatewayRequest(requestName string, sys byte, msgId byte, pb proto.Message, wait bool) error
```

```go
func (r *Robot) routeGatewayResponse(sys byte, msgId byte, body []byte)
```

## 首批路由

| sys | msgId | proto 消息 | 处理函数 |
| --- | --- | --- | --- |
| `255` | `1` | `s_login_result_1` | 登录错误码。 |
| `255` | `2` | `s_actor_list_2` | 角色列表。 |
| `255` | `3` | `s_create_actor_result_3` | 创角结果。 |
| `0` | `1` | `s_actor_login_result_1` | 角色登录结果。 |
| `2` | `1` | `s_actor_init_1` | 属性初始化。 |
| `0` | `10` | `s_heartbeat_10` | 心跳回包。 |
| `0` | `13` | `s_actor_enter_scene_13` | 进入场景。 |

## 建议文件

- `module/robot.go`
- 新增：`module/robot_gateway.go`
- 新增：`module/robot_gateway_route.go`
- 新增：`module/robot_gateway_test.go`

## 验收标准

- 新链路不依赖旧 `uint32 msgId`。
- 未知 `sys + msgId` 只记录日志，不 panic。
- 连接关闭时收包 goroutine 和等待事件能退出。
- 发送函数统一处理 Protobuf marshal、组包、写 WebSocket、等待响应。

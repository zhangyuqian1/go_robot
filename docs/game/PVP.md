# 乱斗战场
## 进入乱斗 空协议发送即可无参数
协议id 101 协议名称 c_chaos_battle_field_enter_21
{

}
### 返回跨服路由
协议ID:0，协议名称:s_start_transmit_5{
  "addr": "10.238.20.96",
  "port": 21100
}
## 清除本服角色
协议ID:0，协议名称:c_client_kick_self_5{

}
### 服务器返回要进入的跨服ip及端口
协议ID:0，协议名称:s_start_transmit_5{
  "addr": "10.238.20.96",
  "port": 21100
}
协议ID:0，协议名称:c_client_kick_self_5{
}
###  重新校验密码登录跨服服务器 以下三条都是start.go里已有的协议
协议ID:255，协议名称:c_check_password_1  
{

}
#### 准备登录的角色id
协议ID:0，协议名称:c_client_ready_login_7{
  "actor_id": 16789055,
}
#### 开始登录到跨服
协议ID:0，协议名称:c_actor_login_1{
    "account_id": 16778620,
  "aid": 16789055,
}
## 压测案例切服顺序

独立案例入口为 `gamecase/pvp.go` 的 `PVP`，在 `CaseDefine.go` 中注册为 `PVP`，启动时选择 `--case PVP`，并确保压测配置中的案例名也为 `PVP`。
`LoginOnline` 已恢复为本服登录在线，不再进入乱斗战场或登录跨服。

`PVP` 按以下顺序执行：

1. 本服登录并等待 `s_actor_enter_scene_13`。
2. 发送 `c_chaos_battle_field_enter_21`，等待 `s_start_transmit_5`，最长 30 秒。
3. 使用推送中的 `addr/port` 替换原 WebSocket URL 的主机和端口，保留 `ws/wss`、路径及查询参数。
4. 发送 `c_client_kick_self_5`，调用 `r.Close()`，等待旧收包、心跳和异步监控协程全部退出。
5. 重置登录状态和停止信号，连接跨服 Gateway 并重新握手，依次发送 `c_check_password_1`、`c_client_ready_login_7`、`c_actor_login_1`。沿用原账号、角色，不在跨服创建角色。
6. 等待跨服的 `s_actor_enter_scene_13`，最长 30 秒；成功后进入登录集合点并保持在线，直到上下文取消。

不能在旧连接退出前调用 `Robot_Init(r)` 覆盖连接和停止信号，否则旧收包协程报错后可能关闭新连接。

实现位置：`gameproto/101.arena.proto` 定义空请求，`module/arena.go` 的 `CChaosBattleFieldEnter` 发送系统号 `101`、消息号 `21`，由系统号 `0`、消息号 `5` 的推送结束请求指标并写入 `GameClient.TransmitTarget`。`module/robot_gateway_route.go` 在进场状态缓存完成后通知 `GameClient.LoginSceneReady`。这两个通知均使用容量为 `1` 的通道，重复推送不阻塞收包协程。

`c_client_kick_self_5` 没有对应的请求响应，不建立等待标记。切服前先保留本服账号和角色 ID；旧连接完全退出后清空认证缓存，避免沿用旧值误判认证成功。跨服认证必须返回同一账号，之后使用保留的角色 ID 登录，不创建新角色。异步监控和心跳协程退出后才允许复用机器人连接。
目前跨服认证参数沿用本服案例中的 `120` 等固定值；如果跨服要求不同的 `select_srv/login_srv/raw_srv` 或认证信息，需要按实际协议修改 `gamecase/pvp.go` 的调用参数。
推送的 `addr/port` 必须能直接提供 Gateway WebSocket 服务；若是内部 TCP 地址，需要先映射到客户端可访问的 WebSocket 地址，不能直接拨号。

# 系统模块对照表
| 消息文件名    | 模块名    | 说明                                    |
| ------------- | --------- | --------------------------------------- |
| `0.default`   | 默认/控制 | 登录角色、心跳、进出场景、实体创建/移除 |
| `1.movement`  | 移动      | 角色移动同步                            |
| `2.property`  | 属性      | 角色属性初始化、属性变化同步            |
| `5.skill`     | 技能/战斗 | 普通攻击                                |
| `7.chat`      | 聊天      | GM 指令、聊天消息                       |
| `9.equip`     | 装备      | 装备打开、穿戴、出售                    |
| `15.misc`     | 杂项      | 游戏设置                                |
| `16.observer` | 观察者    | 怪物出现、实体视野变化                  |
| `19.pet`      | 宠物      | 宠物出战、信息、移除                    |
| `101.arena`   | 竞技场    | 竞技场信息、列表、挑战、结果            |
| `255.login`   | 登录/认证 | 账号检查、创角、角色选择                |

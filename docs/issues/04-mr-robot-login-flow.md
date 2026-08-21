# Issue 04：实现 Mr_Robot 登录状态机

## 背景

Mr_Robot 登录协议来自：

- `gameproto/255.login.proto`
- `gameproto/0.default.proto`
- `gameproto/2.property.proto`

当前旧登录链路 `CReqLogin -> CReqAuth -> CReqUserInfo` 可以删除，不需要兼容。

## 目标

实现账号检查、角色列表、创角或选角、角色登录、属性初始化、进入场景的完整状态机。

## 登录流程

```text
WebSocket 连接并完成 Salt 握手
  -> c_check_password_1        C→S 255.login msgId=1
  -> s_actor_list_2            S→C 255.login msgId=2
       -> 有角色：选择第一个 actor_id
       -> 无角色：发送 c_create_actor_4
  -> s_create_actor_result_3   S→C 255.login msgId=3
  -> c_actor_login_1           C→S 0.default msgId=1
  -> s_actor_login_result_1    S→C 0.default msgId=1
  -> s_actor_init_1            S→C 2.property msgId=1
  -> s_actor_enter_scene_13    S→C 0.default msgId=13
  -> 登录完成
```

## 请求字段

### `c_check_password_1`

- `select_srv`：默认 `101`。
- `name`：机器人账号。
- `pwd`：默认 `e10adc3949ba59abbe56e057f20f883e`。
- `login_srv`：默认同 `select_srv`。
- `raw_srv`：默认同 `select_srv`。
- `ip`、`md5`、`tick`、`id`、`channed_id`：默认空或 0，后续按服务端要求补配置。

### `c_create_actor_4`

- `name`：建议由账号派生，避免固定名字冲突。
- `sex`：默认 `1`。
- `job`：默认 `2`。
- `icon`：默认 `1`。
- `token`：默认 `112233`。
- `ad_id`：默认 `0`。

### `c_actor_login_1`

- `account_id`：来自 `s_actor_list_2.account_id`。
- `aid`：有角色时取第一个 `actors.actor_id`；创角后取 `s_create_actor_result_3.aid`。

## 状态缓存

机器人需要缓存：

- `AccountID`
- `ActorID`
- `ActorName`
- `Handle`
- `RawServerID`
- `CurrentServerID`
- `MapData`
- `SceneID`
- `FubenID`
- `SceneType`
- `PX`
- `PY`
- `LoginDone`
- `CreatePending`

## 建议文件

- 删除或替换：`module/login.go`
- 新增：`module/login.go`
- 新增：`gamecase/login.go`
- 修改：`module/robot.go`
- 修改：`CaseDefine.go`
- 新增：`docs/game/login.md`

## 验收标准

- 有角色账号可完成选角登录并进入场景。
- 无角色账号可完成创角后登录并进入场景。
- `s_login_result_1.code`、`s_actor_login_result_1.result`、`s_create_actor_result_3.result` 异常时返回明确错误。
- 创角后仍无法登录时终止流程，不无限循环。
- 登录完成后的状态缓存可供移动、战斗等模块使用。

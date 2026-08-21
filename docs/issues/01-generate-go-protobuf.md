# Issue 01：整理 gameproto 协议源并生成 Go 协议代码

## 背景

原始 `.proto` 已拿到，后续需要统一整理到 `gameproto/`，不继续维护额外的协议源目录。

当前 Go 项目不能直接使用 JS 版 `Mr_Robot/proto/*_pb.js`，应使用 `gameproto/*.proto` 生成 Go 版本代码。

## 目标

将 Mr_Robot 原始 `.proto` 放入 `gameproto/`，在同一协议目录体系下生成 Go 协议代码，并固化生成流程。

## 当前协议文件示例

核心文件包括：

- `gameproto/255.login.proto`：账号检查、角色列表、创角。
- `gameproto/0.default.proto`：角色登录、心跳、进入场景。
- `gameproto/2.property.proto`：角色属性初始化、属性变化、复活相关。
- `gameproto/1.movement.proto`：移动。
- `gameproto/5.skill.proto`：技能和战斗。
- `gameproto/7.chat.proto`：聊天和 GM 指令。
- `gameproto/9.equip.proto`：装备。
- `gameproto/16.observer.proto`：实体视野和观察者。
- `gameproto/19.pet.proto`：宠物。
- `gameproto/101.arena.proto`：竞技场。
- `gameproto/common.proto`：公共结构，包名为 `proto_common`。

## 关键技术点

当前 `.proto` 文件名包含数字和点号，例如 `255.login.proto`、`0.default.proto`。生成 Go 代码前需要处理：

- 为每个 `.proto` 补充或确认 `option go_package`。
- 处理文件名与 Go 包名映射，避免 `default` 关键字冲突。
- 保持 import 关系，例如 `101.arena.proto` import `common.proto`。
- 协议源和生成代码统一纳入 `gameproto/` 目录体系，不再新增独立的 Mr_Robot 协议子目录。

## 建议包名映射

| proto 文件 | Go 包建议 |
| --- | --- |
| `255.login.proto` | `login` |
| `0.default.proto` | `defaultmsg` |
| `2.property.proto` | `property` |
| `1.movement.proto` | `movement` |
| `5.skill.proto` | `skill` |
| `7.chat.proto` | `chat` |
| `9.equip.proto` | `equip` |
| `16.observer.proto` | `observer` |
| `19.pet.proto` | `pet` |
| `101.arena.proto` | `arena` |
| `common.proto` | `common` 或按原始 `package proto_common` 映射为 Go 包 `common` |

## 工作内容

1. 将原始 `.proto` 文件统一放入 `gameproto/`，后续以 `gameproto/` 为唯一协议源目录。
2. 检查 `gameproto/*.proto` 是否都能被 `protoc` 解析。
3. 为需要生成 Go 代码的 `.proto` 增加 `option go_package`，或通过生成脚本传入映射参数。
4. 编写生成脚本，输入和输出都纳入 `gameproto/` 目录体系。
5. 首批只要求生成登录链路需要的协议，后续再扩展全部子系统。
6. 文档记录生成命令、依赖版本和更新规则。

## 首批必须可用消息

- `c_check_password_1`
- `s_actor_list_2`
- `c_create_actor_4`
- `s_create_actor_result_3`
- `c_actor_login_1`
- `s_actor_login_result_1`
- `s_actor_init_1`
- `s_actor_enter_scene_13`
- `c_heartbeat_2`
- `s_heartbeat_10`

## 建议文件

- 整理：`gameproto/*.proto`
- 生成：`gameproto/**/*.pb.go`
- 新增：`tools/gen_proto.ps1` 或 `tools/gen_proto.sh`
- 新增：`docs/game/proto.md`

## 验收标准

- `protoc` 能从 `gameproto/` 中的协议源生成 Go 文件。
- Go 代码能 import 首批必须消息类型。
- 生成流程文档明确，不需要手工复制 JS 版 pb。
- 不手改生成的 `*.pb.go`。

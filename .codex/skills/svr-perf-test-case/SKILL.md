---
name: svr-perf-test-case
description: Use when working in the Go stress-test repository, especially when adding or changing gamecase scenarios, gateway protocol flows, custom CLI parameters, rate limiters, login/map/battle behavior, or project-specific verification and commit workflows. 适用于本项目新增或修改压测案例、协议流程、限流器和自定义参数。
---

# Svr Perf Test Case

这个 skill 是当前仓库“新增协议 + 接入 gamecase + 配置压测入口”的基础规则。

它只解决一件事：在压测项目中中新增或调整一个压测案例时，应该按什么调用链理解项目、改哪些文件、缓存什么状态、如何限流、如何验证。

## 适用边界

本 skill 聚焦当前 Gateway/WebSocket 压测项目的基础接入流程：

- `gameproto/*.pb.go`
- `gateway/codec.go`
- `module/*.go`
- `module/robot.go`
- `module/robot_gateway.go`
- `gamecase/*.go`
- `CaseDefine.go`
- `startup/main_args.go`
- `global/globalConf.go`
- `main.go`

明确边界：

- `docs/game/*` 是专项协议流程补充，新增案例时要读，但不能替代本 skill 的基础接入规则。
- 不要照搬旧协议栈里的 `pb/cs.pb.go`、`OnSendPb`、`OnReadPb`、`config.xml` 写法。
- 当前项目的协议栈是 `gameproto` + Gateway WebSocket + `sendGatewayRequest` + `routeGatewayResponse`。

## 当前项目真实调用链

一个案例从配置到协议收发的链路是：

```text
自定义参数 / YokaConfig
-> main.go 启动并调用 startup.ApplyCustomCaseParams
-> global.BuildRateLimiters 构造命名限流器
-> CaseDefine.go 根据 CaseName 选择 gamecase.Fn
-> gamecase/*.go 编排业务流程和限流
-> module/*.go 构造请求并调用 sendGatewayRequest
-> gateway.MarshalRequest 组 WebSocket 帧
-> module.AesRobotConnRecv 收包并 gateway.UnmarshalResponse
-> module.routeGatewayResponse 按 MsgId 分发
-> CResp* 解析响应、缓存状态、AsyncFire
```

因此必须明确：

- `gamecase` 层负责业务编排，不写协议解析。
- `module` 层负责单协议请求、响应解析、状态缓存和推送处理，不封装完整业务流程。
- `gateway` 层只放消息 ID 和 Gateway 帧编码解码。
- `GameClient` 是当前项目主要运行态缓存载体。
- `global` 只放全局配置、限流器、案例映射等跨机器人配置。

## 输出风格要求

- 说明文字默认使用中文。
- 示例代码里的解释性注释使用中文。
- 路径、参数名、Go 标识符保持原样，例如 `gamecase/`、`--battle-rate`、`global.MustGetRateLimiter`。
- 不写泛化教程，必须基于当前仓库真实路径和真实代码结构。

## 项目结构速查

| 目录或文件 | 职责 |
| --- | --- |
| `CaseDefine.go` | 在 `global.CaseMap` 注册案例名 |
| `gamecase/` | 案例主循环、上下文取消、集合点、限流等待 |
| `module/` | 机器人状态、单协议请求函数、响应解析、推送处理 |
| `gateway/codec.go` | Gateway 消息 ID 常量和帧编码解码 |
| `module/robot_gateway.go` | 按消息 ID 路由到响应处理函数 |
| `startup/main_args.go` | 将自定义参数解析到 `global` 配置 |
| `global/globalConf.go` | 全局运行配置和命名限流器注册表 |
| `gameproto/` | 生成的 Protobuf 代码，不要手改 `*.pb.go` |
| `docs/game/` | 专项游戏协议流程文档 |

## 新增案例必须检查的改动点

新增或修改一个 gamecase，至少检查这些点：

1. `docs/game/*.md`：确认业务流程、消息顺序、是否有推送、哪些推送可忽略。
2. `gameproto/*.pb.go`：确认 `CS*`、`SC*` 类型名和字段 getter。
3. `gateway/codec.go`：补充消息 ID 常量。
4. `module/<feature>.go`：实现请求构造、请求函数、响应解析、状态缓存和推送处理。
5. `module/robot.go`：按后续流程需要补充 `GameClient` 状态字段或等待 channel。
6. `module/robot_gateway.go`：在 `routeGatewayResponse` 中按 `MsgId` 分发。
7. `gamecase/<feature>.go`：实现机器人启动、集合点、限流和循环编排。
8. `CaseDefine.go`：注册案例名到 `global.CaseMap`。
9. `startup/main_args.go`：解析自定义参数，重置默认值。
10. `main.go`：打印关键参数，方便压测启动排查。
11. `*_test.go`：覆盖解析、路由、循环、参数解析和默认值。

## module 层标准写法

`module` 目录下的协议文件统一参考 `module/login.go`：按协议域拆文件，顶部声明系统号/消息号常量和 requestName 常量，中间只放 `C*`/`CReq*` 单协议请求函数，后面放 `handle*GatewayResponse` 路由和具体响应/推送 handler。

`module` 不能承载完整业务编排。类似选择怪物、判断宠物冷却、装备动作选择、竞技场开场/轮询这类流程判断放在 `gamecase/*.go`，由 `gamecase` 根据 `GameClient` 缓存状态直接调用 `module` 的单协议请求函数。

GM/聊天类协议统一放在 `module/gm.go`，例如 `c_send_chat_msg_1`、`s_tipmsg_with_str_2`、`s_send_chat_msg_5`。不要把 GM/聊天请求或回包处理混入 `module/arena.go`、`module/battle.go` 等业务模块文件；业务案例只调用 `r.CSendChatMsg(...)`。

### 请求函数职责

请求函数只做：构造请求、开始异步标记、发包、等待响应或完成指标请求上报。

```go
func (r *Robot) CSomeAction(arg int32) {
	requestName := "CSomeAction"
	req := &some.CSomeAction_1{
		Arg: arg,
	}
	r.AsyncBegin(requestName, 30, true)
	if err := r.sendGatewayRequest(requestName, someSys, someAction, req, false); err != nil {
		r.Mark.Remove(requestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(requestName)
}
```

请求函数必须注意：

- `sendGatewayRequest` 的 requestName 必须和响应里的 `AsyncFire` 名称一致。
- 请求函数不返回 `error`，不设置 `GameClient.*Err`，发包失败只移除 mark、记录日志并 `return`。
- `AsyncBegin(..., true)` 用于需要阻塞等待响应的协议；`false` 用于只发包或后续推送驱动的协议。
- 即使 `isWait=false`，请求函数也保留 `AsyncWait(requestName)`，用于提交请求指标且不会阻塞。
- 请求函数命名跟随协议名，当前项目既有 `CReqCheckPassword`，也有 `CActorMove`、`CNormalAttack`、`CPetBattle`、`CArenaInfo`，不要新增 `RunSingleXxx`、`XxxOnce`、`EnsureXxx` 这类流程封装。
- 请求函数只允许做“构造请求需要的轻量参数规整”，例如下标小于等于 0 时改为 1；业务条件分支放到 `gamecase`。
- 发包成功后，如果该请求会由响应或推送结束生命周期，对应 handler 必须能 `AsyncFire(requestName, ...)`；否则不要 `AsyncBegin`。
- 业务流程不要塞进单协议请求函数，放到对应 `gamecase/*.go` 中按顺序直调 `C*` / `CReq*`。

### 响应处理规则

响应处理函数负责：反序列化、判断响应体状态码、缓存状态、结束等待。

```go
func (r *Robot) handleSomeActionResult(body []byte) {
	requestName := "CReqSomeAction"
	var resp some.SSomeActionResult_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		r.AsyncFire(requestName, 0, STEP_FAIL, 0)
		return
	}

	if resp.GetErrCode() != 0 {
		r.AsyncFire(requestName, int64(len(body)), STEP_FAIL, resp.GetErrCode())
		return
	}

	// 后续流程依赖的数据写入 GameClient。
	r.Client.SomeID = resp.GetSomeId()
	r.AsyncFire(requestName, int64(len(body)), STEP_SUCC, 0)
}
```

响应处理必须遵守：

- 先 `proto.Unmarshal` 解析响应体，再按响应体里的状态码判断是否成功。
- 失败时调用 `AsyncFire(<requestName>, 0/len(body), STEP_FAIL, 状态码)`，并立即 `return`，不能继续上报成功。
- `proto.Unmarshal` 失败属于解析失败，也要 `STEP_FAIL`，此时没有响应体状态码可用时传 `0`。
- 简单响应直接在 handler 中解析和判断状态码，不要为了模板独立封装 `parseSomeResponse`。
- 后续流程依赖的字段要先缓存到 `GameClient`，再 `AsyncFire(..., STEP_SUCC, 0)`。
- 一个响应可能对应多个请求名时，在 handler 内按现有 mark 判断，例如登录的 `s_actor_list_2` 可对应 `CReqCheckPassword` 或 `CReqGetActorList`。
- 服务端推送没有对应请求生命周期时不要强行 `AsyncFire`；只有流程明确等待该推送时才使用独立 requestName。
- `proto.Unmarshal` 失败时，如果该 handler 对应明确请求生命周期，必须 `AsyncFire(requestName, 0, STEP_FAIL, 0)`，避免 mark 留到超时。
- 一个响应或推送可能结束多个请求生命周期时，可以像 `s_actor_enter_scene_13`、`s_pet_update_1` 一样对多个 requestName 调用 `AsyncFire`；没有对应 mark 时 `AsyncFire` 会直接忽略。

### 登录协议特殊约束

登录流程以 `docs/game/登录.md` 为准，当前项目按 `D:\Job\yoka-server-perf\10th\10thsgs-server-perf` 的直调风格处理：

```go
r.CReqCheckPassword(1001, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 1001, 1001, "10.200.255.60", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
if r.Client.ActorID == 0 {
	r.CReqCreateActor(1, 1, fmt.Sprintf("L%d", r.Index), 1, "e10adc3949ba59abbe56e057f20f883e")
	r.CReqGetActorList()
}
r.CReqClientReadyLogin(r.Client.ActorID)
r.CReqActorLogin(r.Client.AccountID, r.Client.ActorID)
```

必须遵守：

- 登录协议参数直接写在 `gamecase` 调用处，不放入 `global/globalConf.go`，也不新增 CLI 自定义参数。
- `gamecase/login.go`、`gamecase/start.go` 直接编排登录协议，不把案例封装进 `module`。
- `module/login.go` 只保留 `CReq*` 发包和 `handle*` 响应处理，不新增 `EnsureLoggedIn`、`runSingleLogin`、`LoginErr`、`failLogin` 等完整登录状态机封装。
- `c_client_ready_login_7`、`c_actor_login_1` 使用 `AsyncBegin(..., false)` 发送，不在请求函数里等待进场推送。

## 缓存规则

如果满足任一条件，要考虑补运行态缓存：

1. 下一个协议依赖本次回包字段。
2. `gamecase` 后续流程要根据本次返回做分支。
3. 推送数据用于解除等待或推进状态机。
4. 查询类协议返回结果会被后续动作复用。

缓存落点规则：

- 单机器人运行态优先放 `GameClient`，例如当前地图、当前位置、BattleID、等待 channel。
- 机器人级别非 Gateway 状态可放 `Robot`，但要先确认现有结构是否已有模式。
- 全局配置才放 `global`，例如 `BattleCfgID`、限流阈值、服务地址。
- 不要把某个机器人的回包结果放入 `global`。

缓存改动时同步检查：

- `module/robot.go` 是否增加字段。
- 响应前是否需要清空旧值，避免读到上一轮结果。
- `gamecase` 是否读取新字段或等待 channel。
- 是否需要在每轮流程开始时重置 channel 和状态。

## gamecase 层标准写法

`gamecase` 只负责编排：启动机器人、登录初始化、集合点、限流等待、直接调用业务协议。

案例格式以 `gamecase/login.go` 和 `gamecase/movement.go` 为准：每个案例入口在自己的文件里直写 `StartFunc()`、`RobotInit(r)`、集合点、`for/select/ctx.Done()` 循环和限流判断。不要为了复用把单案例入口改成通用 `runXxxLoop(ctx, name, fn, idle)` 闭包式模板。

```go
func Xxx(ctx context.Context) {
	r := StartFunc()
	if r == nil {
		return
	}
	defer r.Close()
	RobotInit(r)

	global.LoginRendezvous.Done(-1)
	global.LoginRendezvous.Wait()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("登出")
			return
		default:
			if global.SomeRate == nil || global.SomeRate.Allow() {
				r.CSomeAction(1)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
```

重要不变量：

- 当前代码里的 `StartFunc()` 不接收参数，不要写成不存在的 `StartFunc(true)`。
- 当前代码里的 `RobotInit(r)` 不返回 `error`，不要写成 `if err := RobotInit(r); err != nil`。
- `RobotInit(r)` 已负责 WebSocket 连接和基础登录协议直调；业务案例从登录完成后的状态继续编排。
- 案例不要封装成 `module` 中的 `RunSingleXxx` 或完整流程函数；简单流程直接在对应 `gamecase/*.go` 中按顺序调用 `CReq*`。
- 每个模块的案例写入对应 `gamecase` 文件，例如登录写 `gamecase/login.go`，移动写 `gamecase/move*.go`，战斗写 `gamecase/battle*.go`。
- 战斗案例按 `docs/game/战斗.md` 编排：`RobotInit(r)` 后、登录集合点 `Done/Wait` 前，依次发 `r.CSendChatMsg(0, "@superman")`、`r.CSendChatMsg(0, fmt.Sprintf("@MoveTo 3 %d %d", x, y))` 和 `r.CMinimapBossInfo(3)`；其中 `x,y` 以 `(20,100)` 为中心在正负 10 内随机分散。`r.CSendChatMsg` 的协议实现仍放 `module/gm.go`。`r.CMinimapBossInfo` 回包缓存小地图 boss 的 `id/x/y/relive` 到 `GameClient`；实际攻击坐标以后续 `s_monster_appear_3` 的 `common.id/name/x/y/handle` 为准，当前战斗文档指定挑战 boss `64`（`骸骨巨龙`）。
- 战斗循环优先选择 `s_monster_appear_3.common.id == 64` 或 `name == "骸骨巨龙"` 的存活怪物作为目标；已同步到其它 boss 时不要攻击，等待骸骨巨龙出现。随后在 `global.BattleRate` 限流边界内按顺序直调 `r.CStartUseSkill(r.Client.Handle, skillID, dir, 0)`、`r.CSkillFrameProxy(...)`、`r.CDamageEntity(target.Handle, skillID, skillID*10, 0)`；默认技能按文档使用 `4209`，`range_idx` 为 `42090`。
- 单案例入口不要通过通用 `postLoginStep`、`runPostLoginLoop`、`func(*Robot)` 闭包模板间接启动；需要复用的只保留小的动作函数，例如 `battleOnce`、`petOnce`、`arenaOnce`。
- `PostLoginMixed` 也按 `login.go` / `movement.go` 直写机器人初始化、集合点和循环，多个业务动作在循环里分别判断各自限流器。
- 协议字段如果只是案例固定示例值，直接在 case 调用中传入，不为了“可配置”先声明变量或放入 `global`。
- `gamecase` 不写协议解析、不直接处理 protobuf 回包。
- `ctx.Done()` 必须能结束循环，不能写无限不可取消循环。

## 限流规则

- 自定义限流参数统一使用 `--<module>-rate`，例如 `--login-rate`、`--move-rate`、`--battle-rate`。
- 限流阈值写入对应 `global.<Module>Threshold`，并在 `startup.ApplyCustomCaseParams` 中构造 `global.RateLimiters["<module>"]`。
- 单案例优先直接使用对应字段，例如 `global.MoveRate`、`global.BattleRate`、`global.PetRate`；混合场景也按字段逐个判断，不新增通用闭包式限流模板。
- 一个案例可以使用多个命名限流器，不要按案例名注册限流器。
- 限流检查应放在业务动作边界，例如进入战斗前、执行移动前，不要用 ticker 代替限流器。

## 自定义参数模式

只有“压测运行控制项”才新增自定义参数，例如限流阈值、案例开关、业务配置 ID；协议请求字段默认不要做成自定义参数。

```go
const defaultBattleCfgID = 10000

global.BattleCfgID = defaultBattleCfgID

case "--battle-cfg-id":
	if value, err := strconv.ParseInt(custom.ParamValue, 10, 32); err == nil {
		global.BattleCfgID = int32(value)
	}
```

参数改动必须同步：

- `global/globalConf.go` 只增加跨机器人、跨案例的运行配置字段。
- `startup/main_args_test.go` 增加默认值和显式值测试。
- `main.go` 打印关键配置。
- `docs/game/*.md` 说明参数名、默认值和示例。
- 登录协议字段如 `select_srv`、`login_srv`、`raw_srv`、`ip`、`pwd`、`token`、`channed_id` 不放入 `global`，在 `gamecase` 调用 `CReq*` 时直接传字面值或直接表达式。

## 现有案例模式

| 案例 | 模式 |
| --- | --- |
| `LoginOnline` | `StartFunc` -> `RobotInit` -> 集合点 -> 保持在线到 `ctx.Done()` |
| `Move` | `StartFunc` -> `RobotInit` -> 集合点 -> `move-rate` 边界调用 `CActorSetDir` / `CActorMove` |
| `Battle` | 登录后集合点前发 GM 无敌、GM 传送、`CMinimapBossInfo`；循环中优先攻击 `common.id=64` 或 `name=骸骨巨龙` 的文档指定 boss 或复活，按 `battle-rate` 直调 `CStartUseSkill` / `CSkillFrameProxy` / `CDamageEntity` |
| `Pet` | 登录后循环，`gamecase` 判断宠物缓存和冷却，直调 `CPetBattle` |
| `Equip` | 登录后循环，`gamecase` 根据背包缓存选择查询、穿戴或回收，直调 `CGetBagItems` / `CTakeOnEquip` / `CRecBagItem` |
| `Arena` | 登录后初始化竞技场，循环按缓存状态直调 `CArenaPkList` / `CArenaFight` |
| `PostLoginMixed` / `混合场景` | 单机器人登录后在 `move`、`battle`、`pet`、`equip`、`arena` 多个限流边界轮询 |

## 检查清单

协议层：

- [ ] `gameproto` 中确认了请求、响应和推送类型。
- [ ] `gateway/codec.go` 已补消息 ID。
- [ ] `module/robot_gateway.go` 已补路由。
- [ ] 响应状态码失败时 `AsyncFire` 传入实际状态码。

缓存层：

- [ ] 后续流程依赖的数据已缓存到 `GameClient` 或合适运行态。
- [ ] 每轮流程开始前已重置旧状态或 channel。
- [ ] 没有把单机器人回包结果放到 `global`。

业务层：

- [ ] `gamecase` 只做编排，没有混入协议解析。
- [ ] 循环能响应 `ctx.Done()`。
- [ ] 限流器放在正确业务边界。
- [ ] 登录阶段进入地图的不变量没有被破坏。

入口和参数：

- [ ] `CaseDefine.go` 已注册案例。
- [ ] 自定义参数有默认值重置和测试。
- [ ] `main.go` 打印关键配置。
- [ ] `docs/game/*.md` 已同步参数或流程变化。

## 验证命令

优先跑定向测试，再按需要跑全量测试：

```powershell
$env:GO111MODULE='on'; $env:GOCACHE='D:\Job\yoka-server-perf\c4-server-perf\.gocache'; go test ./startup ./module ./gamecase -run "TestNamePattern" -v
```

全量测试：

```powershell
$env:GO111MODULE='on'; $env:GOCACHE='D:\Job\yoka-server-perf\c4-server-perf\.gocache'; go test ./...
```

如果用户明确要求跳过测试，最终回复必须说明未执行测试。

## 常见问题

发包成功但一直超时，优先排查：

1. `module/robot_gateway.go` 是否补了 `MsgId` 路由。
2. `gateway/codec.go` 的消息 ID 是否正确。
3. `sendGatewayRequest` 的 requestName 和 `AsyncFire` 名称是否一致。
4. 响应 handler 是否在成功和失败分支都调用了 `AsyncFire`。

回包到了，但后续流程拿不到数据，优先排查：

1. 是否补了 `GameClient` 或 `Robot` 缓存字段。
2. 响应 handler 是否真的赋值。
3. 是否需要在每轮开始前清空旧值。
4. `gamecase` 是否在数据到位前提前读取。

场景写完了，但主程序跑不到，优先排查：

1. `CaseDefine.go` 是否注册。
2. `global.CaseMap` 的 key 是否和启动参数 `--case` 一致。
3. 函数签名是否是 `func Xxx(ctx context.Context)`。

## Git 和工作区安全

- 工作区可能包含用户生成的压测报告或其它无关改动。
- 不要回滚无关文件。
- 只暂存用户要求提交的文件。
- 只有用户明确要求时才提交或推送。
- 不要提交 `.gocache/` 等生成缓存目录。

## 最后一条原则

在当前仓库里：

- 新增协议 = 确认 `gameproto` + 补 `gateway` 消息 ID + 补 `module` 请求/响应 + 补 `routeGatewayResponse`。
- 缓存结果 = 判断是否单机器人运行态；是则优先补到 `GameClient`。
- 形成 gamecase = 在 `gamecase` 按业务依赖顺序编排，再注册到 `CaseDefine.go`，必要时补 `startup` 参数。

# 项目阅读理解文档

本文基于当前仓库源码整理，用于快速理解 `demo-server-perf` 的整体职责、启动链路、机器人运行模型、协议收发方式和新增压测案例时的改动入口。

## 一、项目定位

当前项目是一个 Go 编写的游戏机器人压测客户端，核心目标是：

1. 根据命令行参数创建指定数量的机器人协程。
2. 为每个机器人分配账号、建立 TCP 或 WebSocket 连接。
3. 按压测案例执行登录、鉴权、查询用户信息等协议流程。
4. 使用集合点让机器人在关键阶段对齐。
5. 使用限流器控制协议发送速率。
6. 将压测任务信息和运行数据上报到压测平台。

项目模块名为 `stressTest`，Go 版本为 `1.21.1`，主要依赖包括：

- `github.com/gorilla/websocket`：WebSocket 客户端连接。
- `google.golang.org/protobuf`：Protobuf 协议序列化和反序列化。
- `gitlab.qa.dobest.com/qa/yokamancli`：压测平台数据上报客户端。
- `github.com/orcaman/concurrent-map/v2`：机器人异步等待事件表。
- `go.uber.org/zap`：日志输出。

## 二、目录职责

| 路径 | 职责 |
| --- | --- |
| `main.go` | 程序入口，解析参数、初始化日志、注册案例、启动压测调度和指标上报。 |
| `CaseDefine.go` | 将案例名注册到 `global.CaseMap`，例如 `LoginOnline`、`LoginOnce`。 |
| `global/` | 保存全局运行配置、账号区间、案例映射、集合点和限流器。 |
| `jasmine/` | 项目内置压测调度框架，负责参数结构、任务模型、协程启动、限流器和集合点。 |
| `gamecase/` | 压测案例编排层，负责创建机器人、集合点等待、限流控制和循环调用业务协议。 |
| `module/` | 机器人实现层，负责连接管理、协议发送、协议接收分发、异步等待和具体协议处理。 |
| `gameproto/` | Protobuf 生成代码，包含 `cmsg` 等协议结构，通常不要手工修改。 |
| `msgpackager/` | TCP 协议包读写封装，处理包头、消息 ID、加解密和消息体。 |
| `crypto/` | AES CTR 加密封装。 |
| `tools/` | 环境变量、账号文件、路径、哈希、信号处理等工具函数。 |
| `logger/` | 日志初始化与输出封装。 |
| `docs/` | 项目和游戏协议说明文档。 |
| `docker/` | 容器化配置。 |

## 三、启动流程

程序入口在 `main.go`，主流程如下：

```text
init()
  -> 初始化日志
  -> 从环境变量读取 PROJECT_ID / REPORT_ID / ZONE_ID

main()
  -> global.YokaConfig = jasmine.NewYokaConfig()
  -> global.YokaConfig.CommandParse()
  -> GetCommandlineArgs()
  -> tools.MkdirPath()
  -> CaseInit()
  -> 校验 --case 是否为空、是否已注册
  -> YokaMetric()
  -> 创建 global.LoginRendezvous
  -> global.Jasmine = jasmine.NewStandaloneBoomer(...)
  -> global.Jasmine.Run(...)
```

关键文件：

- `main.go`：启动入口。
- `jasmine/parse.go`：基础命令行参数解析。
- `CaseDefine.go`：案例注册。
- `jasmine/perf.go`、`jasmine/runner.go`：压测任务调度。

## 四、参数解析与全局配置

参数解析分两层：

1. `jasmine.YokaConfig.CommandParse()` 解析通用参数。
2. `main.GetCommandlineArgs()` 从 `TestCase.CustomParams` 中解析项目自定义参数。

### 4.1 通用参数

通用参数在 `jasmine/parse.go` 中解析，主要包括：

| 参数 | 含义 |
| --- | --- |
| `--total-robot-num` | 全局机器人总数。 |
| `--pod-num` | 压测 Pod 总数。 |
| `--pod-id` | 当前 Pod 编号。 |
| `--account-on` | 是否从账号文件读取账号。 |
| `--account-filename` | 账号文件路径。 |
| `--account-prefix` | 自动生成账号时使用的前缀。 |
| `--account-idx` | 当前账号起始下标。 |
| `--log-on` | 是否启用命令行指定日志级别。 |
| `--log-level` | 日志级别。 |
| `--case` | 压测案例名。 |
| `--case-robot-num` | 当前案例总用户数。 |
| `--pod-robot-num` | 当前 Pod 用户数。 |
| `--ramp-up-period` | 协程初始化时长参数。 |
| `--duration` | 压测持续时间，单位秒。 |
| `--loop-num` | 执行轮次。 |
| `--timeout` | 协议等待超时时间。 |

### 4.2 自定义参数

自定义参数在 `main.GetCommandlineArgs()` 中解析，当前仅支持非登录协议字段：

| 参数 | 写入位置 | 含义 |
| --- | --- | --- |
| `--creator` | `global.UserName` | 压测报告创建人。 |
| `--ws` | `global.WebSocketSvr` | Mr_Robot WebSocket 网关地址，默认 `ws://127.0.0.1:21000`。 |
| `--heartbeat-interval` | `global.HeartbeatInterval` | 心跳间隔，单位秒。 |
| `--login-rate` | `global.LoginThreshold` | 登录流程限流阈值，默认 `1`，传 `0` 可关闭。 |
| `--project-id` | `global.ProjectId` | 压测平台项目 ID。 |

登录协议字段不再进入 `global` 或自定义参数，案例在 `gamecase` 中直接调用 `CReq*` 时传入。`--login-rate` 大于 0 时会调用 `jasmine.CreateRateLimiter()` 创建 `global.LoginRate`。

## 五、案例注册与调度

### 5.1 案例注册

`CaseDefine.go` 中通过 `global.CaseMap` 注册案例：

```go
global.CaseMap = map[string]*perf.Task{
    "LoginOnline": {Name: "登入在线", Weight: 1, Fn: gamecase.LoginOnline},
    "LoginOnce":   {Name: "登入一次在线", Weight: 1, Fn: gamecase.LoginOnce},
}
```

启动时传入的 `--case` 必须存在于 `global.CaseMap`，否则程序会打印可用案例并退出。

### 5.2 调度模型

`global.Jasmine = jasmine.NewStandaloneBoomer(global.PodRobotNum, 0, nil)` 会创建本地压测调度器。

`global.Jasmine.Run(jasmine.TimerMode, duration, task)` 会按 `--pod-robot-num` 启动对应数量的 goroutine，并给每个案例函数传入 `context.Context`。

在 `TimerMode` 下：

1. `jasmine/runner.go` 创建带取消能力的 `context`。
2. 启动指定数量的 worker goroutine。
3. 每个 worker 根据权重选择任务并执行 `task.Fn(ctx)`。
4. 定时器到期后调用 `cancel()`。
5. 等待 5 秒后退出进程。

## 六、机器人创建与初始化

机器人创建入口在 `gamecase/start.go`。

### 6.1 账号分配

`StartFunc()` 负责创建 `module.Robot` 并分配账号：

1. 如果 `global.GAccountList` 不为空，从账号文件列表中取账号。
2. 否则使用 `global.AccountPrefix + accountSufix` 生成账号名。
3. 账号后缀通过 `global.AccountDispatcher()` 分配。
4. `global.AccountDispatcher()` 使用 `global.GlobalLock` 保证并发安全。

账号区间在 `main.GetCommandlineArgs()` 中计算：

```text
FirstAccount = (TotalRobotNum / PodNum) * PodId + AccountIdx
LastAccount  = FirstAccount + PodRobotNum - 1
```

### 6.2 客户端初始化

`StartFunc()` 会初始化 `module.GameClient`，重要字段包括：
- `WebSocketSvr`：Mr_Robot WebSocket 地址，来自 `global.WebSocketSvr`。
- `OpenTLS`：当前默认关闭 TLS，本地默认连接 `ws://127.0.0.1:21000`。
- `HttpClient`：带 30 秒超时的 HTTP 客户端。

### 6.3 建立连接

`RobotInit(r)` 负责初始化机器人运行信号并调用 `r.CreateConn()` 建立 WebSocket 连接：

```text
RobotInit
  -> 初始化 TickStopSignal / ConnStopSignal / AsyncMonitSignal
  -> r.CreateConn()
```

`module.Robot` 同时具备 TCP 和 WebSocket 连接能力：

- `CreateConn()`：建立 TCP 连接，并启动 TCP 收包协程。
- `CreateWebSocketConn()`：建立 WebSocket 连接，并启动 WebSocket 收包协程。

当前 `gamecase/start.go` 中实际调用的是 `CreateConn()`。

## 七、协议发送、接收与异步等待

当前仓库实际使用的是 `module.OnSendPb()` / `module.OnReadPb()` 这套协议收发路径。

### 7.1 发送流程

以登录协议为例，`module/login.go` 中的 `CReqLogin()` 会：

```text
CReqLogin
  -> 构造 cmsg.CReqLogin
  -> AsyncBegin("CReqLogin", 30, true)
  -> OnSendPb(req, 3750207756)
  -> AsyncWait("CReqLogin")
```

`OnSendPb()` 的职责：

1. 使用 `proto.Marshal()` 序列化请求体。
2. 如果是 TCP 连接，调用 `msgpackager.WriteMsg()` 写入包。
3. 如果是 WebSocket 连接，手工写入 `msgId + msgBody`，加密后发送二进制消息。
4. 发送失败返回 `false`，调用方负责移除等待事件。

### 7.2 TCP 组包格式

TCP 发送由 `msgpackager.WriteMsg()` 完成，当前包头结构大致为：

```text
uint16 messageLength
uint32 messageID
protobuf messageBody
```

其中 `messageLength` 表示后续消息长度，消息体会根据 `GameClient.EnabledCrypto` 进行加密。

### 7.3 接收流程

TCP 接收流程在 `module/robot.go` 中：

```text
CreateConn
  -> 建立 net.Conn
  -> 初始化 AES
  -> go AesRobotConnRecv()

AesRobotConnRecv
  -> 循环 ReadMsg()
  -> 失败时关闭连接并退出

ReadMsg
  -> Packager.ReadMsg()
  -> 解析 msgID 和 msgBody
  -> OnReadPb(msgBody, msgID)
```

`OnReadPb()` 根据响应消息 ID 分发到对应响应函数：

| 响应消息 ID | 处理函数 |
| --- | --- |
| `203388982` | `CRespLogin()` |
| `566867705` | `CRespAuth()` |
| `874691946` | `CRespUserInfo()` |

### 7.4 异步等待模型

每个 `Robot` 上有一个 `Mark`，类型是 `concurrent-map`，用于记录正在等待的协议事件。

核心方法：

| 方法 | 职责 |
| --- | --- |
| `AsyncBegin(requestName, timeOut, isWait)` | 创建协议等待事件，设置超时和是否等待。 |
| `AsyncWait(requestName)` | 阻塞等待响应、失败或超时。 |
| `AsyncFire(requestName, packetLength, code, errCode)` | 响应处理完成后触发事件并记录结果。 |
| `CloseAsyncMonit()` | 停止异步监控信号。 |

协议函数的标准节奏是：

```text
请求函数 AsyncBegin
  -> 发送协议
  -> AsyncWait
响应函数 proto.Unmarshal
  -> 成功或失败时 AsyncFire
```

## 八、当前登录案例

当前仓库注册了两个案例：`LoginOnline` 和 `LoginOnce`。

### 8.1 LoginOnline

文件：`gamecase/login.go`

流程：

```text
StartFunc()
  -> LoginRendezvous.Done(-1)
  -> LoginRendezvous.Wait()
  -> RobotInit(r)
  -> 循环直到 ctx.Done()
       -> global.LoginRate.Allow()
       -> CReqLogin()
       -> CReqAuth()
       -> CReqUserInfo()
       -> sleep 10ms
```

该案例会在整个压测周期内重复执行登录、鉴权、查询用户信息三段协议。

### 8.2 LoginOnce

文件：`gamecase/login.go`

流程：

```text
StartFunc()
  -> LoginRendezvous.Done(-1)
  -> LoginRendezvous.Wait()
  -> RobotInit(r)
  -> 只在第一次满足限流时调用 CReqLogin()
  -> 后续保持在线直到 ctx.Done()
```

该案例只发送一次登录协议，然后保持机器人在线。

## 九、限流器与集合点

### 9.1 限流器

`jasmine.CreateRateLimiter(podId, podNum, threshold)` 会根据全局阈值和 Pod 信息计算当前 Pod 的限流值。

计算逻辑：

```text
base = threshold / podNum
remainder = threshold % podNum
podId < remainder 的 Pod 多分配 1 个额度
```

如果 `threshold < podNum`，当前实现会为每个 Pod 创建每秒 1 次的限流器。

`gamecase/login.go` 中通过 `global.LoginRate == nil || global.LoginRate.Allow()` 判断当前是否允许发送登录相关协议；默认 `--login-rate 1` 会创建限流器，传 `--login-rate 0` 可关闭且不会空指针。

### 9.2 集合点

集合点实现位于 `jasmine/rendezvous.go`。

启动时 `main.go` 创建：

```go
global.LoginRendezvous = jasmine.NewRendezvousPoint("LoginRendezvous", int64(global.PodRobotNum))
```

案例中每个机器人调用：

```go
global.LoginRendezvous.Done(-1)
global.LoginRendezvous.Wait()
```

含义是：所有机器人都到达集合点后，再一起继续执行后续流程。

## 十、指标上报与日志

### 10.1 指标上报

`main.YokaMetric()` 使用 `yokaman.YoKaManCli()` 启动压测平台上报客户端。

主要配置：

- `SetProjectInfo()`：设置项目 ID 和创建人。
- `SetCaseName()`：设置案例名。
- `SetPodInfo()`：设置当前 Pod 用户数和 Pod ID。
- `SetTestInfo()`：设置报告 ID 和 Pod ID。
- `SetEnvOptions()`：当前使用 `yokaman.Test` 环境。

函数中还注册了 `os.Interrupt` 和 `SIGTERM` 信号处理，退出时会调用 `yokacli.Stop()`。

### 10.2 日志

日志默认在 `init()` 中初始化到 `./logs/robot.log`。

如果启动参数带 `--log-on true --log-level N`，则会在 `GetCommandlineArgs()` 中用指定级别重新初始化日志。

## 十一、启动示例

本地连接 demo server 的典型启动参数：

```bash
go run . \
  --total-robot-num 1 \
  --pod-num 1 \
  --pod-id 0 \
  --log-on true \
  --log-level 2 \
  --account-on false \
  --account-filename account.txt \
  --account-prefix test \
  --account-idx 1 \
  --case LoginOnline \
  --case-robot-num 1 \
  --pod-robot-num 1 \
  --timeout 30 \
  --ramp-up-period 0 \
  --duration 30 \
  --count 1 \
  --ws ws://127.0.0.1:21000 \
  --creator "侯宇祺" \
  --login-rate 1
```

Windows PowerShell 单行示例：

```powershell
go run . --total-robot-num 1 --pod-num 1 --pod-id 0 --log-on true --log-level 2 --account-on false --account-filename account.txt --account-prefix test --account-idx 1 --case LoginOnline --case-robot-num 1 --pod-robot-num 1 --timeout 30 --ramp-up-period 0 --duration 30 --count 1 --ws ws://127.0.0.1:21000 --creator "侯宇祺" --login-rate 1
```

## 十二、新增压测案例的改动入口

新增一个案例时，建议按下面顺序检查和修改：

1. `docs/game/*.md`：先确认业务协议流程、请求响应字段、消息 ID 和是否存在服务端推送。
2. `gameproto/`：确认 Protobuf 类型和字段 getter 是否已经生成。
3. `module/<feature>.go`：新增请求函数和响应处理函数。
4. `module/robot.go`：如果有新响应消息 ID，需要在 `OnReadPb()` 中增加分发；如果流程需要运行态缓存，也在 `Robot` 或 `GameClient` 上补字段。
5. `gamecase/<feature>.go`：新增案例编排函数，只负责流程、集合点、限流和循环，不写协议解析。
6. `CaseDefine.go`：将新案例注册到 `global.CaseMap`。
7. `main.go`：如有新自定义参数，在 `GetCommandlineArgs()` 中解析并打印关键配置。
8. `global/globalConf.go`：新增跨机器人共享配置或限流器字段。
9. `*_test.go`：如果周边已有测试模式，应补充参数解析、响应分发或流程函数测试。

推荐分层原则：

- `gamecase` 只编排业务流程。
- `module` 只做单协议请求、响应解析和机器人运行态处理。
- `global` 只放全局配置，不保存单个机器人的业务回包结果。
- `gameproto/*.pb.go` 是生成文件，不要手工修改。
- 新协议等待名称要保持一致，例如 `AsyncBegin("CReqXxx")`、`AsyncWait("CReqXxx")`、`AsyncFire("CReqXxx", ...)` 使用同一个名称。

## 十三、当前代码注意点

1. 当前仓库手写代码中没有 `gateway/` 和 `startup/` 目录，实际协议链路仍是 `OnSendPb()` / `OnReadPb()` / `msgpackager`。
2. `.codex/skills/svr-perf-test-case/SKILL.md` 中提到的 Gateway WebSocket 新协议栈更像是项目后续演进规范；如果新增代码，应先确认当前分支到底要沿用旧链路，还是迁移到 Gateway 链路。
3. `LoginOnline` 和 `LoginOnce` 都使用 `global.LoginRate == nil || global.LoginRate.Allow()`，默认限流器可用，传 `--login-rate 0` 可关闭。
4. `LoginOnline` 的集合点发生在 `RobotInit()` 之前，即所有机器人先对齐，再统一建连。
5. `LoginOnce` 只调用 `CReqLogin()`，不会继续调用 `CReqAuth()` 和 `CReqUserInfo()`。
6. `OnReadPb()` 当前只分发了三个登录相关响应，新协议必须补充分发逻辑，否则响应不会唤醒等待中的请求。
7. `jasmine/runner.go` 的 `TimerMode` 到期后会 `os.Exit(0)`，本地调试时需要注意进程会被直接退出。

## 十四、阅读路线建议

如果是第一次接手该项目，建议按下面顺序阅读：

1. `README.md`：了解框架背景和基础教程。
2. `main.go`：理解启动入口和参数落地。
3. `CaseDefine.go`：确认当前有哪些可运行案例。
4. `jasmine/task.go`、`jasmine/runner.go`、`jasmine/perf.go`：理解压测任务如何变成 goroutine。
5. `gamecase/start.go`：理解机器人如何创建、账号如何分配、连接如何初始化。
6. `gamecase/login.go`：理解案例如何编排协议。
7. `module/robot.go`：理解连接、收包、发包和异步等待。
8. `module/login.go`：理解一个具体协议如何请求、响应和唤醒。
9. `msgpackager/msgpackager.go`：理解 TCP 包格式和加解密位置。
10. `docs/game/current-login.md`：理解当前登录协议细节。

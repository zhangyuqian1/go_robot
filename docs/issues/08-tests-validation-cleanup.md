# Issue 08：测试、验证和旧代码清理验收

## 背景

本次迁移会删除旧 TCP、旧协议和旧登录链路。必须通过测试和清理清单确认新链路完整可用。

## 目标

建立 Mr_Robot 协议迁移后的验证标准，并清理旧代码残留。

## 测试范围

### Protobuf 生成测试

- `gameproto/` 中的协议源可生成 Go 代码。
- 登录链路消息类型可被 Go import。
- 生成脚本重复执行结果稳定。

### Codec 测试

- CRC16。
- Salt 推导 KeyBuff。
- XOR 加解密对称性。
- 包头 marshal/unmarshal。
- CRC 错误、长度错误、包头不足异常路径。

### Module 测试

- `sys + msgId` 路由分发。
- 登录响应解析。
- 有角色路径。
- 无角色创角路径。
- 登录错误码路径。
- 心跳停止。

### Gamecase 测试

- context 取消后机器人退出。
- 限流器为空时不 panic。
- `LoginOnce` 只执行一次登录流程。

## 清理清单

迁移完成后应确认：

- 没有新代码引用 `CreateConn()`。
- 没有新代码引用 `OnSendPb()`。
- 没有新代码引用 `OnReadPb()`。
- 没有新代码引用 `msgpackager.WriteMsg()` / `msgpackager.ReadMsg()`。
- 没有启动参数继续要求 `--Tcp-Svr`。
- 旧 `LoginOnline` / `LoginOnce` 不再误导用户运行旧协议。

## 建议验证命令

```powershell
$env:GO111MODULE='on'; go test ./...
```

如果仓库存在历史无关失败，应拆为：

```powershell
$env:GO111MODULE='on'; go test ./gateway ./module ./gamecase ./global ./jasmine
```

## 验收标准

- 单元测试覆盖 Protobuf 生成、codec、登录状态机和心跳生命周期。
- 本地单机器人可连接 `ws://127.0.0.1:21000` 并进入场景。
- 压测停止后无心跳写入和收包 goroutine 泄漏。
- 文档更新新案例名、启动参数和常见错误排查。

# Issue 02：实现 Mr_Robot WebSocket Codec 并替换旧 TCP Codec

## 背景

Mr_Robot 网络层使用 WebSocket，连接后先 Salt 握手，然后业务包使用 16 字节明文包头和 XOR 加密包体。

当前旧 TCP `msgpackager/`、旧 AES CTR 加密和旧 `uint32 msgId` 包格式不再需要兼容。

## 目标

新增 Mr_Robot 专用 `gateway` codec，替换旧 TCP 包协议。

## 握手协议

```text
Client -> Server: ClientSalt，4 字节，小端，裸数据
Server -> Client: ServerSalt，4 字节，小端，裸数据
Client -> Server: CRC16(Key)，2 字节，小端，裸数据
Key = (ClientSalt ^ ServerSalt) + ClientSalt + 8654
KeyBuff = Key 的 4 字节小端序
```

## 业务包协议

```text
16 字节明文包头 + XOR 加密包体
```

包头：

| offset | 长度 | 字段 | 说明 |
| --- | --- | --- | --- |
| 0 | 2 | `headerCRC` | 包头 CRC16，不包括自身 2 字节。 |
| 2 | 2 | `bodyLen` | 加密包体长度。 |
| 4 | 2 | `bodyCRC` | 包体 CRC16，参与 XOR 混淆。 |
| 6 | 2 | `seqId` | 自增序号，`0~65535`。 |
| 8 | 4 | 保留 | 固定 0。 |
| 12 | 4 | 保留 | 固定 0。 |

包体明文：

```text
sys(1 byte) + msgId(1 byte) + body(N bytes)
```

其中 `body` 为 `2 字节长度 + Protobuf bytes`。

## 工作内容

1. 实现 CRC16，参数与 `Mr_Robot/lib/CRC16.js` 一致。
2. 实现 Salt、Key、KeyBuff 推导。
3. 实现 XOR 加解密。
4. 实现 `MarshalRequest(sys, msgId, protoBytes)`。
5. 实现 `UnmarshalResponse(packetBytes)`。
6. 维护 `seqId` 自增和回绕。
7. 删除或废弃 `msgpackager/` 的旧包协议路径。

## 建议文件

- 新增：`gateway/codec.go`
- 新增：`gateway/crypto.go`
- 新增：`gateway/packet.go`
- 新增：`gateway/codec_test.go`
- 删除或废弃：`msgpackager/*.go`
- 删除或废弃：`crypto/crypto.go`

## 验收标准

- 固定 Salt 能推导稳定 KeyBuff。
- 同一明文业务包 marshal 后再 unmarshal，`sys`、`msgId`、Protobuf bytes 一致。
- CRC 错误、包头不足、包体长度不匹配时返回明确错误。
- 新代码不再调用 `msgpackager.WriteMsg` / `msgpackager.ReadMsg`。


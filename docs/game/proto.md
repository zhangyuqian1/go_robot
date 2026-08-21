# Game Protocol Generation

`gameproto/` is the only source directory for game `.proto` files. Do not copy or maintain generated JavaScript protocol files from `Mr_Robot/proto` for Go code.

## Requirements

- Go: use the version in `go.mod`.
- `protoc`: available on `PATH`, or use `gameproto/protoc.exe` as the fallback.
- `protoc-gen-go`: install with:

```powershell
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
```

The current local verification used `protoc-gen-go v1.36.10`. `gameproto/protoc.exe` reports `libprotoc 3.2.0`; a newer `protoc` on `PATH` is preferred when available.

## Generate Login Protocols

The first supported scope is the login chain:

```powershell
powershell -ExecutionPolicy Bypass -File .\tools\gen_proto.ps1 -Scope Login
```

This generates Go files under:

- `gameproto/login/255.login.pb.go`
- `gameproto/defaultmsg/0.default.pb.go`
- `gameproto/property/2.property.pb.go`

The script parses every `gameproto/*.proto` first, then generates only the selected scope. This catches broken source files before any Go code is emitted.

## Generate All Protocols

After login-chain code is integrated, the same script can generate every top-level protocol file:

```powershell
powershell -ExecutionPolicy Bypass -File .\tools\gen_proto.ps1 -Scope All
```

## Package Mapping

Numeric protocol filenames are mapped to Go packages by stripping the numeric prefix. Special cases are listed explicitly in the script.

| Proto file | Go import path |
| --- | --- |
| `255.login.proto` | `stressTest/gameproto/login` |
| `0.default.proto` | `stressTest/gameproto/defaultmsg` |
| `2.property.proto` | `stressTest/gameproto/property` |
| `1.movement.proto` | `stressTest/gameproto/movement` |
| `5.skill.proto` | `stressTest/gameproto/skill` |
| `7.chat.proto` | `stressTest/gameproto/chat` |
| `9.equip.proto` | `stressTest/gameproto/equip` |
| `16.observer.proto` | `stressTest/gameproto/observer` |
| `19.pet.proto` | `stressTest/gameproto/pet` |
| `101.arena.proto` | `stressTest/gameproto/arena` |
| `common.proto` | `stressTest/gameproto/common` |

`0.default.proto` maps to `defaultmsg` because `default` is not a safe Go package name.

## Update Rules

1. Edit only `gameproto/*.proto` source files.
2. Run `tools/gen_proto.ps1` for the needed scope.
3. Do not hand-edit generated `*.pb.go` files.
4. Verify imports with:

```powershell
go test ./gameproto
```

[CmdletBinding()]
param(
    [ValidateSet("Login", "All")]
    [string]$Scope = "Login",

    [string]$ModulePath = "stressTest",

    [string]$ProtocPath = "",

    [string]$PluginPath = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = (Resolve-Path (Join-Path $scriptDir "..")).Path
$protoDir = Join-Path $repoRoot "gameproto"

$explicitPackages = @{
    "0.default.proto"        = "defaultmsg"
    "1.movement.proto"       = "movement"
    "2.property.proto"       = "property"
    "5.skill.proto"          = "skill"
    "7.chat.proto"           = "chat"
    "9.equip.proto"          = "equip"
    "16.observer.proto"      = "observer"
    "19.pet.proto"           = "pet"
    "101.arena.proto"        = "arena"
    "255.login.proto"        = "login"
    "common.proto"           = "common"
}

function Get-GoPackageName {
    param([string]$FileName)

    if ($explicitPackages.ContainsKey($FileName)) {
        return $explicitPackages[$FileName]
    }

    $name = [System.IO.Path]::GetFileNameWithoutExtension($FileName)
    $name = $name -replace '^\d+\.', ''
    $name = $name -replace '[^A-Za-z0-9_]', '_'
    $name = $name.ToLowerInvariant()

    if ([string]::IsNullOrWhiteSpace($name)) {
        throw "Cannot infer Go package for $FileName"
    }
    if ($name -eq "default") {
        return "defaultmsg"
    }

    return $name
}

function Resolve-Executable {
    param(
        [string]$ExplicitPath,
        [string]$CommandName,
        [string[]]$FallbackPaths
    )

    if (-not [string]::IsNullOrWhiteSpace($ExplicitPath)) {
        $explicitCommand = Get-Command $ExplicitPath -ErrorAction SilentlyContinue
        if ($explicitCommand) {
            return $explicitCommand.Source
        }
        if (Test-Path $ExplicitPath) {
            return (Resolve-Path $ExplicitPath).Path
        }
        throw "Cannot find explicit executable path: $ExplicitPath"
    }

    $command = Get-Command $CommandName -ErrorAction SilentlyContinue
    if ($command) {
        return $command.Source
    }

    foreach ($fallbackPath in $FallbackPaths) {
        if (Test-Path $fallbackPath) {
            return (Resolve-Path $fallbackPath).Path
        }
    }

    throw "Cannot find $CommandName. Install it or pass the explicit path parameter."
}

if (-not (Test-Path $protoDir)) {
    throw "Protocol directory not found: $protoDir"
}

$goCommand = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCommand) {
    throw "Cannot find go on PATH. Install Go before generating protobuf code."
}

$goPath = (& $goCommand.Source env GOPATH).Trim()
$protoc = Resolve-Executable -ExplicitPath $ProtocPath -CommandName "protoc" -FallbackPaths @((Join-Path $protoDir "protoc.exe"))
$protocGenGo = Resolve-Executable -ExplicitPath $PluginPath -CommandName "protoc-gen-go" -FallbackPaths @((Join-Path $goPath "bin\protoc-gen-go.exe"))
$env:PATH = "$(Split-Path -Parent $protocGenGo);$env:PATH"

$allProtoFiles = @(Get-ChildItem -Path $protoDir -Filter "*.proto" | Sort-Object Name)
if ($allProtoFiles.Count -eq 0) {
    throw "No .proto files found in $protoDir"
}

$loginProtoNames = @(
    "255.login.proto",
    "0.default.proto",
    "2.property.proto"
)

foreach ($protoName in $loginProtoNames) {
    if (-not (Test-Path (Join-Path $protoDir $protoName))) {
        throw "Required login protocol source is missing: $protoName"
    }
}

if ($Scope -eq "Login") {
    $selectedProtoNames = $loginProtoNames
} else {
    $selectedProtoNames = @($allProtoFiles | ForEach-Object { $_.Name })
}

$allProtoArgs = @($allProtoFiles | ForEach-Object { "gameproto/$($_.Name)" })
$selectedProtoArgs = @($selectedProtoNames | ForEach-Object { "gameproto/$_" })
$mappingArgs = @($allProtoFiles | ForEach-Object {
    $packageName = Get-GoPackageName $_.Name
    "--go_opt=M$($_.Name)=$ModulePath/gameproto/$packageName;$packageName"
})

Push-Location $repoRoot
try {
    $descriptorPath = Join-Path $env:TEMP "gameproto-parse-check.pb"
    & $protoc --proto_path=gameproto --descriptor_set_out=$descriptorPath @allProtoArgs
    if ($LASTEXITCODE -ne 0) {
        throw "protoc parse check failed with exit code $LASTEXITCODE"
    }

    & $protoc "--plugin=protoc-gen-go=$protocGenGo" --proto_path=gameproto --go_out=. "--go_opt=module=$ModulePath" @mappingArgs @selectedProtoArgs
    if ($LASTEXITCODE -ne 0) {
        throw "Go protobuf generation failed with exit code $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}

Write-Host "Generated $($selectedProtoNames.Count) $Scope protocol file(s) under gameproto/."

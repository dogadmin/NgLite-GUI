# NGLite 代码重构与迁移说明

## 重构概述

本次重构将原有的命令行控制端代码进行了模块化拆分，并在此基础上实现了 GUI 版本。核心原则：

1. **保持原有功能完整性**：所有原 CLI 功能均保留
2. **代码复用最大化**：CLI 和 GUI 共享核心业务逻辑
3. **架构清晰化**：分层设计，职责明确

## 目录结构变化

### 旧结构

```
NGLite-main/
├── lhunter/
│   └── main.go          # 所有控制端逻辑在一个文件
├── lprey/
│   └── main.go          # 被控端
├── module/              # 功能模块
└── conf/                # 配置
```

### 新结构

```
NGLite-main/
├── cmd/                 # 可执行程序入口
│   ├── lhunter-cli/     # CLI 版本
│   ├── lhunter-gui/     # GUI 版本
│   └── lprey/           # 被控端
├── internal/            # 核心业务逻辑
│   ├── core/            # 会话管理、命令分发
│   ├── transport/       # 网络通信
│   ├── config/          # 配置管理
│   └── logger/          # 日志系统
├── gui/                 # GUI 组件
│   ├── widgets/         # UI 组件
│   └── app.go           # GUI 入口
├── module/              # 功能模块（保持不变）
└── conf/                # 配置（保持不变）
```

## 代码迁移映射

### lhunter/main.go → 新架构

| 原代码位置 | 功能 | 新位置 | 说明 |
|-----------|-----|-------|-----|
| `Huntlistener()` | 监听客户端上线 | `internal/transport/listener.go` | 封装为 Listener 类 |
| `BountyHunter()` | 发送命令到客户端 | `internal/core/command_dispatcher.go` | 封装为 CommandDispatcher.SendCommand() |
| `RandomID()` | 生成随机客户端 ID | `internal/transport/nkn_client.go` | 作为私有函数 |
| `AesEncode()` | AES 加密 | `internal/core/command_dispatcher.go` | 内部调用 module/cipher |
| `RsaDecode()` | RSA 解密 | `cmd/lhunter-cli/main.go` 和 `gui/mainwindow.go` | 在回调中使用 |
| seed 生成 | `-n new` 参数 | `internal/config/config.go` | GenerateNewSeed() |
| 配置初始化 | `init()` | `internal/config/config.go` | LoadConfig() / DefaultConfig() |
| 命令行输入循环 | `main()` | `cmd/lhunter-cli/main.go` | CLI 版本保留 |
| 命令行输入循环 | - | `gui/widgets/command_panel.go` | GUI 版本替换为输入框 |

### 新增核心组件

#### SessionManager（会话管理器）

**文件**: `internal/core/session_manager.go`

**职责**:
- 管理所有在线/离线会话
- 提供会话查询、添加、删除、更新接口
- 支持回调机制（上线/离线事件通知）
- 线程安全（使用 sync.RWMutex）

**原 CLI 中的问题**:
- 没有会话持久化，无法查看历史
- 无法管理多个客户端
- 重启后丢失所有信息

**新方案解决**:
- 内存中维护会话列表
- 支持状态跟踪（在线/离线/失联）
- GUI 可实时显示所有会话

#### CommandDispatcher（命令分发器）

**文件**: `internal/core/command_dispatcher.go`

**职责**:
- 封装 `BountyHunter` 的核心逻辑
- 管理命令的加密、发送、接收
- 统一超时处理
- 错误处理和日志记录

**原实现**:
```go
func BountyHunter(seedid string, prey string, command string) string {
    // 创建客户端
    // 加密命令
    // 发送并等待回复
    return result
}
```

**新实现**:
```go
type CommandDispatcher struct {
    transport *transport.TransportManager
    aesKey    string
}

func (cd *CommandDispatcher) SendCommand(preyID, command string) (string, error) {
    // 使用 TransportManager 管理客户端
    // 统一错误处理
    // 超时控制
}
```

#### TransportManager（传输管理器）

**文件**: `internal/transport/nkn_client.go`

**职责**:
- 管理 NKN Account 和 ClientConfig
- 提供统一的客户端创建接口
- 避免重复初始化

**原实现问题**:
- 每次调用都重新创建 Account
- ClientConfig 配置分散
- 代码重复

**新方案**:
```go
type TransportManager struct {
    account      *nkn.Account
    clientConfig *nkn.ClientConfig
}

func (tm *TransportManager) CreateClient(id string) (*nkn.MultiClient, error)
func (tm *TransportManager) CreateRandomClient() (*nkn.MultiClient, error)
```

#### Listener（监听器）

**文件**: `internal/transport/listener.go`

**职责**:
- 封装 `Huntlistener` 逻辑
- 支持启动/停止控制
- 消息回调机制
- 状态管理

**原实现问题**:
- goroutine 无法停止
- 错误只能打印，无法传递
- 无法查询监听状态

**新方案**:
```go
type Listener struct {
    running  bool
    stopChan chan struct{}
}

func (l *Listener) Start(id string, onMessage func(*nkn.Message)) error
func (l *Listener) Stop() error
func (l *Listener) IsRunning() bool
```

#### Logger（日志系统）

**文件**: `internal/logger/logger.go`

**职责**:
- 统一日志输出
- 支持日志级别（INFO/WARN/ERROR）
- 内存缓存日志（最近 1000 条）
- 支持 GUI 实时显示

**原实现**:
```go
log.Println("...")
fmt.Println("...")
```

**新方案**:
```go
type Logger struct {
    entries []LogEntry
    onLog   func(LogEntry)
}

logger.Info("客户端上线")
logger.Error("连接失败")
```

## CLI 版本迁移

### 旧代码（lhunter/main.go）

```go
func main() {
    // 解析参数
    if MakeSeed == "new" {
        // 生成 seed
    }
    
    // 启动监听
    go Huntlistener(Seed)
    
    // 命令行输入循环
    for {
        // 读取输入
        // 调用 BountyHunter
        // 打印结果
    }
}
```

### 新代码（cmd/lhunter-cli/main.go）

```go
func main() {
    // 解析参数
    if makeSeedFlag == "new" {
        seed, _ := config.GenerateNewSeed()
        fmt.Println(seed)
        return
    }
    
    // 初始化核心组件
    cfg := config.DefaultConfig()
    log := logger.NewLogger()
    tm, _ := transport.NewTransportManager(cfg.SeedID, cfg.TransThreads)
    sessionMgr := core.NewSessionManager()
    dispatcher := core.NewCommandDispatcher(tm, cfg.AESKey)
    
    // 启动监听
    listener := transport.NewListener(tm)
    go listener.Start(cfg.HunterID, onClientOnline)
    
    // 命令行输入循环（保持不变）
    for {
        // 读取输入
        result, _ := dispatcher.SendCommand(preyID, command)
        fmt.Println(result)
    }
}
```

**变化**:
1. 引入核心组件（SessionManager、CommandDispatcher 等）
2. 更清晰的错误处理
3. 统一的日志输出
4. 会话状态管理
5. 可复用的业务逻辑

## GUI 版本实现

### 主要文件

#### cmd/lhunter-gui/main.go

```go
func main() {
    cfg := config.LoadConfig(configPath)
    fyneApp := app.NewWithID("com.nglite.hunter.gui")
    guiApp := gui.NewApp(fyneApp, cfg)
    guiApp.ShowAndRun()
}
```

#### gui/app.go

持有所有核心组件，提供给 GUI 组件访问：

```go
type App struct {
    sessionMgr *core.SessionManager
    dispatcher *core.CommandDispatcher
    transport  *transport.TransportManager
    logger     *logger.Logger
    listener   *transport.Listener
}
```

#### gui/mainwindow.go

主窗口，整合所有 UI 组件：

- 顶部工具栏：Start/Stop/New Seed
- 左侧：会话列表
- 右侧：Tab 区域（控制台/日志/配置）

#### gui/widgets/

- `session_list.go`: 会话列表组件
- `command_panel.go`: 命令交互面板
- `log_viewer.go`: 日志查看器

### GUI 与核心逻辑交互

```
GUI 组件
  ↓ 用户点击 "Start"
MainWindow.onStartListener()
  ↓ 调用
App.listener.Start()
  ↓ 使用
TransportManager.CreateClient()
  ↓ NKN 收到消息
MainWindow.onClientOnline()
  ↓ 调用
SessionManager.AddSession()
  ↓ 回调触发
SessionListWidget.Refresh()
  ↓ 更新
GUI 显示新会话
```

## 兼容性说明

### 完全兼容

1. **被控端（lprey）**: 完全不变，可以与新旧控制端通信
2. **通信协议**: RSA + AES 加密方式不变
3. **NKN 网络**: 使用相同的 seed 和网络参数

### 功能增强

1. **会话管理**: GUI 可查看所有会话，CLI 需手动记录
2. **状态监控**: GUI 实时显示在线/离线，CLI 无此功能
3. **日志系统**: GUI 可查看历史日志，CLI 只能看终端输出

### 使用方式

**旧 CLI**:
```bash
cd lhunter
go run main.go -g <seed>
```

**新 CLI**:
```bash
./bin/lhunter-cli -g <seed>
```

**新 GUI**:
```bash
./bin/lhunter-gui
```

## 测试建议

### 功能测试

1. **seed 生成**:
```bash
./bin/lhunter-cli -n new
```

2. **CLI 启动**:
```bash
./bin/lhunter-cli -g <seed>
```

3. **GUI 启动**:
```bash
./bin/lhunter-gui
```

4. **被控端连接**:
```bash
./bin/lprey -g <seed>
```

5. **命令执行**:
- CLI: 输入 `<preyid> whoami`
- GUI: 选中会话 → 输入 `whoami` → Execute

### 兼容性测试

1. 旧 CLI 控制端 + 旧被控端
2. 新 CLI 控制端 + 旧被控端
3. 新 GUI 控制端 + 旧被控端
4. 新 CLI/GUI 混合使用（同一 seed）

## 后续优化方向

1. **配置持久化**: 保存 GUI 配置到文件
2. **会话导出**: 导出会话列表为 CSV/JSON
3. **命令历史**: 保存命令历史记录
4. **批量操作**: 对多个会话批量执行命令
5. **插件系统**: 支持自定义功能扩展
6. **文件传输**: 实现文件上传/下载
7. **端口转发**: 实现内网穿透功能

## 总结

本次重构：
- ✅ 保留了所有原有功能
- ✅ 提供了两种使用方式（CLI + GUI）
- ✅ 代码结构更清晰，易于维护
- ✅ 为未来功能扩展奠定基础
- ✅ 完全向后兼容


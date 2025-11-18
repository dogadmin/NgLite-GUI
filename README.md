# NGLite GUI 版本使用文档

原版 https://github.com/Maka8ka/NGLite。 很好的一个项目，拿来学习并且尝试改成gui做测试

## 项目结构

```
NGLite-main/
├── cmd/
│   ├── lhunter-cli/    # CLI 控制端
│   ├── lhunter-gui/    # GUI 控制端（新）
│   └── lprey/          # 被控端
├── internal/           # 核心业务逻辑（新）
│   ├── core/          # 会话管理、命令分发
│   ├── transport/     # NKN 通信层
│   ├── config/        # 配置管理
│   └── logger/        # 日志系统
├── gui/               # GUI 组件（新）
│   ├── widgets/       # UI 组件
│   └── app.go         # GUI 应用入口
├── module/            # 功能模块
│   ├── cipher/        # 加密
│   ├── command/       # 命令执行
│   └── getmac/        # 主机信息
└── conf/              # 配置常量
```

## 编译说明

### 前置要求

- Go 1.18+
- Fyne 依赖（已自动安装）

### 编译 CLI 版本

```bash
go build -o bin/lhunter-cli ./cmd/lhunter-cli
go build -o bin/lprey ./cmd/lprey
```

### 编译 GUI 版本

#### macOS

```bash
go build -o bin/lhunter-gui ./cmd/lhunter-gui
```

#### Linux

```bash
go build -o bin/lhunter-gui ./cmd/lhunter-gui
```

#### Windows

```bash
go build -o bin/lhunter-gui.exe ./cmd/lhunter-gui
```

### 打包 GUI 应用（可选）

使用 Fyne 工具打包为应用程序：

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest

fyne package -os darwin -name "NGLite Hunter" -src ./cmd/lhunter-gui
fyne package -os linux -name "NGLite Hunter" -src ./cmd/lhunter-gui
fyne package -os windows -name "NGLite Hunter" -src ./cmd/lhunter-gui
```

## 使用说明

### CLI 版本使用

#### 控制端（lhunter-cli）

**生成新 seed：**
```bash
./bin/lhunter-cli -n new
```

**启动控制端：**
```bash
./bin/lhunter-cli -g <seed_id>
```

**发送命令：**
```
> aa:bb:cc:dd:ee:ff192.168.1.100 whoami
> aa:bb:cc:dd:ee:ff192.168.1.100 pwd
```

#### 被控端（lprey）

```bash
./bin/lprey -g <seed_id>
```

### GUI 版本使用

#### 启动 GUI 控制端

```bash
./bin/lhunter-gui
```

#### 主要功能

**1. 启动监听**
- 点击顶部 "Start" 按钮启动监听
- 状态栏显示 "● 监听中" 表示成功

**2. 查看在线会话**
- 左侧显示所有上线的客户端
- 显示信息：PreyID、MAC、IP、OS、状态、最后上线时间

**3. 执行命令**
- 点击左侧会话列表选中目标
- 右侧 "会话控制台" Tab 中输入命令
- 点击 "Execute" 或按 Enter 执行
- 输出区域显示命令结果

**4. 查看日志**
- 切换到 "全局日志" Tab
- 查看所有系统事件（上线、命令、错误）
- 点击 "Clear Logs" 清空日志

**5. 配置管理**
- 切换到 "配置设置" Tab
- 查看当前配置信息
- 点击 "New Seed" 按钮生成新频道

**6. 停止监听**
- 点击 "Stop" 按钮停止监听

## 功能对比

| 功能 | CLI 版本 | GUI 版本 |
|-----|---------|---------|
| 生成 Seed | ✅ `-n new` | ✅ 按钮 |
| 指定频道 | ✅ `-g <seed>` | ✅ 配置面板 |
| 接收上线 | ✅ 日志输出 | ✅ 会话列表 |
| 发送命令 | ✅ 命令行输入 | ✅ GUI 输入框 |
| 查看历史 | ❌ | ✅ 输出区保留 |
| 会话管理 | ❌ | ✅ 列表管理 |
| 状态监控 | ❌ | ✅ 实时状态 |
| 日志查看 | ✅ 终端输出 | ✅ 独立 Tab |

## 核心逻辑重构

原 `lhunter/main.go` 中的功能已拆分为：

| 原功能 | 新位置 |
|--------|-------|
| `Huntlistener` | `internal/transport/listener.go` |
| `BountyHunter` | `internal/core/command_dispatcher.go` |
| Session 管理 | `internal/core/session_manager.go` |
| NKN 客户端 | `internal/transport/nkn_client.go` |
| 配置读取 | `internal/config/config.go` |
| 日志系统 | `internal/logger/logger.go` |

CLI 和 GUI 版本共享所有核心逻辑，仅交互层不同。

## 技术栈

- **语言**: Go 1.18+
- **GUI 框架**: Fyne v2.7+
- **P2P 通信**: NKN SDK
- **加密**: RSA 2048 + AES-CBC 256

## 配置文件

GUI 版本支持通过 JSON 配置文件启动：

```bash
./bin/lhunter-gui -c config.json
```

**config.json 示例：**
```json
{
  "seed_id": "fa801f84020cadc6914ef9b11482b4ccaf09e5cc282e77881c38bdded436cc75",
  "hunter_id": "monitor",
  "aes_key": "whatswrongwithUu",
  "trans_threads": 4
}
```

## 常见问题

**Q: GUI 无法启动？**

A: 确保系统已安装图形界面支持：
- macOS: 原生支持
- Linux: 需要 X11 或 Wayland
- Windows: 原生支持

**Q: 编译 GUI 时报错？**

A: 检查 Fyne 依赖：
```bash
go get fyne.io/fyne/v2@latest
go mod tidy
```

**Q: 如何在无图形界面服务器上运行？**

A: 使用 CLI 版本：
```bash
./bin/lhunter-cli -g <seed>
```

**Q: 会话列表不刷新？**

A: 会话列表每 3 秒自动刷新，也可手动点击 "刷新" 按钮

**Q: 命令执行超时？**

A: 默认超时 30 秒，检查：
- 被控端是否在线
- 网络连接是否正常
- PreyID 是否正确

## 后续开发计划

- [ ] 文件传输功能
- [ ] 端口转发/内网穿透
- [ ] 批量命令执行
- [ ] 命令历史记录保存
- [ ] 会话分组管理
- [ ] 插件系统

## 许可证

与原项目保持一致




------------------------------
更新目录，支持了文件管理，socks5代理。代理感觉有点问题
窗口自适应。


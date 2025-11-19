# NGLite GUI 版本

> 致敬原版：[https://github.com/Maka8ka/NGLite](https://github.com/Maka8ka/NGLite)  
> 本项目基于原版NGLite进行学习和改进，增加了GUI界面和文件管理功能

## ⚠️ 免责声明

**本项目仅供安全研究和学习使用！**

- ✅ 本工具设计用于合法的系统管理和运维工作
- ✅ 使用前请确保获得目标系统所有者的明确授权
- ❌ **严禁用于任何非法用途，包括但不限于未授权访问、恶意攻击、数据窃取等**
- ❌ 使用本工具产生的一切法律后果由使用者自行承担
- ❌ 项目作者和贡献者不对任何滥用行为负责

**请遵守所在国家/地区的法律法规，合法合规使用本工具！**

---

## 项目简介

NGLite GUI 版本是基于区块链P2P网络（NKN）的跨平台远程管理工具，支持：

- 🖥️ **跨平台支持**：Windows / macOS / Linux
- 🔒 **匿名通信**：基于NKN区块链网络，无需公网IP
- 🎨 **图形界面**：提供现代化GUI和传统CLI两种模式
- 📁 **文件管理**：远程浏览、上传、下载文件
- 💻 **命令执行**：远程执行系统命令
- 🔐 **加密传输**：RSA + AES双重加密
- 🌐 **SOCKS5代理**：通过被控端网络上网（⚠️ 实验性功能，不稳定）

## 编译

### 前置要求

- Go 1.18+
- Fyne 依赖（GUI版本需要）

### 编译所有版本

```bash
# Mac/Linux 版本
go build -o bin/lprey cmd/lprey/main.go
go build -o bin/lhunter-gui cmd/lhunter-gui/main.go
go build -o bin/lhunter-cli cmd/lhunter-cli/main.go

# Windows 版本
GOOS=windows GOARCH=amd64 go build -o bin/lprey.exe cmd/lprey/main.go
GOOS=windows GOARCH=amd64 go build -o bin/lhunter-cli.exe cmd/lhunter-cli/main.go
```

## 快速开始

### 1. 生成频道 Seed

```bash
# 使用CLI生成
./bin/lhunter-cli -n new

# 或使用GUI，点击 "New Seed" 按钮
./bin/lhunter-gui
```

### 2. 启动被控端（Prey）

```bash
# 在被控机器上
./bin/lprey -g <你的seed>
```

### 3. 启动控制端（Hunter）

#### GUI 模式（推荐）

```bash
./bin/lhunter-gui
```

1. 点击 "Start" 启动监听
2. 等待被控端上线（左侧会话列表）
3. 选择会话，执行命令或管理文件

#### CLI 模式

```bash
./bin/lhunter-cli -g <你的seed>

# 交互式命令格式
> <preyid> <命令>
```

## 主要功能

### 会话管理
- 实时显示在线被控端
- 显示MAC地址、IP、操作系统、最后上线时间
- 自动检测离线超时

### 命令执行
- 远程执行任意系统命令
- 实时显示命令输出
- 命令历史保存

### 文件管理
- 远程浏览文件系统
- 列出所有磁盘/盘符
- 上传/下载文件（支持≤10MB）
- 删除文件和目录
- 文件大小显示

### SOCKS5代理 ⚠️ 实验性功能
- 通过被控端网络进行上网
- 本地监听SOCKS5端口（如1080）
- 流量通过NKN加密隧道转发
- **警告：此功能不稳定，可能无法正常工作**
- **已知问题**：
  - NKN连接初始化慢（30-60秒）
  - 连接可能超时失败
  - 延迟高（100-300ms）
  - 不适合高带宽应用
  - 需要稳定的网络环境

#### SOCKS5使用步骤
1. 在GUI中选择在线会话
2. 切换到"SOCKS5代理"标签页
3. 输入本地端口（如1080）
4. 点击"启动代理"
5. 等待连接建立（可能需要1-2分钟）
6. 配置浏览器使用SOCKS5代理：127.0.0.1:1080

#### 测试命令
```bash
# 测试代理是否工作
curl --socks5 127.0.0.1:1080 https://ifconfig.me
```

### 全局日志
- 记录所有操作事件
- 实时日志查看
- 日志清理功能

## 技术特点

### 通信机制
- 基于 NKN（New Kind of Network）区块链P2P网络
- 约8万个分布式节点
- 无需中心服务器
- 无需公网IP或域名

### 安全加密
- RSA 2048位非对称加密（初始握手）
- AES-CBC 256位对称加密（数据传输）
- 防止中间人攻击

### 匿名性
- P2P直连通信
- 无中心服务器日志
- 流量混淆于NKN网络
- （注：完全匿名性取决于网络环境）

## 配置说明

GUI版本支持配置文件：

```bash
./bin/lhunter-gui -c config.json
```

**config.json 示例：**
```json
{
  "seed_id": "your_seed_here",
  "hunter_id": "monitor",
  "aes_key": "your_aes_key",
  "trans_threads": 4
}
```

## 项目结构

```
NGLite/
├── cmd/              # 命令行入口
│   ├── lhunter-gui/ # GUI控制端
│   ├── lhunter-cli/ # CLI控制端
│   └── lprey/       # 被控端
├── gui/             # GUI组件
│   └── widgets/     # UI控件
├── internal/        # 核心逻辑
│   ├── core/       # 会话管理、命令分发
│   ├── transport/  # NKN通信层
│   ├── config/     # 配置管理
│   └── logger/     # 日志系统
├── module/          # 功能模块
│   ├── cipher/     # 加密解密
│   ├── command/    # 命令执行
│   ├── fileops/    # 文件操作
│   └── getmac/     # 主机信息
└── conf/           # 配置常量
```

## 优势与劣势

### ✅ 优势
- 理论上完全匿名（除非监测所有中间节点）
- 无需购买服务器、域名、CDN等资源
- 无需实名认证
- 跨平台支持
- 免杀性能良好
- GUI界面友好

### ❌ 劣势
- 连接较多（P2P网络特性）
- 体积较大（可用UPX压缩）
- 首次连接需要等待
- 依赖网络环境
- SOCKS5代理功能不稳定（实验性）

## 常见问题

**Q: 被控端无法连接？**  
A: 检查网络连接，确保能访问NKN网络节点

**Q: 文件上传失败？**  
A: 目前仅支持≤10MB的文件

**Q: GUI无法启动？**  
A: 确保系统支持图形界面（X11/Wayland/Quartz）

**Q: 编译错误？**  
A: 运行 `go mod tidy` 安装依赖

**Q: SOCKS5代理无法连接？**  
A: 这是实验性功能，经常失败。建议：
- 等待更长时间（1-2分钟）
- 检查两端网络是否稳定
- 查看终端日志排查问题
- 重启代理重试
- 如果仍然失败，此功能可能暂时不可用

## 致谢

- 原版项目：[Maka8ka/NGLite](https://github.com/Maka8ka/NGLite)
- NKN 区块链网络：[https://nkn.org/](https://nkn.org/)
- Fyne GUI 框架：[https://fyne.io/](https://fyne.io/)

## 许可证

MIT License

---

## 📢 再次声明

**本项目仅供学习研究，严禁用于非法用途！**

使用本工具前请确保：
1. ✅ 已获得目标系统所有者的明确书面授权
2. ✅ 用途符合当地法律法规
3. ✅ 理解并承担所有使用后果

**一切违法使用产生的后果由使用者自行承担！**

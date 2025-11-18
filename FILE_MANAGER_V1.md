# NGLite 文件管理功能 - 第一阶段

## ✅ 已实现功能（第一阶段）

### 1. 列出盘符
- **功能**：获取被控端所有可用磁盘驱动器
- **Windows**：显示 A-Z 所有可用盘符（C:\、D:\ 等）
- **Linux/macOS**：显示根目录 /

### 2. 列出目录
- **功能**：浏览被控端文件系统
- **显示信息**：
  - 文件/文件夹名称
  - 文件大小（格式化显示 KB/MB/GB）
  - 文件/文件夹图标（📁 文件夹 / 📄 文件）

### 3. 目录导航
- **功能**：在文件系统中浏览
- **操作**：
  - 点击文件夹进入子目录
  - "上级目录" 按钮返回父目录
  - 显示当前路径

## 使用方法

### GUI 界面使用

#### 1. 启动 GUI 控制端
```bash
./bin/lhunter-gui
```

#### 2. 连接被控端
- 点击 "Start" 启动监听
- 在被控端运行 `lprey`（自动上线）
- 左侧会话列表会显示新上线的机器

#### 3. 使用文件管理器
1. **选择会话**：在左侧列表点击目标机器
2. **切换到文件管理器 Tab**
3. **操作步骤**：
   - 点击 "列出盘符" → 查看所有驱动器
   - 点击某个盘符 → 进入该盘符根目录
   - 点击文件夹 → 进入子目录
   - 点击 "上级目录" → 返回父目录
   - 点击 "刷新" → 重新加载当前目录

#### 4. 界面布局
```
┌────────────────────────────────────────────┐
│ 文件管理器 Tab                              │
├────────────────────────────────────────────┤
│ 当前路径: C:\Users\Admin\                  │
│ [列出盘符] [刷新] [上级目录]               │
├────────────────────────────────────────────┤
│ 📁 Desktop (4.2 KB)                        │
│ 📁 Documents (<DIR>)                       │
│ 📁 Downloads (<DIR>)                       │
│ 📄 file.txt (1.5 MB)                       │
│ 📄 readme.md (3.2 KB)                      │
├────────────────────────────────────────────┤
│ 操作日志：                                  │
│ 正在加载: C:\Users\Admin\                  │
│ 加载完成，共 15 项                          │
└────────────────────────────────────────────┘
```

### CLI 命令行使用

虽然主要是 GUI 功能，但也可通过 CLI 发送 JSON 命令：

#### 列出盘符
```bash
> aa:bb:cc:dd:ee:ff192.168.1.100 {"action":"list_drives"}
```

返回：
```json
{
  "drives": ["C:\\", "D:\\"],
  "success": true
}
```

#### 列出目录
```bash
> aa:bb:cc:dd:ee:ff192.168.1.100 {"action":"list_dir","path":"C:\\"}
```

返回：
```json
{
  "path": "C:\\",
  "files": [
    {
      "name": "Users",
      "path": "C:\\Users",
      "size": 0,
      "is_dir": true,
      "mod_time": "2025-11-18T10:00:00Z"
    }
  ],
  "success": true
}
```

## 技术实现

### 新增模块

#### 1. `module/fileops/fileops.go`
文件操作核心模块：
- `ListDrives()` - 列出所有盘符
- `ListDirectory(path)` - 列出目录内容
- `GetFileInfo(path)` - 获取文件信息
- `HandleFileCommand(json)` - 处理 JSON 格式命令

#### 2. `gui/widgets/file_manager.go`
GUI 文件管理器组件：
- 文件列表显示
- 目录导航
- 操作按钮
- 日志输出

### 协议设计

#### 命令格式（JSON）
```json
{
  "action": "list_drives|list_dir|get_info",
  "path": "可选，目标路径"
}
```

#### 响应格式（JSON）
```json
{
  "success": true|false,
  "error": "错误信息（如果失败）",
  "drives": [...],    // list_drives
  "files": [...],     // list_dir
  "path": "当前路径"
}
```

### 数据流程

```
GUI 点击 → CommandDispatcher.ListDrives(preyID)
    ↓
生成 JSON 命令：{"action":"list_drives"}
    ↓
AES 加密 → NKN 发送
    ↓
被控端接收 → AES 解密
    ↓
识别 JSON 格式 → fileops.HandleFileCommand()
    ↓
执行 ListDrives() → 返回 JSON 结果
    ↓
AES 加密 → NKN 返回
    ↓
控制端接收 → AES 解密 → 解析 JSON
    ↓
更新 GUI 文件列表显示
```

## 编译版本

### macOS 版本
- ✅ `bin/lhunter-cli` - CLI 控制端
- ✅ `bin/lhunter-gui` - GUI 控制端
- ✅ `bin/lprey` - 被控端

### Windows 版本
- ✅ `bin/lhunter-cli.exe` - CLI 控制端
- ✅ `bin/lprey.exe` - 被控端

## 跨平台支持

| 功能 | Windows | Linux | macOS |
|-----|---------|-------|-------|
| 列出盘符 | ✅ A-Z | ✅ / | ✅ / |
| 列出目录 | ✅ | ✅ | ✅ |
| 文件大小 | ✅ | ✅ | ✅ |
| 目录导航 | ✅ | ✅ | ✅ |

## 后续功能（第二阶段预览）

### 即将实现
- [ ] 文件下载（被控端 → 控制端）
- [ ] 文件上传（控制端 → 被控端）
- [ ] 文件删除
- [ ] 文件重命名
- [ ] 创建目录
- [ ] 文件搜索
- [ ] 大文件分块传输
- [ ] 传输进度显示

## 测试建议

### 1. 本地测试（macOS）
```bash
# 终端 1: 启动 GUI
./bin/lhunter-gui

# 终端 2: 启动被控端
./bin/lprey
```

### 2. 跨机器测试（Windows 被控）
```bash
# macOS: 启动 GUI
./bin/lhunter-gui

# Windows: 启动被控端
lprey.exe
```

### 3. 测试步骤
1. ✅ 启动控制端和被控端
2. ✅ 确认被控端上线（会话列表显示）
3. ✅ 选择会话 → 切换到文件管理器 Tab
4. ✅ 点击 "列出盘符" 查看驱动器
5. ✅ 点击盘符进入目录
6. ✅ 浏览文件和文件夹
7. ✅ 测试 "上级目录" 导航
8. ✅ 测试 "刷新" 功能

## 注意事项

1. **权限问题**
   - 某些系统目录可能需要管理员权限
   - 被控端以当前用户权限运行

2. **路径格式**
   - Windows: `C:\Users\Admin\`
   - Linux/macOS: `/home/user/`

3. **大目录加载**
   - 包含大量文件的目录可能加载较慢
   - 30 秒超时限制

4. **网络延迟**
   - NKN P2P 网络延迟取决于节点距离
   - 通常在 1-5 秒内响应

## 更新日志

### v1.0.0 - 2025-11-18
- ✅ 实现盘符列表功能
- ✅ 实现目录浏览功能
- ✅ 实现文件信息显示
- ✅ 添加 GUI 文件管理器 Tab
- ✅ 支持 Windows/Linux/macOS
- ✅ 跨平台编译完成

---

**下一步**：第二阶段将实现文件上传/下载功能！


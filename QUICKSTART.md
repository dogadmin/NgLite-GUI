# NGLite 快速开始指南

## 5 分钟上手 GUI 版本

### 步骤 1: 编译程序

```bash
cd /Users/misfai/Desktop/NGLite-main

./build.sh
```

编译完成后，`bin/` 目录将包含三个可执行文件：
- `lhunter-cli` - CLI 控制端
- `lhunter-gui` - GUI 控制端  
- `lprey` - 被控端

### 步骤 2: 启动 GUI 控制端

```bash
./bin/lhunter-gui
```

### 步骤 3: 启动监听

1. 在 GUI 窗口中点击顶部的 **"Start"** 按钮
2. 等待状态栏显示 **"状态: ● 监听中"**
3. 查看 "全局日志" Tab，确认显示 "监听器已启动"

### 步骤 4: 启动被控端（模拟客户端上线）

在另一个终端窗口：

```bash
./bin/lprey
```

被控端会自动连接到默认频道并上线。

### 步骤 5: 查看上线的客户端

在 GUI 中：
1. 左侧 "在线会话列表" 应该显示新上线的客户端
2. 显示信息包括：MAC 地址、IP 地址、操作系统、状态等

### 步骤 6: 执行命令

1. 点击左侧会话列表中的客户端（单击选中）
2. 右侧切换到 "会话控制台" Tab
3. 在底部输入框输入命令，例如：
   ```
   whoami
   ```
4. 按 **Enter** 或点击 **"Execute"** 按钮
5. 输出区域将显示命令执行结果

### 步骤 7: 查看日志

切换到 "全局日志" Tab，可以看到：
- 监听器启动事件
- 客户端上线事件  
- 命令发送事件
- 错误信息等

## 使用自定义频道

### 方法 1: 生成新 seed

```bash
./bin/lhunter-cli -n new
```

复制输出的 64 位十六进制字符串。

### 方法 2: 使用新 seed

**GUI 方式**:
1. 关闭当前 GUI
2. 编辑 `config.json`:
```json
{
  "seed_id": "YOUR_NEW_SEED_HERE",
  "hunter_id": "monitor",
  "aes_key": "whatswrongwithUu",
  "trans_threads": 4
}
```
3. 启动 GUI: `./bin/lhunter-gui -c config.json`

**CLI 方式**:
```bash
./bin/lhunter-cli -g YOUR_NEW_SEED_HERE
./bin/lprey -g YOUR_NEW_SEED_HERE
```

## 常用命令示例

### Windows 被控端

```
whoami
hostname
ipconfig
dir C:\
systeminfo
tasklist
```

### Linux 被控端

```
whoami
hostname
ifconfig
pwd
ls -la
uname -a
ps aux
```

## 同时管理多个客户端

1. 在不同机器上启动多个被控端（使用相同 seed）
2. GUI 会话列表将显示所有客户端
3. 点击不同客户端即可切换目标
4. 每个会话的命令历史独立显示

## CLI 版本快速使用

如果你更喜欢命令行：

```bash
./bin/lhunter-cli

> aa:bb:cc:dd:ee:ff192.168.1.100 whoami
> aa:bb:cc:dd:ee:ff192.168.1.100 pwd
```

格式：`<MAC地址><IP地址> <命令>`

## 停止程序

**GUI**:
- 点击 "Stop" 按钮停止监听
- 关闭窗口退出程序

**CLI**:
- `Ctrl + C` 停止程序

**被控端**:
- `Ctrl + C` 停止程序

## 故障排除

### GUI 启动失败

**macOS**:
```bash
xcode-select --install
go get fyne.io/fyne/v2@latest
```

**Linux**:
```bash
sudo apt-get install libgl1-mesa-dev xorg-dev
go get fyne.io/fyne/v2@latest
```

### 客户端无法上线

1. 检查网络连接
2. 确认控制端和被控端使用相同 seed
3. 查看日志中的错误信息

### 命令执行超时

1. 检查客户端是否在线（状态为 "● 在线"）
2. 确认 PreyID 正确
3. 网络延迟可能较高，等待 30 秒超时

### 会话列表为空

1. 确认已点击 "Start" 按钮
2. 确认状态显示 "● 监听中"
3. 确认被控端已启动并连接成功
4. 点击 "刷新" 按钮手动刷新列表

## 下一步

- 阅读 [GUI_README.md](GUI_README.md) 了解完整功能
- 阅读 [MIGRATION.md](MIGRATION.md) 了解架构设计
- 查看源码了解实现细节
- 参与后续功能开发

## 安全提示

- 仅在授权的环境中使用
- 妥善保管 seed 和密钥
- 定期更换通信频道
- 遵守当地法律法规


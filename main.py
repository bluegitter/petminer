#!/usr/bin/env python3
"""
MiningPet 命令行启动器
兼容README中描述的 python main.py 启动方式
"""

import subprocess
import sys
import os
import time
import signal

def check_requirements():
    """检查运行环境"""
    # 检查Go
    try:
        subprocess.run(["go", "version"], capture_output=True, check=True)
        print("✅ Go环境检测成功")
    except (subprocess.CalledProcessError, FileNotFoundError):
        print("❌ 错误: 未检测到Go，请先安装Go 1.21+")
        return False
    
    # 检查Node.js
    try:
        subprocess.run(["node", "--version"], capture_output=True, check=True)
        print("✅ Node.js环境检测成功")
    except (subprocess.CalledProcessError, FileNotFoundError):
        print("❌ 错误: 未检测到Node.js，请先安装Node.js 16+")
        return False
    
    return True

def main():
    print("🐾 MiningPet - 命令行挂机游戏")
    print("================================")
    
    if not check_requirements():
        sys.exit(1)
    
    # 启动后端
    print("🚀 启动后端服务器...")
    os.chdir("backend")
    
    # 安装Go依赖
    subprocess.run(["go", "mod", "tidy"], check=True)
    
    # 启动后端进程
    backend_process = subprocess.Popen(
        ["go", "run", "cmd/server/main.go"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE
    )
    
    os.chdir("..")
    
    # 等待后端启动
    print("⏳ 等待后端服务器启动...")
    time.sleep(3)
    
    # 检查后端状态
    try:
        import urllib.request
        urllib.request.urlopen("http://localhost:8080/api/v1/pets")
        print("✅ 后端服务器启动成功")
    except:
        print("❌ 后端服务器启动失败")
        backend_process.terminate()
        sys.exit(1)
    
    # 如果有前端，也启动前端
    if os.path.exists("frontend"):
        print("🎨 启动前端开发服务器...")
        os.chdir("frontend")
        
        # 检查并安装依赖
        if not os.path.exists("node_modules"):
            print("📦 安装前端依赖...")
            subprocess.run(["npm", "install"], check=True)
        
        # 启动前端
        frontend_process = subprocess.Popen(["npm", "start"])
        os.chdir("..")
        
        print("✅ 前端开发服务器启动成功")
        print("")
        print("🎮 游戏已启动!")
        print("前端地址: http://localhost:3000")
        print("后端API: http://localhost:8080")
    else:
        frontend_process = None
        print("🎮 后端服务器已启动!")
        print("API地址: http://localhost:8080")
    
    print("")
    print("示例输出:")
    print("[Pet-Lucky] 开始探索北方森林……")
    print("[遭遇] 碰到了一只野猪，进入战斗……")
    print("[胜利] 宠物获胜，获得肉块*2，经验+5")
    print("[社交] 遇到 Alice 的宠物，成为了朋友！")
    print("[惊喜] 发现神秘矿石 → 奖励已到账")
    print("")
    print("按 Ctrl+C 停止游戏")
    
    # 信号处理
    def signal_handler(sig, frame):
        print("\n🛑 正在停止游戏...")
        backend_process.terminate()
        if frontend_process:
            frontend_process.terminate()
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    
    # 等待进程
    try:
        backend_process.wait()
    except KeyboardInterrupt:
        signal_handler(signal.SIGINT, None)

if __name__ == "__main__":
    main()
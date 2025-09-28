# 开发指南

## 项目结构

```
miningpet/
├── backend/                 # Go后端
│   ├── cmd/server/         # 服务器入口
│   ├── internal/           # 内部包
│   │   ├── handlers/       # HTTP处理器
│   │   ├── models/         # 数据模型
│   │   └── services/       # 业务逻辑
│   └── pkg/               # 公共包
│       └── websocket/     # WebSocket支持
├── frontend/              # React前端
│   ├── src/
│   │   ├── components/    # React组件
│   │   ├── hooks/         # 自定义Hook
│   │   └── services/      # API服务
└── docs/                  # 文档

```

## 快速开始

### 方式1：使用启动脚本
```bash
./start.sh
```

### 方式2：使用Python启动器（兼容README）
```bash
python main.py
```

### 方式3：手动启动

**启动后端：**
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

**启动前端：**
```bash
cd frontend
npm install
npm start
```

## API文档

### 宠物管理
- `POST /api/v1/pets` - 创建宠物
- `GET /api/v1/pets` - 获取所有宠物
- `GET /api/v1/pets/:id` - 获取特定宠物
- `POST /api/v1/pets/:id/explore` - 开始探索

### 事件系统
- `GET /api/v1/events` - 获取事件历史
- `WS /ws` - WebSocket实时事件流

## 核心特性

### 1. 宠物系统
- 随机生成名字和性格
- 等级、经验、属性成长
- 5种性格类型影响战斗

### 2. 探索机制
- 每10秒触发一次随机事件
- 事件类型：探索、战斗、发现、社交、奖励

### 3. 战斗系统
- 轻量级回合制（自动结算）
- 基于属性和性格的战斗力计算

### 4. 奖励系统
- 常规奖励：金币、经验、道具
- 稀有掉落：5%概率获得大奖

### 5. 实时更新
- WebSocket推送所有事件
- 前端实时显示宠物状态

## 技术栈

**后端：**
- Go 1.21+
- Gin Web框架
- Gorilla WebSocket
- 内存存储（可扩展为数据库）

**前端：**
- React 18
- Tailwind CSS
- Lucide图标
- WebSocket客户端

## 开发建议

### 添加新事件类型
1. 在 `models/event.go` 中定义事件类型
2. 在 `services/pet_service.go` 中实现事件逻辑
3. 在前端 `Terminal.jsx` 中添加颜色支持

### 扩展宠物属性
1. 修改 `models/pet.go` 中的Pet结构
2. 更新前端 `PetCard.jsx` 显示
3. 调整战斗力计算公式

### 持久化存储
- 实现数据库接口
- 添加Redis缓存层
- 修改服务层以支持持久化

## 性能优化

- 事件历史限制为1000条
- WebSocket连接池管理
- 前端事件列表虚拟滚动
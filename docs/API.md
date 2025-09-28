# MiningPet API 文档

## 基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **WebSocket**: `ws://localhost:8080/ws`

## API端点

### 1. 创建宠物

**POST** `/pets`

创建一个新的宠物。

**请求体:**
```json
{
  "owner_name": "张三"
}
```

**响应:**
```json
{
  "id": "uuid",
  "name": "Lucky",
  "owner": "张三",
  "personality": "brave",
  "level": 1,
  "experience": 0,
  "health": 100,
  "max_health": 100,
  "attack": 10,
  "defense": 5,
  "coins": 0,
  "location": "起始村庄",
  "status": "等待中",
  "last_activity": "2023-12-07T10:30:00Z",
  "created_at": "2023-12-07T10:30:00Z"
}
```

### 2. 获取所有宠物

**GET** `/pets`

获取系统中所有宠物的列表。

**响应:**
```json
{
  "pets": [
    {
      "id": "uuid",
      "name": "Lucky",
      // ... 宠物完整信息
    }
  ]
}
```

### 3. 获取特定宠物

**GET** `/pets/{id}`

根据ID获取特定宠物的详细信息。

**响应:**
```json
{
  "id": "uuid",
  "name": "Lucky",
  // ... 宠物完整信息
}
```

### 4. 开始探索

**POST** `/pets/{id}/explore`

让指定宠物开始探索。

**响应:**
```json
{
  "message": "Exploration started"
}
```

### 5. 获取事件历史

**GET** `/events?limit=50`

获取最近的事件记录。

**查询参数:**
- `limit`: 返回事件数量限制，默认50

**响应:**
```json
{
  "events": [
    {
      "id": "uuid",
      "pet_id": "uuid",
      "pet_name": "Lucky",
      "type": "battle",
      "message": "[Lucky] 击败了野猪！获得经验+15，金币+5",
      "timestamp": "2023-12-07T10:35:00Z",
      "data": {
        "enemy": "野猪",
        "is_victory": true,
        "experience": 15,
        "coins": 5
      }
    }
  ]
}
```

## WebSocket 事件

连接到 `ws://localhost:8080/ws` 以接收实时事件更新。

**消息格式:**
```json
{
  "type": "event",
  "data": {
    "id": "uuid",
    "pet_id": "uuid",
    "pet_name": "Lucky",
    "type": "discovery",
    "message": "[Lucky] 发现了宝箱，获得20金币！",
    "timestamp": "2023-12-07T10:40:00Z",
    "data": {
      "coins": 20
    }
  }
}
```

## 事件类型

| 类型 | 描述 | 数据字段 |
|------|------|----------|
| `explore` | 探索新区域 | `location` |
| `battle` | 战斗事件 | `enemy`, `is_victory`, `experience`, `coins` |
| `discovery` | 发现宝物 | `coins` |
| `social` | 社交互动 | `friend_name` |
| `reward` | 普通奖励 | `coins` |
| `rare_find` | 稀有发现 | `coins` |
| `level_up` | 等级提升 | `new_level` |

## 性格类型

| 性格 | 英文 | 战斗加成 | 特点 |
|------|------|----------|------|
| 勇敢 | brave | +5 | 战斗力较强 |
| 贪婪 | greedy | +2 | 更容易获得金币 |
| 友好 | friendly | +0 | 社交事件较多 |
| 谨慎 | cautious | +3 | 防御能力强 |
| 好奇 | curious | +1 | 探索事件较多 |

## 错误码

| 状态码 | 错误信息 | 描述 |
|--------|----------|------|
| 400 | Bad Request | 请求参数错误 |
| 404 | Pet not found | 宠物不存在 |
| 500 | Internal Server Error | 服务器内部错误 |
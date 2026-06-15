# 📖 408 考研择校助手

基于 Go Agent 的智能择校与信息查询系统，面向计算机考研（408 科目）学生。

## 1. 项目概述

项目采用 **Go 语言** 自研轻量级 **ReAct (Reasoning and Acting) 大模型 Agent 算力环**，通过 **Tool Calling（工具调用）** 驱动后端关系型数据库，解决传统大模型在"考研分数线、招生人数"等硬性数据上频繁出现的幻觉问题，为用户提供精准、可追溯的择校建议。

---

## 2. 系统技术栈

| 层级 | 技术 |
|------|------|
| 前端 | 微信小程序原生开发 + Vant Weapp 组件库 |
| 网关与业务后端 | Go 1.22+ / Fiber v3 |
| 持久层 ORM | GORM v2 |
| 关系型数据库 | MySQL 8.0 |
| 缓存与状态机 | Redis 7.0 |
| AI 核心 | 大模型原生 API (DeepSeek-V3 / Qwen-Plus) + 手写 Tool Calling 分发器 |

---

## 3. 项目结构

```
cs-assistant/
├── docker-compose.yml           # MySQL + Redis 本地开发环境
├── backend/                     # Go 后端
│   ├── cmd/server/main.go       # 入口点
│   ├── config/
│   │   ├── config.yaml          # 配置文件
│   │   └── config.go            # 配置结构体与加载
│   └── internal/
│       ├── agent/               # ReAct Agent 控制循环
│       ├── cache/redis.go       # Redis 连接与健康检查
│       ├── db/mysql.go          # GORM 连接与自动迁移
│       ├── handler/auth.go      # 登录接口
│       ├── middleware/auth.go   # Bearer Token 鉴权中间件
│       ├── model/               # 数据模型 (GORM + API 类型)
│       ├── thirdparty/wechat.go # 微信 code2Session
│       └── tool/                # Tool Calling 工具定义与执行
├── miniprogram/                 # 微信小程序前端
└── docs/                        # 项目文档
```

---

## 4. 快速开始

### 4.1 环境要求

- Go 1.22+
- Docker Desktop（用于运行 MySQL 和 Redis）

### 4.2 启动开发环境

```bash
# 1. 启动 MySQL 8.0 + Redis 7.0
docker compose up -d

# 2. 确认服务健康
docker compose ps

# 3. 启动后端
cd backend
go run ./cmd/server/

# 4. 验证
curl http://localhost:8080/health
# → ok
```

### 4.3 配置

编辑 `backend/config/config.yaml`，或通过环境变量覆盖敏感字段：

| 环境变量 | 对应配置 |
|----------|----------|
| `MYSQL_PASSWORD` | MySQL 密码 |
| `REDIS_PASSWORD` | Redis 密码 |
| `LLM_API_KEY` | 大模型 API Key |
| `WECHAT_APP_ID` | 微信小程序 AppID |
| `WECHAT_APP_SECRET` | 微信小程序 AppSecret |

---

## 5. API 接口

### 5.1 健康检查

```
GET /health
```

### 5.2 微信登录

```
POST /api/v1/auth/login
Content-Type: application/json

{"code": "wx.login() 返回的临时 code"}

→ { "code": 0, "message": "ok", "data": { "token": "...", "expires_at": "..." } }
```

### 5.3 流式对话接口（待实现）

```
POST /api/v1/chat/stream
Authorization: Bearer <token>
Content-Type: application/json

{"message": "我想考北京的211，408专业课，去年复试线330分以下的学校有哪些？"}

→ Content-Type: text/event-stream

data: {"type": "thought", "content": "用户正在寻找北京、211、408科目、分数线 <= 330 的学校..."}
data: {"type": "action", "content": "正在为您查询数据库..."}
data: {"type": "chunk", "content": "根据查询，为您找到以下符合条件的北京211高校：\n1. ..."}
data: [DONE]
```

---

## 6. 数据库 Schema (MySQL)

### 6.1 学校信息表 (`schools`)

```sql
CREATE TABLE `schools` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL UNIQUE COMMENT '学校名称',
    `region` VARCHAR(20) NOT NULL COMMENT '省份/地区',
    `tier` VARCHAR(10) NOT NULL COMMENT '档次: 985/211/双非',
    `is_985` TINYINT(1) DEFAULT 0 COMMENT '是否985',
    `is_211` TINYINT(1) DEFAULT 0 COMMENT '是否211',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='高校基本信息表';

ALTER TABLE `schools` ADD INDEX `idx_region_tier` (`region`, `tier`);
```

### 6.2 历年招录数据表 (`major_scores`)

```sql
CREATE TABLE `major_scores` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `school_id` INT UNSIGNED NOT NULL COMMENT '关联schools表id',
    `year` INT NOT NULL COMMENT '年份',
    `major_code` VARCHAR(20) NOT NULL COMMENT '专业代码',
    `major_name` VARCHAR(50) NOT NULL COMMENT '专业名称',
    `score_line` INT NOT NULL COMMENT '复试线',
    `intake_num` INT NOT NULL COMMENT '招生人数',
    `note` TEXT COMMENT '备注',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`school_id`) REFERENCES `schools`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='历年招录明细表';

ALTER TABLE `major_scores` ADD INDEX `idx_school_year_major` (`school_id`, `year`, `major_code`);
```

### 6.3 用户表 (`users`)

```sql
CREATE TABLE `users` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `open_id` VARCHAR(64) NOT NULL UNIQUE COMMENT '微信用户唯一标识',
    `nickname` VARCHAR(64) DEFAULT '' COMMENT '用户昵称',
    `avatar_url` VARCHAR(512) DEFAULT '' COMMENT '头像 URL',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
```

> **注意：** 启动时 GORM AutoMigrate 会自动创建/更新这三张表，无需手动执行 DDL。

---

## 7. Redis 缓存与会话设计

| Key 设计 | 数据结构 | TTL | 业务场景 |
| --- | --- | --- | --- |
| `session:${token}` | String (JSON) | 7 天 | 微信小程序登录态，存储 `UserSession{user_id, open_id}` |
| `chat:history:${session_id}` | List | 1 小时 | 大模型多轮对话上下文窗口 |
| `rate:limit:${openid}:${api}` | String | 1 分钟 | 接口防刷限流 |

---

## 8. Agent 核心架构与 ReAct 工作流

在 Go 后端，放弃使用 Python 的 LangChain 框架，完全基于 **Go 结构体与大模型接口规范** 实现 ReAct。

### 8.1 工具定义 (Go Struct)

通过标准 JSON 标签定义大模型能识别的"工具定义"：

```go
type Tool struct {
    Type     string   `json:"type"` // 固定为 "function"
    Function Function `json:"function"`
}

type Function struct {
    Name        string     `json:"name"`
    Description string     `json:"description"`
    Parameters  Parameters `json:"parameters"`
}

type Parameters struct {
    Type       string                 `json:"type"` // 固定 "object"
    Properties map[string]interface{} `json:"properties"`
    Required   []string               `json:"required"`
}
```

### 8.2 ReAct 控制循环核心逻辑

```go
func ExecuteAgentLoop(ctx context.Context, userInput string, history []Message) (string, error) {
    // 1. 组装 System Prompt (赋予其择校专家身份，并注入可用工具列表)
    // 2. 将 history + userInput 投喂给大模型
    // 3. 进入 ReAct 状态循环（限制最大循环次数 MaxLoops = 3，防止死循环）
    for i := 0; i < MaxLoops; i++ {
        resp, err := CallLLMWithTools(messages, availableTools)

        // 判定大模型是否需要调用工具
        if resp.FinishReason != "tool_calls" {
            // 大模型判定不需要查数据库了，直接给出了最终回复
            return resp.Content, nil
        }

        // 大模型给出了 Action：决定调用某个工具
        toolCall := resp.ToolCalls[0]
        var observation string

        // 动态路由分发工具调用
        switch toolCall.Function.Name {
        case "query_school_scores":
            args := parseArgs(toolCall.Function.Arguments)
            observation = tools.QuerySchoolScoresFromDB(db, args)
        case "compare_schools":
            observation = tools.CompareSchools(db, parseArgs(toolCall.Function.Arguments))
        }

        // 将 Action 和 Observation 作为新一轮上下文 Append 到 messages 中
        messages = append(messages, Message{Role: "assistant", ToolCalls: resp.ToolCalls})
        messages = append(messages, Message{Role: "tool", Content: observation, ToolCallID: toolCall.ID})
    }
    return "", errors.New("Agent 思考环超出最大步数限制")
}
```

---

## 9. 微信登录流程

```
小程序 wx.login()
      │
      ▼  code
POST /api/v1/auth/login
      │
      ▼  code
微信 code2Session API ──→ openid
      │
      ▼
MySQL: SELECT → 不存在则 INSERT (自动注册)
      │
      ▼
crypto/rand 生成 64 位 hex token
      │
      ▼
Redis: SET session:<token> = {user_id, open_id}  TTL 7 天
      │
      ▼
返回 { token, expires_at }
```

后续请求携带 `Authorization: Bearer <token>`，中间件从 Redis 还原用户身份并注入请求上下文，每次验证自动滑动续期。

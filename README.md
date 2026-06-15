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
    `city` VARCHAR(30) NOT NULL DEFAULT '' COMMENT '城市',
    `tier` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '档次: 985/211/双一流/双非',
    `is_985` TINYINT(1) DEFAULT 0 COMMENT '是否985',
    `is_211` TINYINT(1) DEFAULT 0 COMMENT '是否211',
    `is_double_first_class` TINYINT(1) DEFAULT 0 COMMENT '是否双一流',
    `is_408` TINYINT(1) DEFAULT 0 COMMENT '是否考408统考',
    `is_self_score` TINYINT(1) DEFAULT 0 COMMENT '是否34所自主划线',
    `cs_rank` VARCHAR(10) DEFAULT '' COMMENT '计算机学科评估',
    `official_url` VARCHAR(256) DEFAULT '' COMMENT '学校研招网',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='高校基本信息表';

ALTER TABLE `schools` ADD INDEX `idx_region_tags` (`region`, `is_985`, `is_211`, `is_double_first_class`);
```

### 6.2 历年招录数据表 (`admission_records`)

```
学校 → 学院 → 学硕/专硕 → 专业方向 → 年份
├─ 一志愿核心数据（独立列）
├─ 复试政策（权重/机试/统一复试）
└─ 调剂数据（JSON，多批次）
```

```sql
CREATE TABLE `admission_records` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `school_id` INT UNSIGNED NOT NULL COMMENT '关联schools表id',
    `college_name` VARCHAR(50) NOT NULL COMMENT '学院名称: 计算机科学与技术学院',

    `degree_type` TINYINT UNSIGNED NOT NULL COMMENT '1=学硕, 2=专硕',
    `major_code` VARCHAR(10) NOT NULL COMMENT '专业代码: 085400',
    `major_name` VARCHAR(50) NOT NULL COMMENT '专业名称: 电子信息',
    `direction_code` VARCHAR(10) DEFAULT '' COMMENT '方向代码: 01',
    `direction_name` VARCHAR(50) DEFAULT '不区分研究方向' COMMENT '方向名称: 计算机视觉',

    `year` INT NOT NULL COMMENT '招录年份',

    -- 一志愿核心数据
    `first_choice_score_line` INT NOT NULL COMMENT '一志愿复试分数线',
    `first_choice_retest_num` INT DEFAULT 0 COMMENT '一志愿复试人数',
    `first_choice_actual_num` INT DEFAULT 0 COMMENT '一志愿录取人数',
    `first_choice_avg_score` DECIMAL(5,2) DEFAULT 0.00 COMMENT '一志愿录取均分',

    `national_line` INT NOT NULL DEFAULT 0 COMMENT '当年国家线',
    `planned_intake` INT DEFAULT 0 COMMENT '计划招生人数（含推免）',
    `first_choice_intake` INT DEFAULT 0 COMMENT '一志愿录取人数',
    `transfer_intake` INT DEFAULT 0 COMMENT '调剂录取人数',
    `exemption_intake` INT DEFAULT 0 COMMENT '推免录取人数',

    -- 初试科目（核心筛选维度）
    `exam_english_type` TINYINT UNSIGNED DEFAULT 1 COMMENT '1=英一, 2=英二',
    `exam_math_type` TINYINT UNSIGNED DEFAULT 1 COMMENT '1=数一, 2=数二, 3=无',
    `exam_cs_type` TINYINT UNSIGNED DEFAULT 1 COMMENT '1=统考408, 2=自命题',
    `exam_cs_name` VARCHAR(50) DEFAULT '408计算机学科专业基础' COMMENT '专业课科目名，如: 822 计算机基础综合',
    `sub_has_ds` TINYINT(1) DEFAULT 0 COMMENT '数据结构',
    `sub_has_os` TINYINT(1) DEFAULT 0 COMMENT '操作系统',
    `sub_has_co` TINYINT(1) DEFAULT 0 COMMENT '计算机组成原理',
    `sub_has_cn` TINYINT(1) DEFAULT 0 COMMENT '计算机网络',
    `sub_has_other` TINYINT(1) DEFAULT 0 COMMENT '其他（离散/软工）',

    -- 复试政策
    `is_joint_retest` TINYINT(1) DEFAULT 0 COMMENT '一志愿与调剂是否统一复试',
    `initial_weight` TINYINT UNSIGNED DEFAULT 50 COMMENT '初试权重（%）',
    `retest_weight` TINYINT UNSIGNED DEFAULT 50 COMMENT '复试权重（%）',
    `has_machine_test` TINYINT(1) DEFAULT 0 COMMENT '是否有上机考试',
    `machine_test_software` VARCHAR(50) DEFAULT '' COMMENT '上机考试软件/环境',

    -- 调剂数据（JSON 列）
    `transfer_info` JSON COMMENT '调剂批次详情 [{batch_name, score_line, retest_num, actual_num, avg_score, source_schools}]',

    `note` TEXT COMMENT '备注: 单科线、408改考说明等',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`school_id`) REFERENCES `schools`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='院校专业招录明细表';

ALTER TABLE `admission_records` ADD INDEX `idx_college_degree_year` (`school_id`, `college_name`, `degree_type`, `year`);
```

### 6.3 复试学生明细表 (`retest_rosters`)

关联 `admission_records`，记录每位复试学生的初试各科分数、复试成绩、录取状态。

```sql
CREATE TABLE `retest_rosters` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `admission_record_id` INT UNSIGNED NOT NULL COMMENT '关联admission_records表id',

    -- 身份与来源
    `candidate_no` VARCHAR(20) DEFAULT '' COMMENT '考生编号（脱敏）',
    `student_name` VARCHAR(20) DEFAULT '匿名考生' COMMENT '姓名（脱敏）',
    `is_first_choice` TINYINT(1) DEFAULT 1 COMMENT '1=一志愿, 0=调剂生',
    `first_choice_school_code` VARCHAR(10) DEFAULT '' COMMENT '一志愿报考学校代码（调剂生特有）',
    `first_choice_school_name` VARCHAR(50) DEFAULT '' COMMENT '一志愿报考学校名称（调剂生特有）',

    -- 初试4门科目
    `initial_politics` TINYINT UNSIGNED NOT NULL COMMENT '政治',
    `initial_english` TINYINT UNSIGNED NOT NULL COMMENT '英语',
    `initial_math` TINYINT UNSIGNED NOT NULL COMMENT '数学',
    `initial_cs_408` TINYINT UNSIGNED NOT NULL COMMENT '专业课（408或自命题）',
    `initial_total_score` INT NOT NULL COMMENT '初试总分',

    -- 复试与加权总分
    `retest_written_score` DECIMAL(5,2) DEFAULT 0.00 COMMENT '复试笔试/机试',
    `retest_interview_score` DECIMAL(5,2) DEFAULT 0.00 COMMENT '复试面试',
    `final_score` DECIMAL(5,2) NOT NULL COMMENT '综合总分',

    -- 分组与录取状态
    `retest_group` VARCHAR(30) DEFAULT '未分组' COMMENT '复试分组',
    `is_admitted` TINYINT(1) DEFAULT 0 COMMENT '0=未录取, 1=拟录取',

    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`admission_record_id`) REFERENCES `admission_records`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='复试学生明细表';

ALTER TABLE `retest_rosters` ADD INDEX `idx_record_choice` (`admission_record_id`, `is_first_choice`);
```

### 6.4 用户表 (`users`)

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

> **注意：** 启动时 GORM AutoMigrate 会自动创建/更新这四张表，无需手动执行 DDL。

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

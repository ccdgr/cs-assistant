# 📖 基于 Go Agent 的 408 考研择校助手——项目设计文档

## 1. 项目概述

本项是一个面向计算机考研（408科目）学生的智能择校与信息查询系统。项目采用 **Go 语言** 自研轻量级 **ReAct (Reasoning and Acting) 大模型 Agent 算力环**，通过 **Tool Calling（工具调用）** 驱动后端关系型数据库，解决传统大模型在“考研分数线、招生人数”等硬性数据上频繁出现的幻觉（说瞎话）问题，为用户提供精准、可追溯的择校建议。

---

## 2. 系统技术栈拓扑

* **前端（User Interface）：** 微信小程序原生开发 + Vant Weapp 组件库
* **网关与业务后端（Backend）：** Go 1.22+ / **Fiber v3**（高性能 Web 框架）
* **持久层（ORM）：** **GORM v2**
* **关系型数据库（Database）：** **MySQL 8.0**（存储结构化高校与分数线数据）
* **缓存与状态机（Cache/Session）：** **Redis 7.0**（分布式会话、大模型上下文滑动窗口、接口防刷限流）
* **AI 核心（LLM Engine）：** 大模型原生 API (如 DeepSeek-V3 / Qwen-Plus) + 手写 Tool Calling 分发器

---

## 3. 数据库 Schema 设计 (MySQL)

为了保证 Agent 能够精准查询，数据库设计需保持高度规范化。核心包含两张表：学校基本信息表、历年分数与招生表。

### 3.1 学校信息表 (`schools`)

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

-- 核心索引：便于大模型进行条件筛选
ALTER TABLE `schools` ADD INDEX `idx_region_tier` (`region`, `tier`);

```

### 3.2 历年招录数据表 (`major_scores`)

```sql
CREATE TABLE `major_scores` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `school_id` INT UNSIGNED NOT NULL COMMENT '关联schools表id',
    `year` INT NOT NULL COMMENT '年份，如2024, 2025',
    `major_code` VARCHAR(20) NOT NULL COMMENT '专业代码，如085400',
    `major_name` VARCHAR(50) NOT NULL COMMENT '专业名称，如软件工程',
    `score_line` INT NOT NULL COMMENT '复试线',
    `intake_num` INT NOT NULL COMMENT '招生人数/计划招生',
    `note` TEXT COMMENT '备注：如政治单科线、专业课改408年份说明',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`school_id`) REFERENCES `schools`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='历年招录明细表';

-- 核心索引：用于高频的点查和范围查找
ALTER TABLE `major_scores` ADD INDEX `idx_school_year_major` (`school_id`, `year`, `major_code`);

```

---

## 4. Redis 缓存与会话设计

Redis 在本项目中不单单是缓存，它承担了 **Agent 状态机内存** 的职责。

| Key 设计 | 数据结构 | TTL (过期时间) | 业务场景描述 |
| --- | --- | --- | --- |
| `session:${token}` | Hash | 7 天 | 微信小程序登录态，存储用户 `openid`、用户基本目标等。 |
| `chat:history:${session_id}` | List | 1 小时 | 存储大模型多轮对话的上下文窗口，每次新对话 `RPUSH` 写入。 |
| `rate:limit:${openid}:${api}` | String | 1 分钟 | 基于 Redis 计数器（或令牌桶）实现的防刷限流，防止 API 被恶意刷满。 |

---

## 5. Agent 核心架构与 ReAct 工作流

在 Go 后端，放弃使用 Python 的 LangChain 框架，完全基于 **Go 结构体与大模型接口规范** 实现 ReAct。

### 5.1 工具定义 (Go Struct)

通过标准 JSON 标签定义大模型能看懂的“工具定义”：

```go
// Tool 定义大模型识别的 Function 格式
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

### 5.2 ReAct 控制循环核心逻辑

在后端服务中，Agent 的核心执行环如下：

```go
func ExecuteAgentLoop(ctx context.Context, userInput string, history []Message) (string, error) {
    // 1. 组装 System Prompt (赋予其择校专家身份，并注入可用工具列表)
    // 2. 将 history + userInput 投喂给大模型
    // 3. 进入 ReAct 状态循环（限制最大循环次数 MaxLoops = 3，防止死循环）
    for i := 0; i < MaxLoops; i++ {
        resp, err := CallLLMWithTools(messages, availableTools)
        
        // 判定大模型是否需要调用工具
        if resp.FinishReason != "tool_calls" {
            // 大模型判定不需要查数据库了，直接给出了最终回复 (Thought -> Final Answer)
            return resp.Content, nil
        }
        
        // 大模型给出了 Action：决定调用某个工具
        toolCall := resp.ToolCalls[0] 
        var observation string
        
        // 动态路由分发工具调用
        switch toolCall.Function.Name {
        case "query_school_scores":
            // 解析大模型生成的参数
            args := parseArgs(toolCall.Function.Arguments)
            // 调用 GORM 查 MySQL
            observation = tools.QuerySchoolScoresFromDB(db, args) 
        case "compare_schools":
            observation = tools.CompareSchools(db, parseArgs(toolCall.Function.Arguments))
        }
        
        // 将 Action 和 Observation 作为新一轮上下文 Append 到 messages 中，继续驱动大模型思考
        messages = append(messages, Message{Role: "assistant", ToolCalls: resp.ToolCalls})
        messages = append(messages, Message{Role: "tool", Content: observation, ToolCallID: toolCall.ID})
    }
    return "", errors.New("Agent 思考环超出最大步数限制")
}

```

---

## 6. 核心接口定义 (API Specification)

项目采用现代流式交互，全面应用 **SSE (Server-Sent Events)** 协议。

### 6.1 流式对话接口

* **请求地址：** `POST /api/v1/chat/stream`
* **请求头：** `Content-Type: application/json`, `Authorization: Bearer <Token>`
* **请求体：**

```json
{
  "message": "我想考北京的211，408专业课，去年复试线330分以下的学校有哪些？"
}

```

* **响应格式 (SSE)：** `Content-Type: text/event-stream`

```text
data: {"type": "thought", "content": "用户正在寻找北京、211、408科目、分数线 <= 330 的学校。我需要调用 query_school_scores 工具。"}

data: {"type": "action", "content": "正在为您查询数据库中北京地区符合条件的高校数据..."}

data: {"type": "chunk", "content": "根据查询，为您找到以下符合条件的北京211高校：\n1. **北京交通大学** (085400)：2025年复试线 325分。\n2. **北京工业大学** (081200)：2025年复试线 318分。"}
data: [DONE]

```

---

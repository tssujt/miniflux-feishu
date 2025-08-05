# Miniflux-Feishu Integration

将 Miniflux 的新文章 webhook 拆分成多个 Entry，然后转发到飞书群聊的服务。

## 功能特性

- 接收 Miniflux 的 `new_entries` webhook 事件
- 将每个新文章拆分为单独的消息发送到飞书
- 简洁的消息结构（标题、内容、链接）
- 自动过滤 HTML 标签，提供清洁的文本内容
- 支持动态指定飞书 webhook URL（通过 URL 参数）

## 使用方法

### 1. 运行服务

```bash
# 安装依赖
go mod tidy

# 编译运行
go build -o main ./cmd/server
./main
```

或者直接运行：

```bash
go run ./cmd/server
```

### 2. 环境变量（可选）

```env
PORT=8000  # 服务端口，默认 8000
```

### 3. 服务接口

服务提供以下接口：

- `POST /webhook/miniflux?webhook_url=YOUR_FEISHU_WEBHOOK_URL` - 接收 Miniflux webhook
- `GET /health` - 健康检查

### 4. 配置 Miniflux

在 Miniflux 中配置 webhook：

1. 进入 Miniflux 设置页面
2. 在 "Webhooks" 部分添加新的 webhook
3. 设置 URL 为：`http://your-server:8000/webhook/miniflux?webhook_url=https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_WEBHOOK_KEY`


## 飞书消息格式

参考文档 [webhook 触发器](https://www.feishu.cn/hc/zh-CN/articles/807992406756-webhook-%E8%A7%A6%E5%8F%91%E5%99%A8)

每个新文章将以飞书标准的文本消息格式发送到 webhook：

### 消息结构示例

```json
{
  "msg_type": "text",
  "content": {
    "title": "[RSS 源标题] - 文章标题",
    "content": "文章内容摘要（去除 HTML 标签，限制 300 字符）",
    "url": "https://example.com/article-url"
  }
}
```

### 实际示例

```json
{
  "msg_type": "text",
  "content": {
    "title": "[技术博客] - Go 语言并发编程最佳实践",
    "content": "在 Go 语言中，goroutine 和 channel 是实现并发的核心机制。本文将详细介绍如何正确使用这些特性来构建高效的并发程序...",
    "url": "https://blog.example.com/go-concurrency-best-practices"
  }
}
```

### 字段说明

- **msg_type**: 消息类型，固定为 `"text"`
- **content**: 消息内容对象，包含：
  - **title**: 包含 RSS 源标题前缀的完整标题，格式为 `[RSS源标题] - 文章标题`
  - **content**: 文章内容摘要，自动去除 HTML 标签，超过 300 字符会被截断并添加省略号
  - **url**: 文章的原始链接 URL

## License

MIT License

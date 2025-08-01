# Miniflux-Feishu Integration

将 Miniflux 的新文章 webhook 拆分成多个 Entry，然后转发到飞书群聊的服务。

## 功能特性

- 接收 Miniflux 的 `new_entries` webhook 事件
- 将每个新文章拆分为单独的消息发送到飞书
- 富文本格式展示文章信息（标题、内容摘要、标签、发布时间）
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

每个新文章将以富文本格式发送到飞书，包含：

- 📰 RSS 源标题（加粗）
- 🔗 文章标题（链接到原文）
- 📄 内容摘要（如果有）
- 🏷️ 标签（如果有）
- 🕐 发布时间

## License

MIT License

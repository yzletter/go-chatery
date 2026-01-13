# go-chatery

一个 Go 语言实现的简易即时聊天项目，包含两个版本：
- 版本一：基于 Channel 的单机聊天室（Gin + WebSocket）
- 版本二：基于 RabbitMQ + MySQL 的可扩展聊天（WebSocket + MQ）

## 功能概览
- WebSocket 双向通信
- 版本一：Hub 广播、多用户实时聊天、首条消息作为用户名
- 版本二：RabbitMQ 解耦、消息落库、心跳保活、简单内容过滤

## 目录结构
- `im_channel`: 版本一（前端页面 + WebSocket 服务）
- `im_rabbitmq`: 版本二（HTTP 接口 + WebSocket + MQ + MySQL）
- `images`: 效果图与架构图

## 版本一：用 Channel 实现的聊天室

### 效果图
![效果图](/images/channel.png)

### 架构图
以 User1 在群里发言为例
![架构图](/images/im版本一架构图.jpg)

### 运行方式
1. 启动服务
   ```bash
   go run ./im_channel --port 5678
   ```
2. 浏览器打开 `http://localhost:5678/`
3. 页面第一条消息代表用户名，不会被广播给其他用户

注意：`im_channel/home.html` 中 WebSocket 地址写死为 `ws://127.0.0.1:5678/chat`，修改端口时请同步调整。

## 版本二：用 RabbitMQ 实现即时聊天

### 架构图
![架构图](/images/im版本二架构图.jpg)

### 运行依赖
- Go（以 `go.mod` 为准）
- RabbitMQ（默认 `localhost:5672`）
- MySQL（默认 `localhost:3306`）

### 配置
- RabbitMQ 账号密码：`im_rabbitmq/common/config.go`
- MySQL DSN：`im_rabbitmq/mysql/mysql.go`

### 初始化数据库
```bash
mysql < im_rabbitmq/mysql/create_table.sql
```

### 启动服务
```bash
go run ./im_rabbitmq
```
默认监听 `localhost:8081`。

### 接口与消息格式
- `GET /register?id=1001`: 为用户创建 exchange + queue（每个用户独立）
- `WS /speak/:id`: 建立 WebSocket 连接并收发消息
  - 说明：当前 `im_rabbitmq/main.go` 路由为 `/speak/`，如需按用户 ID 建立连接，请将路由改为 `/speak/:id` 或改用 query 参数读取 ID

发送消息示例（JSON）：
```json
{
  "message_from": 1001,
  "message_to": 1002,
  "content": "hello"
}
```
服务端会填充 `id` / `created_at` 并写入 MySQL，然后把消息分别发送给发送方与接收方。

### 消息流简述
1. WebSocket 收到客户端消息
2. 内容过滤（空内容直接丢弃）
3. 写入 MySQL
4. 发布到 RabbitMQ（`<uid>_exchange`，fanout）
5. 消费队列并写回 WebSocket

### 心跳机制
服务端每 3 秒发送一次 ping，期望 5 秒内收到 pong，超时会断开连接。

## 测试
- `go test ./im_rabbitmq/mysql -run TestInit`（需要本地 MySQL 可连通）

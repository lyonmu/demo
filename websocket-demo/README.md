# WebSocket Demo

这是一个使用 `github.com/gorilla/websocket` 和 `github.com/gin-gonic/gin` 实现的 WebSocket 演示应用。

## 功能特性

- ✅ WebSocket 服务器端实现
- ✅ 定时推送 JSON 消息（每 5 秒）
- ✅ 消息结构体封装
- ✅ 多客户端连接支持
- ✅ 美观的 Web 测试界面

## 消息结构

推送的消息使用以下结构体：

```go
type Message struct {
    ID        int       `json:"id"`
    Type      string    `json:"type"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    Data      DataInfo  `json:"data"`
}

type DataInfo struct {
    Status string  `json:"status"`
    Value  float64 `json:"value"`
    Count  int     `json:"count"`
}
```

## 运行方式

### 方式一：直接运行（不使用代理）

1. 安装依赖：

```bash
go mod tidy
```

2. 运行服务器：

```bash
go run main.go
```

3. 打开浏览器访问：

```
http://localhost:8080
```

4. 点击"连接"按钮建立 WebSocket 连接

5. 服务器会每 5 秒自动推送一条 JSON 消息

### 方式二：使用 Envoy 代理

1. 启动后端服务：

```bash
go run main.go
```

后端服务将在 `http://localhost:8080` 运行

2. 启动 Envoy 代理：

使用 Docker Compose：

```bash
docker-compose -f envoy-compose.yml up -d
```

或者直接使用 Docker：

```bash
docker run -d \
  --name envoy-proxy \
  --network host \
  -v $(pwd)/envoy.yaml:/etc/envoy/envoy.yaml \
  envoyproxy/envoy:v1.36-latest \
  -c /etc/envoy/envoy.yaml
```

3. 通过 Envoy 代理访问：

```
http://localhost:19894
```

4. 点击"连接"按钮，WebSocket 连接将通过 Envoy 代理转发到后端

**Envoy 配置说明：**

- Envoy 监听端口：`19894`
- 后端服务地址：`127.0.0.1:8080`
- WebSocket 端点：`/ws`
- Admin 管理端口：`19901`（访问 `http://localhost:19901` 查看 Envoy 管理界面）

## API 端点

- `GET /` - Web 测试页面
- `GET /ws` - WebSocket 连接端点
- `GET /health` - 健康检查端点，返回当前连接的客户端数量

## 示例消息

```json
{
  "id": 1,
  "type": "notification",
  "content": "这是一条定时推送的消息",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "status": "active",
    "value": 1.5,
    "count": 10
  }
}
```

## 技术栈

- Go 1.24+
- Gin Web Framework
- Gorilla WebSocket
- Envoy Proxy（可选）

## Envoy 代理配置

Envoy 配置文件位于 `envoy.yaml`，主要配置：

- **监听端口**：19894
- **后端集群**：127.0.0.1:8080
- **WebSocket 支持**：已启用，支持 WebSocket 升级
- **路由规则**：
  - `/ws` - WebSocket 连接端点
  - `/` - 其他 HTTP 请求

Envoy 会自动处理 WebSocket 升级请求，将连接透明地代理到后端服务。

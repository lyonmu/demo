## consul demo（Gin + cmux + WebSocket + Prometheus）

一个最小可用的示例服务，演示：
- 使用 Gin 提供 HTTP API
- 提供 WebSocket 端点并定时向客户端广播消息
- 启动时向 Consul 注册服务
- 使用 cmux 在单个 TCP 监听上复用协议（此处用于 HTTP）
- 在 `/metrics` 暴露 Prometheus 指标

### 环境要求
- Go 1.22+（推荐）
- 已运行的 Consul Agent（需要根据环境调整地址与 Token）

### 项目结构
- `main.go`：程序入口，包含 Gin、cmux、Consul 注册、WebSocket 广播与 Prometheus 指标
- `go.mod`、`go.sum`：依赖管理

### Consul 配置
根据你的环境修改 `main.go` 中以下位置：
- Consul 地址与 Token（函数 `initConsul`）
- 服务对外地址、服务 ID 与健康检查 URL（`AgentServiceRegistration`）

示例（请替换为你的实际 IP/Token）：
```go
// initConsul()
config := capi.DefaultConfig()
config.Address = "192.168.100.156:8500"
config.Token = "<YOUR_TOKEN>"

// ServiceRegister
regEnvoy := &capi.AgentServiceRegistration{
	Name:    "demo",
	Port:    8080,
	Address: "192.168.8.24",
	ID:      "192.168.8.24:8080",
	Check: &capi.AgentServiceCheck{
		HTTP:     "http://192.168.8.24:8080/demo/health",
		Interval: "10s",
		Timeout:  "5s",
		DeregisterCriticalServiceAfter: "10s",
	},
}
```

### 构建与运行
在项目根目录执行：

```bash
go mod tidy
go run main.go
```

默认监听端口为 `:8080`。

### 接口说明
- 健康检查：`GET /demo/health`
  - 示例：
    ```bash
    curl -s http://127.0.0.1:8080/demo/health
    ```

- WebSocket：`GET /demo/ws`
  - 服务器每 1 秒向所有已连接客户端推送一条 JSON 消息
  - 使用 `wscat` 测试：
    ```bash
    # 首次安装：npm i -g wscat
    wscat -c ws://127.0.0.1:8080/demo/ws
    ```

- Prometheus 指标：`GET /metrics`
  - 示例：
    ```bash
    curl -s http://127.0.0.1:8080/metrics | head
    ```

### 注意事项
- Gin 运行模式：开发模式下会有提示，生产环境建议设置：
  ```bash
  export GIN_MODE=release
  ```
- 反向代理信任：Gin 默认信任所有代理会给出警告，若在代理后运行，请显式配置可信代理：
  ```go
  // 示例：仅信任本机
  // router.SetTrustedProxies([]string{"127.0.0.1"})
  ```
- cmux 使用：我们在单独的 goroutine 中启动基于 `http.Server` 的 Gin 服务，主 goroutine 调用 `m.Serve()` 开始复用处理，避免死锁。
- Prometheus：已使用非弃用 API `collectors.NewGoCollector()` 与 `collectors.NewProcessCollector()`。

### 常见问题排查
- 启动出现死锁：确认 HTTP 服务器在 goroutine 中启动，主 goroutine 调用了 `m.Serve()`（本项目已按此实现）。
- Consul 注册失败：检查地址/Token 是否正确，健康检查 URL 是否能被 Consul Agent 访问。
- 端口被占用：修改 `main.go` 中监听端口（默认 `8080`）。

### 许可证
MIT


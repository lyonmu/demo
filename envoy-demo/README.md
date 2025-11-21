# Envoy Proxy Demos

这是一个 Envoy Proxy 的学习和演示项目集合，包含各种场景下的 Envoy 配置示例。

## 📂 示例列表

### [L4-L7 Proxy Demo](./L4-L7-porxy-demo/)

演示 Envoy 作为 L4/L7 代理的高级功能：

- **HTTPS 终止**：使用自签名证书。
- **多端口监听**：同时监听多个端口并转发到不同的上游服务。
- **静态集群**：配置静态 IP 的上游集群。
- **Admin 接口**：启用 Envoy 管理界面。
- **访问日志**：配置标准输出日志。

## 📚 学习资源

- [Envoy 官方文档](https://www.envoyproxy.io/docs/envoy/latest/)
- [Envoy GitHub 仓库](https://github.com/envoyproxy/envoy)

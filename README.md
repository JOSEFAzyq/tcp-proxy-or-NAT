# TCP PROXY 内网穿透程序
## 依赖
- `gin`,路由服务,供外网访问.
- `github.com/gorilla/websocket`,升级成`websocket`服务
- 其他均为原生组件
## 使用方法
前置环境:客户端能单向访问到服务端ip地址以及端口.
### 修改 config_template.yaml 文件为 config.yaml,其中配置改为自己的
```
server:
  host: 0.0.0.0 # 穿透服务端ip地址,客户端可以单向访问到这个地址
  port: 8083 # 穿透服务端端口号地址,客户端可以单向访问到这个端口
  proxy: https://yourdomain.com # 你想要访问的内网地址,这里可以改成ip.
```
### 先在服务端启动 `go run main.go -run server`
能看到一堆调试信息以及`启动socket以及tcp服务`表示启动成功

### 再在客户端(被代理端)启动`go run main.go -run client`
看到`连接 server 成功`表示启动成功

### 在域名上访问`穿透服务域名/proxy/{*path}`则可以访问到被代理的客户端了
比如: `穿透服务域名/proxy/api/xxx` 就能访问到 server.proxy/api/xxx
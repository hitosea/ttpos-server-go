# 调用链跟踪

## 准备工作
1. 使用 docker-compose.mid.yaml 启动中间件，组件已包含jaeger all in one
2. jaeger ui  http://ip:16686,  如果使用了easytier 组网到coder 可以使用  http://easytier-ip:16686
3. 接入端口默认 4318
   - 中台模块通过增加config.yaml 
        ```yaml
        # 调用链监控配置
        otlp:
            enabled: true
            serviceName: "ttpos-erp"
            endpoint: "10.144.144.11:4318"
            path: "/v1/traces"
        ```
   - main 模块在 .env 修改配置

    ```
    # OTLP 配置
    OTLP_ENABLED=true
    OTLP_SERVICE_NAME=ttpos-main
    OTLP_ENDPOINT=172.17.0.9:4318
    OTLP_PATH=/v1/traces
    OTLP_LOG_HEADERS=device-id,client-version,x-ttpos-company-id
    # END ------ OTLP 配置 ------
    ```
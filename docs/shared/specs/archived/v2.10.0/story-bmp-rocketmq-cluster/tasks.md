# 任务分解: BMP RocketMQ 集群配置支持

> 对应设计: [story-bmp-rocketmq-cluster](./design.md)

---

## 1. 📅 进度概览

- **总预估工时**: 4 小时
- **负责人**: rikugun
- **状态**: 待开始

| ID | 任务名称 | 优先级 | 状态 | 负责人 | 预估工时 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| T-01 | 修改配置生成脚本 `init_conf.sh` | P0 | 待开始 | rikugun | 1h |
| T-02 | 更新所有 BMP 模块配置文件模板 | P0 | 待开始 | rikugun | 1h |
| T-03 | 本地验证单节点与多节点配置生成 | P0 | 待开始 | rikugun | 1h |
| T-04 | 集成测试与服务启动验证 | P1 | 待开始 | rikugun | 1h |

---

## 2. 📝 详细任务列表

### T-01: 修改配置生成脚本 `init_conf.sh`
- [ ] 修改 `ttpos-bmp/hack/init_conf.sh`
- [ ] 增加 `ROCKETMQ_NAME_SRV_ADDR` 的处理逻辑
- [ ] 实现逗号/分号分隔的字符串解析
- [ ] 格式化为 YAML 列表字符串并 Export

### T-02: 更新所有 BMP 模块配置文件模板
- [ ] 修改 `ttpos-bmp/app/ttpos-message/manifest/config/config.tpl.yaml`
- [ ] 修改 `ttpos-bmp/app/ttpos-takeout/manifest/config/config.tpl.yaml`
- [ ] 修改 `ttpos-bmp/app/ttpos-erp/manifest/config/config.tpl.yaml`
- [ ] 修改 `ttpos-bmp/app/ttpos-manager/manifest/config/config.tpl.yaml`
- [ ] 修改 `ttpos-bmp/app/ttpos-shop/manifest/config/config.tpl.yaml`
- [ ] 修改 `ttpos-bmp/app/ttpos-websocket/manifest/config/config.tpl.yaml`
- [ ] 确认移除 `nameSrvAdders` 值的方括号内的引号

### T-03: 本地验证
- [ ] 创建测试脚本或手动运行验证 `init_conf.sh` 输出
- [ ] 检查生成的 `config.yaml` 格式是否正确

### T-04: 集成测试
- [ ] 启动 `ttpos-message` 服务
- [ ] 验证日志中 RocketMQ 连接信息


include hack/hack-cli.mk
include hack/hack-ttpos.mk

# 显示帮助信息
.PHONY: help
help:
	@echo "TTPOS 业务中台模块 - 可用命令列表"
	@echo ""
	@echo "📦 环境配置:"
	@echo "  conf              生成配置文件（依赖上级目录的 .env 文件）"
	@echo "  cli               安装/更新最新的 GoFrame CLI 工具"
	@echo "  cli.install       检查并安装 CLI 工具"
	@echo "  envsubst.install  安装 envsubst 工具"
	@echo "  migrate.install   安装 migrate 数据库迁移工具"
	@echo ""
	@echo "🚀 容器运行:"
	@echo "  mid               在容器中启动中间件（Nacos、RocketMQ 等）"
	@echo "  up                在容器中构建并启动全部中台服务"
	@echo "  run               启动中间件及中台应用服务"
	@echo ""
	@echo "🛠️  本地开发:"
	@echo "  run.manager       本地单独启动管理模块 (14001:http 14002:grpc)"
	@echo "  run.shop          本地单独启动门店模块 (14011:http 14012:grpc)"
	@echo "  run.erp           本地单独启动ERP模块 (14021:http 14022:grpc)"
	@echo "  run.takeout       本地单独启动外送模块 (14031:http 14032:grpc)"
	@echo ""
	@echo "🗄️  数据库:"
	@echo "  migrate           迁移升级所有应用的数据库"
	@echo ""
	@echo "💡 使用示例:"
	@echo "  make conf         # 首次运行或 .env 变更后执行"
	@echo "  make mid          # 启动中间件"
	@echo "  make run.manager  # 本地开发管理模块"
	@echo ""
	@echo "📚 更多信息请参考 README.MD"

# 默认显示帮助信息
.DEFAULT_GOAL := help



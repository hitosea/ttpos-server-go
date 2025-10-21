#!/bin/bash

# TTPOS 消息中心服务开发脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        error "$1 命令未找到，请先安装"
        exit 1
    fi
}

# 生成代码
gen_code() {
    info "开始生成代码..."
    
    # 生成 Protobuf 代码
    info "生成 Protobuf 代码..."
    cd ../.. && gf gen pb -p app/ttpos-message
    cd app/ttpos-message
    
    # 生成 DAO 代码（需要数据库连接）
    if [ "$1" == "--with-dao" ]; then
        info "生成 DAO 代码..."
        gf gen dao || warn "DAO 生成失败，可能需要配置数据库连接"
    fi
    
    # 生成 Service 接口代码
    info "生成 Service 接口代码..."
    gf gen service || warn "Service 生成失败"
    
    info "代码生成完成"
}

# 初始化数据库
init_db() {
    info "初始化数据库..."
    
    if [ -z "$DB_HOST" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASS" ]; then
        error "请设置数据库环境变量: DB_HOST, DB_USER, DB_PASS"
        exit 1
    fi
    
    DB_NAME="messages"
    
    # 创建数据库
    mysql -h$DB_HOST -u$DB_USER -p$DB_PASS -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"
    
    # 执行迁移脚本
    for sql_file in manifest/sql/*_up.sql; do
        info "执行迁移脚本: $sql_file"
        mysql -h$DB_HOST -u$DB_USER -p$DB_PASS $DB_NAME < $sql_file
    done
    
    info "数据库初始化完成"
}

# 启动服务
run_service() {
    info "启动服务..."
    
    # 检查配置文件
    if [ ! -f "manifest/config/config.yaml" ]; then
        warn "配置文件不存在，复制模板..."
        cp manifest/config/config.tpl.yaml manifest/config/config.yaml
        warn "请编辑 manifest/config/config.yaml 配置文件"
        exit 1
    fi
    
    # 运行服务
    go run main.go
}

# 构建服务
build_service() {
    info "构建服务..."
    
    mkdir -p bin
    go build -o bin/ttpos-message main.go
    
    info "构建完成: bin/ttpos-message"
}

# 清理
clean() {
    info "清理编译产物..."
    
    rm -rf bin/
    rm -rf output/
    rm -rf log/
    
    info "清理完成"
}

# 帮助信息
show_help() {
    echo "TTPOS 消息中心服务开发脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  gen           - 生成代码（Protobuf + Service）"
    echo "  gen-all       - 生成所有代码（包括 DAO）"
    echo "  init-db       - 初始化数据库"
    echo "  run           - 运行服务"
    echo "  build         - 构建服务"
    echo "  clean         - 清理编译产物"
    echo "  help          - 显示帮助信息"
    echo ""
}

# 主函数
main() {
    case "$1" in
        gen)
            gen_code
            ;;
        gen-all)
            gen_code --with-dao
            ;;
        init-db)
            init_db
            ;;
        run)
            run_service
            ;;
        build)
            build_service
            ;;
        clean)
            clean
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"


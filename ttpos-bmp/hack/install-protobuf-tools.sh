#!/bin/bash
# TTPOS BMP - Protobuf 工具链安装脚本
# 固定版本: protoc v3.21.12, protoc-gen-go-grpc v1.6.0

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 版本配置
# 注意: protoc 的 GitHub tag 是 v21.12, 但 libprotoc 版本显示为 3.21.12
PROTOC_TAG="21.12"           # GitHub release tag
PROTOC_VERSION="3.21.12"     # libprotoc --version 输出的版本号
PROTOC_GEN_GO_VERSION="v1.36.11"
PROTOC_GEN_GO_GRPC_VERSION="v1.6.0"

# 检测系统架构
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="aarch_64" ;;
    *) echo -e "${RED}不支持的架构: $ARCH${NC}"; exit 1 ;;
esac

echo -e "${GREEN}=== TTPOS BMP Protobuf 工具链安装 ===${NC}"
echo -e "${YELLOW}目标版本:${NC}"
echo -e "  protoc:               ${PROTOC_VERSION}"
echo -e "  protoc-gen-go:        ${PROTOC_GEN_GO_VERSION}"
echo -e "  protoc-gen-go-grpc:   ${PROTOC_GEN_GO_GRPC_VERSION}"
echo ""

# 安装 protoc
echo -e "${GREEN}[1/3] 安装 protoc ${PROTOC_VERSION}${NC}"
PROTOC_ZIP="protoc-${PROTOC_TAG}-${OS}-${ARCH}.zip"
PROTOC_URL="https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_TAG}/${PROTOC_ZIP}"

if command -v protoc &> /dev/null; then
    CURRENT_VERSION=$(protoc --version | awk '{print $2}')
    if [ "$CURRENT_VERSION" = "$PROTOC_VERSION" ]; then
        echo -e "${YELLOW}  protoc $PROTOC_VERSION 已安装，跳过${NC}"
    else
        echo -e "${YELLOW}  当前版本: $CURRENT_VERSION, 需要升级到 $PROTOC_VERSION${NC}"
        INSTALL_PROTOC=true
    fi
else
    echo -e "${YELLOW}  protoc 未安装${NC}"
    INSTALL_PROTOC=true
fi

if [ "$INSTALL_PROTOC" = true ]; then
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    echo -e "${YELLOW}  下载 ${PROTOC_URL}${NC}"
    if ! curl -fsSL -o "$PROTOC_ZIP" "$PROTOC_URL"; then
        echo -e "${RED}  ✗ 下载失败，请检查网络连接或手动下载${NC}"
        echo -e "${YELLOW}  手动下载地址: ${PROTOC_URL}${NC}"
        cd - > /dev/null 2>&1
        rm -rf "$TMP_DIR"
        exit 1
    fi

    echo -e "${YELLOW}  解压并安装到 ~/.local${NC}"
    if ! unzip -q "$PROTOC_ZIP"; then
        echo -e "${RED}  ✗ 解压失败${NC}"
        cd - > /dev/null 2>&1
        rm -rf "$TMP_DIR"
        exit 1
    fi

    mkdir -p ~/.local/bin ~/.local/include
    # 先删除旧的 protoc（如果存在）
    rm -f ~/.local/bin/protoc
    cp -r bin/* ~/.local/bin/
    cp -r include/* ~/.local/include/

    cd - > /dev/null 2>&1
    rm -rf "$TMP_DIR"
    echo -e "${GREEN}  ✓ protoc $PROTOC_VERSION 安装完成${NC}"
fi

# 安装 protoc-gen-go
echo -e "${GREEN}[2/3] 安装 protoc-gen-go ${PROTOC_GEN_GO_VERSION}${NC}"
go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}
echo -e "${GREEN}  ✓ protoc-gen-go ${PROTOC_GEN_GO_VERSION} 安装完成${NC}"

# 安装 protoc-gen-go-grpc
echo -e "${GREEN}[3/3] 安装 protoc-gen-go-grpc ${PROTOC_GEN_GO_GRPC_VERSION}${NC}"
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}
echo -e "${GREEN}  ✓ protoc-gen-go-grpc ${PROTOC_GEN_GO_GRPC_VERSION} 安装完成${NC}"

# 更新 PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# 验证安装
echo ""
echo -e "${GREEN}=== 验证安装 ===${NC}"
echo -e "${YELLOW}protoc 版本:${NC}"
protoc --version

echo -e "${YELLOW}protoc-gen-go 版本:${NC}"
protoc-gen-go --version

echo -e "${YELLOW}protoc-gen-go-grpc 版本:${NC}"
protoc-gen-go-grpc --version

echo ""
echo -e "${GREEN}✓ Protobuf 工具链安装完成${NC}"
echo -e "${YELLOW}提示: 请确保 \$GOPATH/bin 在你的 PATH 中${NC}"
echo -e "      export PATH=\"\$PATH:\$(go env GOPATH)/bin\""

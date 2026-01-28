#!/bin/bash

# =====================================================
# 生成多门店库存查询 SQL 脚本（横向报表格式）
# =====================================================
# 使用方法：修改下面的配置，然后运行脚本
# 输出格式：门店名称 | SA01001 | SA01002 | ...
#          每列值：默认库存 + 在途库存
# =====================================================

# ========== 配置区域（修改这里）==========

# 门店 ID 列表（company_uuid，用逗号分隔，不带 'shop' 前缀）
# 示例：SHOP_IDS="8267304538112000,9876543210000000,1234567890123456"
SHOP_IDS="5752387276800000,5522988212224000,5005268492288000,3087884357632000,4197894328320000,4053593493504000,3847871270912000,3662462062592000,3367514411008000,3169446793216000,2992442970112000,2788834676736000,2618629820416000,2445149216768000,2277263806464000,2101535051776000,1919875551232000,1745354756096000,1559157018624000,1379607252992000,1167111229440000,9231705128960000,7264576512000000,5498438983680000,3631470387200000,1515821506560000,8723170856960000,8535580610560000,8051063001088000,7813128523776000,7600687026176000,5444022046720000,5250979205120000,5001267122176000,4805464432640000,4613872816128000,4418766376960000,4229506797568000,4024875094016000,3782477877248000,3448951017472000,2947521978368000,8722592047104000,8501761941504000,8100551598080000,7863653113856000,7648065888256000,7400123801600000,7191251656704000,6977459593216000,6789240201216000,6542950670336000,5999171739648000,5567347171328000,4912616316928000,4358842359808000,4149605310464000,3870122057728000,3446618988544000,2876210421760000,2598648160256000,2269470793728000,1958987436032000,7496072699904000,7222025265152000,6773247320064000,6517281529856000,6041098002432000,5098579173376000"

# erpcode 列表（用逗号分隔）
# 示例：ERPCODE_LIST="ITEM001,ITEM002,ITEM003"
ERPCODE_LIST="SA01001,SA01002,SA01003,SA01004"

# 输出文件名
OUTPUT_FILE="ttpos-scripts/db_generator/output/warehouse_stock_query.sql"

# ========== 配置区域结束 ==========

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_header() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# 检查配置
if [ -z "$SHOP_IDS" ]; then
    echo "错误：SHOP_IDS 不能为空，请修改脚本配置"
    exit 1
fi

if [ -z "$ERPCODE_LIST" ]; then
    echo "错误：ERPCODE_LIST 不能为空，请修改脚本配置"
    exit 1
fi

# 转换 erpcode 列表为数组
ERPCODE_ARRAY=$(echo "$ERPCODE_LIST" | tr ',' ' ')

# 转换 erpcode 列表为 SQL IN 格式
ERPCODE_IN=$(echo "$ERPCODE_LIST" | sed 's/,/", "/g')
ERPCODE_IN="\"$ERPCODE_IN\""

# 创建输出文件头部
cat > "$OUTPUT_FILE" <<EOF
-- =====================================================
-- 自动生成的多门店库存查询 SQL（横向报表格式）
-- =====================================================
-- 生成时间: $(date '+%Y-%m-%d %H:%M:%S')
-- 查询 erpcode: $ERPCODE_LIST
-- 输出格式：门店名称 | erpcode1 | erpcode2 | ...
--          每列值：默认库存 + 在途库存
-- =====================================================

EOF

# 计数器
COUNT=0
TOTAL=$(echo "$SHOP_IDS" | tr ',' '\n' | wc -l)

# 遍历每个门店 ID
for SHOP_ID in $(echo "$SHOP_IDS" | tr ',' ' '); do
    COUNT=$((COUNT + 1))
    SHOP_DB="shop${SHOP_ID}"

    # 如果不是第一个，添加 UNION ALL
    if [ $COUNT -gt 1 ]; then
        echo "" >> "$OUTPUT_FILE"
        echo "UNION ALL" >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
    fi

    # 生成每个门店的查询 SQL
    cat >> "$OUTPUT_FILE" <<EOF
-- 门店 $COUNT/$TOTAL: $SHOP_DB
SELECT
    tc.name AS 门店名称,
EOF

    # 为每个 erpcode 生成一列（默认 + 在途）
    ERPCODE_COUNT=$(echo "$ERPCODE_ARRAY" | wc -w)
    ERPCODE_IDX=0
    for ERPCODE in $ERPCODE_ARRAY; do
        ERPCODE_IDX=$((ERPCODE_IDX + 1))

        cat >> "$OUTPUT_FILE" <<EOF
    (
        IFNULL(MAX(CASE WHEN tm.code = "$ERPCODE" AND w.is_default = 1 THEN wi.stock END), 0) +
        IFNULL(MAX(CASE WHEN tm.code = "$ERPCODE" AND w.type = "transit" THEN wi.stock END), 0)
    ) AS \`$ERPCODE\`
EOF

        # 如果不是最后一个 erpcode，添加逗号
        if [ $ERPCODE_IDX -lt $ERPCODE_COUNT ]; then
            echo "," >> "$OUTPUT_FILE"
        fi
    done

    # 添加 FROM 和 JOIN 子句
    cat >> "$OUTPUT_FILE" <<EOF

FROM saas.ttpos_company tc
LEFT JOIN $SHOP_DB.ttpos_material tm
    ON tm.delete_time = 0
    AND tm.code IN ($ERPCODE_IN)
LEFT JOIN $SHOP_DB.ttpos_warehouse_item wi
    ON wi.material_code = tm.code
    AND wi.delete_time = 0
LEFT JOIN $SHOP_DB.ttpos_warehouse w
    ON w.uuid = wi.warehouse_uuid
    AND w.delete_time = 0
    AND (w.is_default = 1 OR w.type = "transit")
WHERE tc.uuid = $SHOP_ID
    AND tc.delete_time = 0
GROUP BY tc.name
EOF

done

# 添加排序
cat >> "$OUTPUT_FILE" <<'EOF'

ORDER BY 门店名称;
EOF

# 输出提示信息
echo ""
print_header "生成完成"
echo "查询门店数: $TOTAL"
echo "查询 erpcode: $ERPCODE_LIST"
echo "输出文件: $OUTPUT_FILE"
echo ""
echo "输出格式示例："
echo "  门店名称 | SA01001  | SA01002  | SA01003  | SA01004"
echo "  ---------|----------|----------|----------|----------"
echo "  北京店   | 200      | 230      | 120      | 90"
echo "  上海店   | 160      | 205      | 105      | 75"
echo ""
echo "  (库存值 = 默认库存 + 在途库存)"
echo ""
print_success "SQL 已生成成功！"
echo ""
echo -e "${YELLOW}执行方式：${NC}"
echo "  1. 查看生成的 SQL："
echo "     cat $OUTPUT_FILE"
echo ""
echo "  2. 直接执行查询："
echo "     mysql -u用户名 -p < $OUTPUT_FILE"
echo ""
echo "  3. 导出为 CSV："
echo "     mysql -u用户名 -p < $OUTPUT_FILE > result.csv"
echo ""
echo "  4. 格式化输出（表格形式）："
echo "     mysql -u用户名 -p -t < $OUTPUT_FILE"
echo ""

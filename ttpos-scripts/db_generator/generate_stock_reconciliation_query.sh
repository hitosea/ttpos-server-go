#!/bin/bash

# =====================================================
# 生成多门店盘点单查询 SQL 脚本
# =====================================================
# 使用方法：修改下面的配置，然后运行脚本
# 输出格式：门店名称 | 盘点单号 | 状态 | 物品code | 物品名称 | 账面数量 | 实盘数量 | 差异 | 创建时间 | 更新时间 | 删除时间
# =====================================================

# ========== 配置区域（修改这里）==========

# 门店 ID 列表（company_uuid，用逗号分隔，不带 'shop' 前缀）
SHOP_IDS="5752387276800000,5522988212224000,5005268492288000,3087884357632000,4197894328320000,4053593493504000,3847871270912000,3662462062592000,3367514411008000,3169446793216000,2992442970112000,2788834676736000,2618629820416000,2445149216768000,2277263806464000,2101535051776000,1919875551232000,1745354756096000,1559157018624000,1379607252992000,1167111229440000,9231705128960000,7264576512000000,5498438983680000,3631470387200000,1515821506560000,8723170856960000,8535580610560000,8051063001088000,7813128523776000,7600687026176000,5444022046720000,5250979205120000,5001267122176000,4805464432640000,4613872816128000,4418766376960000,4229506797568000,4024875094016000,3782477877248000,3448951017472000,2947521978368000,8722592047104000,8501761941504000,8100551598080000,7863653113856000,7648065888256000,7400123801600000,7191251656704000,6977459593216000,6789240201216000,6542950670336000,5999171739648000,5567347171328000,4912616316928000,4358842359808000,4149605310464000,3870122057728000,3446618988544000,2876210421760000,2598648160256000,2269470793728000,1958987436032000,7496072699904000,7222025265152000,6773247320064000,6517281529856000,6041098002432000,5098579173376000"

# 查询时间范围（格式：YYYY-MM-DD）
START_DATE="2026-01-25"
END_DATE="2026-01-27"

# 输出文件名
OUTPUT_FILE="ttpos-scripts/db_generator/output/stock_reconciliation_query_25_27.sql"

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

# 计算时间戳（开始日期 00:00:00，结束日期 23:59:59）
# 使用 UTC+7 时区（泰国时间）
# 本地时间 00:00:00 UTC+7 = UTC 前一天 17:00:00
# 本地时间 23:59:59 UTC+7 = UTC 当天 16:59:59
START_UTC_DATE=$(date -d "$START_DATE -1 day" +%Y-%m-%d)
END_UTC_DATE=$(date -d "$END_DATE" +%Y-%m-%d)
START_TIMESTAMP=$(date -u -d "$START_UTC_DATE 17:00:00" +%s)
END_TIMESTAMP=$(date -u -d "$END_UTC_DATE 16:59:59" +%s)

echo ""
print_header "生成盘点单查询 SQL"
echo "查询时间范围: $START_DATE 00:00:00 ~ $END_DATE 23:59:59"
echo "时间戳范围: $START_TIMESTAMP ~ $END_TIMESTAMP"
echo ""

# 创建输出文件头部
cat > "$OUTPUT_FILE" <<EOF
-- =====================================================
-- 自动生成的多门店盘点单查询 SQL
-- =====================================================
-- 生成时间: $(date '+%Y-%m-%d %H:%M:%S')
-- 查询时间范围: $START_DATE 00:00:00 ~ $END_DATE 23:59:59
-- 时间戳范围: $START_TIMESTAMP ~ $END_TIMESTAMP
-- 输出格式：门店名称 | 盘点单号 | 状态 | 物品code | 物品名称 | 账面数量 | 实盘数量 | 差异 | 创建时间 | 更新时间 | 删除时间
-- 状态说明：0-已保存 1-已提交 2-已审核 3-已驳回
-- 时区：UTC+7（泰国时间）
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
    sr.order_no AS 盘点单号,
    CASE sr.status
        WHEN 0 THEN '已保存'
        WHEN 1 THEN '已提交'
        WHEN 2 THEN '已审核'
        WHEN 3 THEN '已驳回'
        ELSE '未知'
    END AS 状态,
    tm.code AS 物品code,
    sri.material_name AS 物品名称,
    sri.booked_quantity AS 账面数量,
    sri.counted_quantity AS 实盘数量,
    ROUND(sri.counted_quantity - sri.booked_quantity, 4) AS 差异,
    DATE_FORMAT(CONVERT_TZ(FROM_UNIXTIME(sr.create_time), '+00:00', '+07:00'), '%Y-%m-%d %H:%i:%s') AS 创建时间,
    DATE_FORMAT(CONVERT_TZ(FROM_UNIXTIME(sr.update_time), '+00:00', '+07:00'), '%Y-%m-%d %H:%i:%s') AS 更新时间,
    CASE WHEN sr.delete_time > 0
        THEN DATE_FORMAT(CONVERT_TZ(FROM_UNIXTIME(sr.delete_time), '+00:00', '+07:00'), '%Y-%m-%d %H:%i:%s')
        ELSE '-'
    END AS 删除时间
FROM saas.ttpos_company tc
LEFT JOIN $SHOP_DB.ttpos_stock_reconciliation sr
    ON sr.delete_time = 0
    AND sr.create_time >= $START_TIMESTAMP
    AND sr.create_time <= $END_TIMESTAMP
LEFT JOIN $SHOP_DB.ttpos_stock_reconciliation_item sri
    ON sri.stock_reconciliation_uuid = sr.uuid
    AND sri.delete_time = 0
LEFT JOIN $SHOP_DB.ttpos_material tm
    ON tm.uuid = sri.material_uuid
    AND tm.delete_time = 0
WHERE tc.uuid = $SHOP_ID
    AND tc.delete_time = 0
    AND sr.order_no IS NOT NULL
EOF

done

# 添加排序
cat >> "$OUTPUT_FILE" <<'EOF'

ORDER BY 门店名称, 创建时间;
EOF

# 输出提示信息
echo ""
print_header "生成完成"
echo "查询门店数: $TOTAL"
echo "查询时间: $START_DATE ~ $END_DATE"
echo "输出文件: $OUTPUT_FILE"
echo ""
echo "输出格式示例："
echo "  门店名称 | 盘点单号      | 状态   | 物品code  | 物品名称 | 账面数量 | 实盘数量 | 差异  | 创建时间            | ..."
echo "  ---------|---------------|--------|-----------|----------|----------|----------|-------|---------------------|-----"
echo "  北京店   | SR20260124001 | 已提交 | SA01001   | 可乐     | 100.0000 | 98.0000  | -2.00 | 2026-01-24 10:30:00 | ..."
echo "  北京店   | SR20260124001 | 已提交 | SA01002   | 雪碧     | 50.0000  | 52.0000  | 2.00  | 2026-01-24 10:30:00 | ..."
echo "  上海店   | SR20260125002 | 已审核 | SA01001   | 可乐     | 80.0000  | 80.0000  | 0.00  | 2026-01-25 14:15:00 | ..."
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
echo "     mysql -u用户名 -p -N < $OUTPUT_FILE > result.csv"
echo ""
echo "  4. 格式化输出（表格形式）："
echo "     mysql -u用户名 -p -t < $OUTPUT_FILE"
echo ""

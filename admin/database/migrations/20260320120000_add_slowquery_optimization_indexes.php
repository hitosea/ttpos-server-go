<?php
/**
 * 慢查询优化索引补充
 *
 * 问题来源：生产环境 2026-03-20 慢查询分析
 *
 * 1. 商品销售统计报表查询 (91秒 → 预期 <5秒)
 *    SQL: SELECT ... FROM ttpos_statistics_product sp
 *         LEFT JOIN ttpos_product_package pp ON sp.product_package_uuid = pp.uuid
 *         LEFT JOIN ttpos_product_category pc ON pp.category_uuid = pc.uuid
 *         LEFT JOIN ttpos_product_category ppc ON pc.parent_uuid = ppc.uuid
 *         GROUP BY sp.product_package_uuid
 *    问题: 全表扫描 3,646,659 行，多表JOIN无索引
 *    代码: main/app/repository/statistics.go:1979-2095
 *
 * 2. 仓库出库单查询 (10秒 → 预期 <1秒)
 *    SQL: SELECT * FROM ttpos_warehouse_out_form_item
 *         WHERE takeout_order_uuid = XXX AND reduce_stock = 0 AND delete_time = 0
 *    问题: 扫描 428,748 行，只有单列索引 idx_takeout_order_uuid，复合索引更优
 *
 * 3. 厨房生产单查询 (3-4秒 → 预期 <1秒)
 *    SQL: SELECT * FROM ttpos_production_order_product
 *         WHERE status = X AND product_package_uuid IN (...)
 *    问题: 缺少 (status, product_package_uuid) 复合索引
 *    代码: main/app/repository/production_order.go:298-312
 */

use think\migration\Migrator;

class AddSlowqueryOptimizationIndexes extends Migrator
{
    /**
     * 应用到所有商户数据库和 saas 主库
     * 这些表都在商户库中，统计报表查询频繁
     */
    const TARGET = 'all';

    public function change()
    {
        // ==================== 1. 商品销售统计表 ====================
        // 优化：商品销售统计报表查询 (91秒 → <5秒)
        if ($this->hasTable('statistics_product')) {
            $table = $this->table('statistics_product');

            // 索引1：product_package_uuid - JOIN 主键
            if (!$table->hasIndexByName('idx_product_package_uuid')) {
                $table->addIndex('product_package_uuid', [
                    'name' => 'idx_product_package_uuid',
                ]);
            }

            // 索引2：create_time - 时间范围过滤
            if (!$table->hasIndexByName('idx_create_time')) {
                $table->addIndex('create_time', [
                    'name' => 'idx_create_time',
                ]);
            }

            // 索引3：复合索引 - 覆盖 GROUP BY 和时间过滤
            if (!$table->hasIndexByName("idx_package_time")) {
                $table->addIndex(['product_package_uuid', 'create_time'], [
                    'name' => 'idx_package_time',
                ]);
            }

            $table->update();
        }

        // ==================== 2. 仓库出库单明细表 ====================
        // 优化：外卖订单仓库查询 (10秒 → <1秒)
        if ($this->hasTable('warehouse_out_form_item')) {
            $table = $this->table('warehouse_out_form_item');

            // 索引1：外卖订单复合索引（替代现有单列索引）
            // 注意：已有 idx_takeout_order_uuid，此复合索引包含更多过滤条件
            if (!$table->hasIndexByName('idx_takeout_reduce_delete')) {
                $table->addIndex(['takeout_order_uuid', 'reduce_stock', 'delete_time'], [
                    'name' => 'idx_takeout_reduce_delete',
                ]);
            }

            // 索引2：销售账单复合索引（同样的查询模式）
            if (!$table->hasIndexByName('idx_salebill_reduce_delete')) {
                $table->addIndex(['sale_bill_uuid', 'reduce_stock', 'delete_time'], [
                    'name' => 'idx_salebill_reduce_delete',
                ]);
            }

            $table->update();
        }

        // ==================== 3. 生产订单商品表 ====================
        // 优化：厨房打印机生产单查询 (3-4秒 → <1秒)
        if ($this->hasTable('production_order_product')) {
            $table = $this->table('production_order_product');

            // 索引：(status, product_package_uuid) 复合索引
            // 注意：已有 idx_status_delete_finished 用于 KDS 厨显，此索引用于打印机查询
            if (!$table->hasIndexByName('idx_status_package')) {
                $table->addIndex(['status', 'product_package_uuid'], [
                    'name' => 'idx_status_package',
                ]);
            }

            // 索引：(status, sale_bill_uuid) 复合索引 - 销售账单查询
            if (!$table->hasIndexByName('idx_status_salebill')) {
                $table->addIndex(['status', 'sale_bill_uuid'], [
                    'name' => 'idx_status_salebill',
                ]);
            }

            // 索引：(status, takeout_order_uuid) 复合索引 - 外卖订单查询
            if (!$table->hasIndexByName('idx_status_takeout')) {
                $table->addIndex(['status', 'takeout_order_uuid'], [
                    'name' => 'idx_status_takeout',
                ]);
            }

            $table->update();
        }

        // ==================== 4. 商品包装表 ====================
        // 优化：统计报表 JOIN 查询
        if ($this->hasTable('product_package')) {
            $table = $this->table('product_package');

            // 索引1：category_uuid - JOIN 外键
            if (!$table->hasIndexByName('idx_category_uuid')) {
                $table->addIndex('category_uuid', [
                    'name' => 'idx_category_uuid',
                ]);
            }

            // 索引2：(is_show_kitchen, delete_time) - 厨显过滤
            if (!$table->hasIndexByName('idx_show_kitchen_delete')) {
                $table->addIndex(['is_show_kitchen', 'delete_time'], [
                    'name' => 'idx_show_kitchen_delete',
                ]);
            }

            $table->update();
        }

        // ==================== 5. 商品分类表 ====================
        // 优化：统计报表父子分类 JOIN
        if ($this->hasTable('product_category')) {
            $table = $this->table('product_category');

            // 索引：parent_uuid - 父子分类 JOIN
            if (!$table->hasIndexByName('idx_parent_uuid')) {
                $table->addIndex('parent_uuid', [
                    'name' => 'idx_parent_uuid',
                ]);
            }

            $table->update();
        }

        // ==================== 6. 打印机商品关联表 ====================
        // 优化：厨房打印机商品查询
        if ($this->hasTable('product_printer_product_item')) {
            $table = $this->table('product_printer_product_item');

            // 索引：(product_printer_uuid, product_package_uuid) 复合索引
            if (!$table->hasIndexByName('idx_printer_package')) {
                $table->addIndex(['product_printer_uuid', 'product_package_uuid'], [
                    'name' => 'idx_printer_package',
                ]);
            }

            $table->update();
        }
    }
}

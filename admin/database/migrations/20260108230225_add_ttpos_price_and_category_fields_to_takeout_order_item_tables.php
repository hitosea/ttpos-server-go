<?php
declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

/**
 * 为外卖订单商品表和修饰符表添加店内价格和分类字段
 *
 * 添加字段:
 * - ttpos_price: 店内价格(外卖平台价格存储在原 price 字段)
 * - ttpos_category_uuid: TTPOS分类UUID
 * - ttpos_category_name: TTPOS分类名称
 * - ttpos_parent_category_uuid: TTPOS父分类UUID
 * - ttpos_parent_category_name: TTPOS父分类名称
 */
final class AddTtposPriceAndCategoryFieldsToTakeoutOrderItemTables extends AbstractMigration
{
    public function change(): void
    {
        // 1. 为 ttpos_takeout_order_item 表添加字段
        if ($this->hasTable('takeout_order_item')) {
            $table = $this->table('takeout_order_item');
            // 检查字段是否已存在
            if (!$table->hasColumn('ttpos_price')) {
                $table->addColumn('ttpos_price', 'decimal', ['precision' => 20, 'scale' => 4, 'default' => '0.0000', 'comment' => 'TTPOS店内价格(元,4位小数)', 'after' => 'price']);
            }
            if (!$table->hasColumn('ttpos_category_uuid')) {
                $table->addColumn('ttpos_category_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'TTPOS分类UUID(关联ttpos_product_category.uuid)', 'after' => 'ttpos_item_erp_code']);
            }
            if (!$table->hasColumn('ttpos_category_name')) {
                $table->addColumn('ttpos_category_name', 'string', ['limit' => 255, 'default' => '', 'comment' => 'TTPOS分类名称', 'after' => 'ttpos_category_uuid']);
            }
            if (!$table->hasColumn('ttpos_parent_category_uuid')) {
                $table->addColumn('ttpos_parent_category_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'TTPOS父分类UUID(关联ttpos_product_category.parent_uuid)', 'after' => 'ttpos_category_name']);
            }
            if (!$table->hasColumn('ttpos_parent_category_name')) {
                $table->addColumn('ttpos_parent_category_name', 'string', ['limit' => 255, 'default' => '', 'comment' => 'TTPOS父分类名称', 'after' => 'ttpos_parent_category_uuid']);
            }
            $table->update();
        }

        // 2. 为 ttpos_takeout_order_item_modifier 表添加字段
        if ($this->hasTable('takeout_order_item_modifier')) {
            $modifierTable = $this->table('takeout_order_item_modifier');
            // 检查字段是否已存在
            if (!$modifierTable->hasColumn('ttpos_price')) {
                $modifierTable->addColumn('ttpos_price', 'decimal', ['precision' => 20, 'scale' => 4, 'default' => '0.0000', 'comment' => 'TTPOS店内价格(元,4位小数)', 'after' => 'price']);
            }
            if (!$modifierTable->hasColumn('ttpos_category_uuid')) {
                $modifierTable->addColumn('ttpos_category_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'TTPOS分类UUID(关联ttpos_product_category.uuid)', 'after' => 'ttpos_flavor_name']);
            }
            if (!$modifierTable->hasColumn('ttpos_category_name')) {
                $modifierTable->addColumn('ttpos_category_name', 'string', ['limit' => 255, 'default' => '', 'comment' => 'TTPOS分类名称', 'after' => 'ttpos_category_uuid']);
            }
            if (!$modifierTable->hasColumn('ttpos_parent_category_uuid')) {
                $modifierTable->addColumn('ttpos_parent_category_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'TTPOS父分类UUID(关联ttpos_product_category.parent_uuid)', 'after' => 'ttpos_category_name']);
            }
            if (!$modifierTable->hasColumn('ttpos_parent_category_name')) {
                $modifierTable->addColumn('ttpos_parent_category_name', 'string', ['limit' => 255, 'default' => '', 'comment' => 'TTPOS父分类名称', 'after' => 'ttpos_parent_category_uuid']);
            }
            $modifierTable->update();
        }
        
    }
}


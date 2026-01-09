<?php
declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

/**
 * 修改外卖订单商品表和修饰符表的分类名称字段类型为 text
 *
 * 修改字段:
 * - ttpos_takeout_order_item.ttpos_category_name: varchar(255) -> text
 * - ttpos_takeout_order_item.ttpos_parent_category_name: varchar(255) -> text
 * - ttpos_takeout_order_item_modifier.ttpos_category_name: varchar(255) -> text
 * - ttpos_takeout_order_item_modifier.ttpos_parent_category_name: varchar(255) -> text
 */
final class ModifyTakeoutOrderItemCategoryNameFieldsToText extends AbstractMigration
{
    public function change(): void
    {
        // 1. 修改 ttpos_takeout_order_item 表的字段类型
        if ($this->hasTable('takeout_order_item')) {
            $table = $this->table('takeout_order_item');
            // 检查字段是否存在
            if ($table->hasColumn('ttpos_category_name')) {
                $table->changeColumn('ttpos_category_name', 'text', ['null' => false, 'default' => '', 'comment' => 'TTPOS分类名称']);
            }
            if ($table->hasColumn('ttpos_parent_category_name')) {
                $table->changeColumn('ttpos_parent_category_name', 'text', ['null' => false, 'default' => '', 'comment' => 'TTPOS父分类名称']);
            }
            $table->update();
        }

        // 2. 修改 ttpos_takeout_order_item_modifier 表的字段类型
        if ($this->hasTable('takeout_order_item_modifier')) {
            $modifierTable = $this->table('takeout_order_item_modifier');
            // 检查字段是否存在
            if ($modifierTable->hasColumn('ttpos_category_name')) {
                $modifierTable->changeColumn('ttpos_category_name', 'text', ['null' => false, 'default' => '', 'comment' => 'TTPOS分类名称']);
            }
            if ($modifierTable->hasColumn('ttpos_parent_category_name')) {
                $modifierTable->changeColumn('ttpos_parent_category_name', 'text', ['null' => false, 'default' => '', 'comment' => 'TTPOS父分类名称']);
            }
            $modifierTable->update();
        }
    }
}


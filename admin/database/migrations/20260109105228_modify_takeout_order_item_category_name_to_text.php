<?php
declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

/**
 * 修改外卖订单商品表和修饰符表的分类名称字段类型为 TEXT
 *
 * 修改字段:
 * - ttpos_category_name: 从 VARCHAR(255) 改为 TEXT
 * - ttpos_parent_category_name: 从 VARCHAR(255) 改为 TEXT
 *
 * 原因: 这些字段存储多语言 JSON 字符串，长度可能超过 255 个字符
 */
final class ModifyTakeoutOrderItemCategoryNameToText extends AbstractMigration
{
    public function change(): void
    {
        // 1. 修改 ttpos_takeout_order_item 表的字段
        if ($this->hasTable('takeout_order_item')) {
            $table = $this->table('takeout_order_item');
            
            // 检查字段是否存在，如果存在则修改类型
            if ($table->hasColumn('ttpos_category_name')) {
                $table->changeColumn('ttpos_category_name', 'text', [
                    'null' => true,
                    'comment' => 'TTPOS分类名称（多语言JSON）'
                ])->update();
            }
            
            if ($table->hasColumn('ttpos_parent_category_name')) {
                $table->changeColumn('ttpos_parent_category_name', 'text', [
                    'null' => true,
                    'comment' => 'TTPOS父分类名称（多语言JSON）'
                ])->update();
            }
        }

        // 2. 修改 ttpos_takeout_order_item_modifier 表的字段
        if ($this->hasTable('takeout_order_item_modifier')) {
            $modifierTable = $this->table('takeout_order_item_modifier');
            
            // 检查字段是否存在，如果存在则修改类型
            if ($modifierTable->hasColumn('ttpos_category_name')) {
                $modifierTable->changeColumn('ttpos_category_name', 'text', [
                    'null' => true,
                    'comment' => 'TTPOS分类名称（多语言JSON）'
                ])->update();
            }
            
            if ($modifierTable->hasColumn('ttpos_parent_category_name')) {
                $modifierTable->changeColumn('ttpos_parent_category_name', 'text', [
                    'null' => true,
                    'comment' => 'TTPOS父分类名称（多语言JSON）'
                ])->update();
            }
        }
    }
}


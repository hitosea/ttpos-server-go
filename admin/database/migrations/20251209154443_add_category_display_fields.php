<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCategoryDisplayFields extends Migrator
{
    /**
     * 为 ttpos_product_category 表添加显示渠道控制字段
     * - is_display_in_store: 是否在店内显示
     * - is_display_in_takeout: 是否在外卖平台显示
     */
    public function change()
    {
        $table = $this->table('product_category');
        
        // 检查字段是否存在，如果不存在则添加
        if (!$table->hasColumn('is_display_in_store')) {
            $table->addColumn('is_display_in_store', 'integer', [
                'limit' => 1,
                'signed' => false,
                'default' => 1,
                'comment' => '是否在店内显示: 1-是 0-否',
                'after' => 'status'
            ]);
        }
        
        if (!$table->hasColumn('is_display_in_takeout')) {
            $table->addColumn('is_display_in_takeout', 'integer', [
                'limit' => 1,
                'signed' => false,
                'default' => 0,
                'comment' => '是否在外卖平台显示: 1-是 0-否',
                'after' => 'is_display_in_store'
            ]);
        }
        
        // 添加索引（如果不存在）
        if (!$table->hasIndexByName('idx_is_display_in_store')) {
            $table->addIndex('is_display_in_store', ['name' => 'idx_is_display_in_store']);
        }
        
        if (!$table->hasIndexByName('idx_is_display_in_takeout')) {
            $table->addIndex('is_display_in_takeout', ['name' => 'idx_is_display_in_takeout']);
        }
        
        $table->update();
    }
}


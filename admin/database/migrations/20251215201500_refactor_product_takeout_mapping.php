<?php

use think\migration\Migrator;
use think\migration\db\Column;

class RefactorProductTakeoutMapping extends Migrator
{
    /**
     * 重构外卖商品映射：
     * 1. 为 ttpos_product_package_takeout 表添加通用的 source 和 source_product_id 字段
     * 2. 删除 grab_product_id 字段（改用通用字段）
     * 3. 删除 ttpos_product_map 表（功能合并到 ttpos_product_package_takeout）
     */
    public function change()
    {
        // 1. 修改 ttpos_product_package_takeout 表
        if ($this->hasTable('product_package_takeout')) {
            $table = $this->table('product_package_takeout');
            
            // 添加通用的来源字段
            if (!$table->hasColumn('source')) {
                $table->addColumn('source', 'string', ['limit' => 50, 'default' => '', 'comment' => '来源平台(grab/foodpanda/lineman等)', 'after' => 'image_file_uuid']);
            }
            
            if (!$table->hasColumn('source_product_id')) {
                $table->addColumn('source_product_id', 'string', ['limit' => 500, 'default' => '', 'comment' => '来源平台商品唯一ID', 'after' => 'source']);
            }
            
            // 检查是否存在 grab_product_id 字段，如果存在则数据迁移后删除
            if ($table->hasColumn('grab_product_id')) {
                // 删除旧字段
                $table->removeColumn('grab_product_id');
            }
            
            // 添加联合索引
            if (!$table->hasIndex(['source', 'source_product_id'])) {
                $table->addIndex(['source', 'source_product_id'], ['name' => 'idx_source_product']);
            }
            
            $table->update();
        }
        
        // 2. 删除 ttpos_product_map 表（功能已合并到 ttpos_product_package_takeout）
        if ($this->hasTable('product_map')) {
            $this->table('product_map')->drop()->save();
        }
    }
}


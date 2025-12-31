<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSelectionRangeFields extends Migrator
{
    /**
     * 为商品表、属性组表、套餐分组表添加选择范围字段
     * 
     * 涉及表：
     * 1. ttpos_product_package - 添加 sauce_min_selection
     * 2. ttpos_product_package_attribute_group - 添加 min_selection
     * 3. ttpos_product_package_group - 添加 optional_min_count，修改 optional_count 注释
     * 
     * 版本兼容性：v2.11 → v2.12
     * 
     * 注意：旧数据迁移逻辑已移至 Go 命令，因为数据迁移较耗时
     * 执行方式：在部署后手动执行以下命令
     * ```bash
     * cd main && ./ttpos migrate-product-selection-range
     * ```
     */
    public function change()
    {
        // 1. 为 ttpos_product_package 表添加 sauce_min_selection 字段
        $this->addSauceMinSelectionField();
        
        // 2. 为 ttpos_product_package_attribute_group 表添加 min_selection 字段
        $this->addMinSelectionField();
        
        // 3. 为 ttpos_product_package_group 表添加 optional_min_count 字段
        $this->addOptionalMinCountField();
    }
    
    /**
     * 为商品套餐表添加小料最小选择数量字段
     */
    private function addSauceMinSelectionField()
    {
        $table = $this->table('product_package');
        
        if (!$table->hasColumn('sauce_min_selection')) {
            $table->addColumn('sauce_min_selection', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '小料最小选择数量',
                'after' => 'sauce_required'
            ])->update();
        }
    }
    
    /**
     * 为商品属性组表添加最小选择数量字段
     */
    private function addMinSelectionField()
    {
        $table = $this->table('product_package_attribute_group');
        
        if (!$table->hasColumn('min_selection')) {
            $table->addColumn('min_selection', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '最小选择数量',
                'after' => 'is_must'
            ])->update();
        }
    }
    
    /**
     * 为套餐分组表添加最小可选数量字段
     */
    private function addOptionalMinCountField()
    {
        $table = $this->table('product_package_group');
        
        if (!$table->hasColumn('optional_min_count')) {
            $table->addColumn('optional_min_count', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '最小可选数量',
                'after' => 'group_type'
            ])->update();
        }
        
        // 修改 optional_count 字段注释（字段名保持不变）
        // 注意：ThinkPHP migration 不直接支持修改注释，需要通过原生SQL
        $this->execute("
            ALTER TABLE ttpos_product_package_group 
            MODIFY COLUMN optional_count INT NOT NULL DEFAULT 0 
            COMMENT '最大可选数量，表示本组商品中最多可以选择多少个商品'
        ");
    }
}


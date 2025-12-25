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
     */
    public function change()
    {
        // 1. 为 ttpos_product_package 表添加 sauce_min_selection 字段
        $this->addSauceMinSelectionField();
        
        // 2. 为 ttpos_product_package_attribute_group 表添加 min_selection 字段
        $this->addMinSelectionField();
        
        // 3. 为 ttpos_product_package_group 表添加 optional_min_count 字段
        $this->addOptionalMinCountField();
        
        // 4. 迁移旧数据
        $this->migrateOldData();
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
    
    /**
     * 迁移旧数据到新字段
     * 
     * 转换规则：
     * 1. 加料范围：
     *    - 不开启必选，不设置最大可选 → 可选0到加料值数量（min=0, max=加料数量）
     *    - 开启必选 → 最小值=1（min=1）
     *    - 开启最大可选 → 最大值=具体的值（max=设置的值）
     * 2. 属性范围：
     *    - is_must = 1 → min_selection = 1
     *    - max_selection = 0 → 设置为属性值数量
     * 3. 套餐分组范围：
     *    - group_type = 1 (可选) → optional_min_count = 1
     *    - group_type = 0 (固定) → optional_min_count = 分组商品数量
     */
    private function migrateOldData()
    {
        // 1. 迁移加料：sauce_required → sauce_min_selection
        // 开启必选时，最小选择=1
        $this->execute("
            UPDATE ttpos_product_package 
            SET sauce_min_selection = CASE 
                WHEN sauce_required = 1 THEN 1 
                ELSE 0 
            END
            WHERE sauce_min_selection = 0
        ");
        
        // 2. 修正加料：sauce_max_selection = 0 的情况
        // 如果未设置最大可选（为0），则设置为该商品的加料数量
        // 加料数据在 ttpos_product_bom 表中，product_sauce_uuid > 0 表示加料
        $this->execute("
            UPDATE ttpos_product_package pp
            SET pp.sauce_max_selection = (
                SELECT COUNT(DISTINCT pb.product_sauce_uuid)
                FROM ttpos_product_bom pb
                WHERE pb.product_package_uuid = pp.uuid
                AND pb.product_sauce_uuid > 0
                AND pb.delete_time = 0
            )
            WHERE pp.sauce_max_selection = 0
            AND EXISTS (
                SELECT 1
                FROM ttpos_product_bom pb
                WHERE pb.product_package_uuid = pp.uuid
                AND pb.product_sauce_uuid > 0
                AND pb.delete_time = 0
            )
        ");
        
        // 3. 迁移 is_must → min_selection
        $this->execute("
            UPDATE ttpos_product_package_attribute_group 
            SET min_selection = CASE 
                WHEN is_must = 1 THEN 1 
                ELSE 0 
            END
            WHERE min_selection = 0
        ");
        
        // 4. 修正 max_selection = 0 的情况
        // 如果 max_selection 为 0，设置为属性值数量
        $this->execute("
            UPDATE ttpos_product_package_attribute_group ppag
            SET ppag.max_selection = (
                SELECT COUNT(*)
                FROM ttpos_product_package_attribute ppa
                WHERE ppa.product_package_attribute_group_uuid = ppag.uuid
                AND ppa.delete_time = 0
            )
            WHERE ppag.max_selection = 0
            AND EXISTS (
                SELECT 1
                FROM ttpos_product_package_attribute ppa
                WHERE ppa.product_package_attribute_group_uuid = ppag.uuid
                AND ppa.delete_time = 0
            )
        ");
        
        // 5. 迁移套餐分组的可选范围
        // 可选分组：设置 optional_min_count = 1
        $this->execute("
            UPDATE ttpos_product_package_group 
            SET optional_min_count = 1
            WHERE group_type = 1 
            AND optional_min_count = 0
        ");
    }
}


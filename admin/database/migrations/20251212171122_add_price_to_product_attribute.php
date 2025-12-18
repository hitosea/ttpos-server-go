<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddPriceToProductAttribute extends Migrator
{
    /**
     * 为商品属性表和商品包属性表添加价格字段
     * 用于记录外卖平台属性的加价信息
     */
    public function change()
    {
        // 1. 商品属性表添加价格字段
        if ($this->hasTable('product_attribute')) {
            $table = $this->table('product_attribute');
            
            if (!$table->hasColumn('price')) {
                $table->addColumn('price', 'decimal', [
                    'precision' => 20,
                    'scale' => 4,
                    'default' => 0.00,
                    'comment' => '属性价格（外卖平台属性加价）',
                    'after' => 'sort'
                ]);
            }
            
            $table->update();
        }

        // 2. 商品包属性表添加价格字段
        if ($this->hasTable('product_package_attribute')) {
            $table = $this->table('product_package_attribute');
            
            if (!$table->hasColumn('price')) {
                $table->addColumn('price', 'decimal', [
                    'precision' => 20,
                    'scale' => 4,
                    'default' => 0.00,
                    'comment' => '属性价格（商品包级别的属性加价）',
                    'after' => 'is_default_selected'
                ]);
            }
            
            $table->update();
        }
    }
}

<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddImageUrlToProductPackage extends Migrator
{
    /**
     * 为商品包表添加 image_url 字段
     * 用于存储外部图片URL地址（当没有本地图片文件时使用）
     */
    public function change()
    {
        // 为 ttpos_product_package 表添加 image_url 字段
        if ($this->hasTable('product_package')) {
            $table = $this->table('product_package');
            
            if (!$table->hasColumn('image_url')) {
                $table->addColumn('image_url', 'string', ['limit' => 1000, 'default' => '', 'comment' => '外部图片URL地址（当image_file_uuid为空时使用）', 'after' => 'image_file_uuid']);
            }
            
            $table->update();
        }
    }
}


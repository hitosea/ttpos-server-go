<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddClassificationCodeToProductCategoryTable extends Migrator
{
    /**
     * 添加 classification_code 字段到 ttpos_product_category 表
     */
    public function change()
    {
        $table = $this->table('product_category');

        // 检查字段是否已存在
        if (!$table->hasColumn('code')) {
            $table->addColumn('code', 'string', ['limit' => 255, 'default' => '', 'comment' => '分类编码', 'after' => 'sort'])
                  ->update();
        }
    }
}

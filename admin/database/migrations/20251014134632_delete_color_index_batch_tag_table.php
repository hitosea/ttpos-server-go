<?php

use think\migration\Migrator;

class DeleteColorIndexBatchTagTable extends Migrator
{
    /**
     * 删除分批类型表 color 字段唯一索引
     */
    public function change()
    {
        // 判断 batch_tag 表是否存在
        if ($this->hasTable('batch_tag')) {
            $table = $this->table('batch_tag');
            // 判断唯一索引 unique_color 是否存在，存在则删除
            if ($table->hasIndex('color')) {
                $table->removeIndex("color");
                $table->update();
            }
        }              
    }
}

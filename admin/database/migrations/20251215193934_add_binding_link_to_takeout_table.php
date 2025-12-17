<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddBindingLinkToTakeoutTable extends Migrator
{
    /**
     * 添加 binding_link 字段到 takeout 表
     * - 新增 binding_link 字段，用于缓存平台绑定链接
     */
    public function change()
    {
        $table = $this->table('takeout');
        // 检查字段是否已存在
        if (!$table->hasColumn('skip')) {
            $table->addColumn('skip', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'after' => 'is_bound', 'comment' => '是否跳过绑定(1:跳过 0:不跳过)'])->update();
        }
        // 检查字段是否已存在
        if (!$table->hasColumn('binding_link')) {
            $table->addColumn('binding_link', 'string', ['limit' => 500, 'default' => '', 'after' => 'skip', 'comment' => '平台绑定链接（缓存用）'])->update();
        }
    }
}


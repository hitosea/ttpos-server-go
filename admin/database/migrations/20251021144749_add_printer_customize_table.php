<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPrinterCustomizeTable extends Migrator
{
    /**
     * 添加打印机定制表
     */
    public function change()
    {
        // 检查表是否已存在
        if ($this->hasTable('printer_customize')) {
            return;
        }

        $table = $this->table('printer_customize', [
            'id' => true,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '打印机定制表'
        ]);

        $table->addColumn('uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => 'ID'])
              ->addColumn('name', 'string', ['limit' => 255, 'default' => '', 'comment' => '名称'])
              ->addColumn('template_id', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '模板ID'])
              ->addColumn('is_adv', 'integer', ['null' => false, 'default' => 0, 'comment' => '是否高级'])
              ->addColumn('is_use', 'integer', ['null' => false, 'default' => 0, 'comment' => '是否使用'])
              ->addColumn('data', 'text', ['limit' => \Phinx\Db\Adapter\MysqlAdapter::TEXT_LONG, 'null' => true, 'comment' => '定制数据'])
              ->addColumn('create_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
              ->addColumn('update_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
              ->addColumn('delete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
              ->addIndex(['uuid'])
              ->addIndex(['name'])
              ->addIndex(['is_use'])
              ->addIndex(['create_time'])
              ->addIndex(['delete_time'])
              ->create();
    }
}

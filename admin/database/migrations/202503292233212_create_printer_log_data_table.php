<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreatePrinterLogDataTable extends Migrator
{
    public function change()
    {
        if (!$this->hasTable('printer_log_data')) {
            $table = $this->table('printer_log_data', ['comment' => '打印日志数据表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
                ->addColumn('log_uuid', 'biginteger', ['default' => 0, 'comment' => '打印日志UUID'])
                ->addColumn('data', 'text', ['limit' => 4294967295, 'comment' => '打印数据'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->addIndex(['log_uuid'], ['unique' => true])
                ->create();
        }
    }
} 
<?php

use think\migration\Migrator;

class CreateLanPrinterScanTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        if (!$this->hasTable('lan_printer_scan')) {
            $table = $this->table('lan_printer_scan', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '局域网打印机扫描表']);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => 'uuid'])
                ->addColumn('ip', 'string', ['null' => false, 'default' => '', 'comment' => 'ip'])
                ->addColumn('port', 'integer', ['null' => false, 'default' => 0, 'comment' => '端口'])
                ->addColumn('status', 'integer', ['null' => false, 'default' => 0, 'comment' => '状态 0: 离线 1: 在线'])
                ->addColumn('source_device_sn', 'string', ['null' => false, 'default' => '', 'comment' => '来源设备SN'])
                ->addColumn('remark', 'string', ['null' => false, 'default' => '', 'comment' => '备注'])
                ->addColumn('create_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '删除时间'])
                ->create();
        }
    }
}

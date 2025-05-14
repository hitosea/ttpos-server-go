<?php

use think\migration\Migrator;

class AddIsUsbToPrinter extends Migrator
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
        $table = $this->table('printer');
        if (!$table->hasColumn('is_usb')) {
            $table->addColumn('is_usb', 'integer', ['default' => 0, 'comment' => '是否usb', 'after' => 'sort']);
        }
        if (!$table->hasColumn('status')) {
            $table->addColumn('status', 'integer', ['default' => 1, 'comment' => '状态, 0-离线 1-在线', 'after' => 'is_usb']);
        }
        if (!$table->hasColumn('last_heartbeat_time')) {
            $table->addColumn('last_heartbeat_time', 'integer', ['default' => 0, 'comment' => '最后心跳时间', 'after' => 'status']);
        }
        $table->update();
    }
}

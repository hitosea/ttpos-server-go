<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddPrinterStatusCheckWidthFields extends Migrator
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
        // 检查表是否存在
        if ($this->hasTable('printer')) {
            $table = $this->table('printer');
            
            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('enable_status_check')) {
                $table->addColumn('enable_status_check', 'integer', ['limit' => 1, 'default' => 1, 'comment' => '是否启用状态检查 0-关闭 1-开启', 'after' => 'width']);
            }
            
            $table->update();
        }
    }
}

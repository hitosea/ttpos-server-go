<?php

use think\migration\Migrator;
use Phinx\Db\Adapter\MysqlAdapter;

class UpdateOrderAbnormalLogTable extends Migrator
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
        if ($this->hasTable('sale_order_abnormal_log')) {
            $table = $this->table('sale_order_abnormal_log', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '订单异常日志表']);
            $table->rename('sale_order_abnormal_record');
            $table->save();
        }
    }
}

<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsMegerToStatisticsSale extends Migrator
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
        $table = $this->table('statistics_sale');
        if (!$table->hasColumn('is_meger')) {
            $table->addColumn('is_meger', 'integer', ['default' => 0, 'comment' => '是否合单', 'after' => 'no_refund_tax'])
                ->update();
        }
        if (!$table->hasColumn('is_special')) {
            $table->addColumn('is_special', 'integer', ['default' => 0, 'comment' => '是否特殊订单', 'after' => 'is_meger'])
                ->update();
        }
    }
}

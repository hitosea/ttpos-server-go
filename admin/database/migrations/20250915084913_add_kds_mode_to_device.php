<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddKdsModeToDevice extends Migrator
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
        $table = $this->table('device');
        if (!$table->hasColumn('kds_mode')) {
            $table->addColumn('kds_mode', 'integer', ['default' => 0, 'comment' => '厨显端模式 0-默认，传菜模式; 1-制菜模式; 2-制菜+传菜模式', 'after' => 'queue_url'])
                ->update();
        }
        $table = $this->table('production_order_product');
        if (!$table->hasColumn('made_time')) {
            $table->addColumn('made_time', 'integer', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '制作完成时间', 'after' => 'finished_time'])
                ->update();
        }
        if (!$table->hasColumn('make_status')) {
            $table->addColumn('make_status', 'integer', ['null' => false, 'default' => 0, 'comment' => '制作状态 0-默认，未制作完成，1-已制作完成，2-已恢复到制作中', 'after' => 'finished_time'])
                ->update();
        }
    }
}

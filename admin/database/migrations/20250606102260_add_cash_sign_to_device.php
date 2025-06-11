<?php

use think\migration\Migrator;

class AddCashSignToDevice extends Migrator
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
        if (!$table->hasColumn('cash_sign')) {
            $table->addColumn('cash_sign', 'string', ['default' => '', 'comment' => '收银终端标识', 'after' => 'user_agent']);
        }
        if (!$table->hasColumn('cash_box_id')) {
            $table->addColumn('cash_box_id', 'string', ['default' => '', 'comment' => '现金箱ID', 'after' => 'cash_sign']);
        }
        if (!$table->hasColumn('access_token')) {
            $table->addColumn('access_token', 'string', ['default' => '', 'comment' => '访问令牌', 'after' => 'cash_box_id']);
        }
        if (!$table->hasColumn('queue_url')) {
            $table->addColumn('queue_url', 'string', ['default' => '', 'comment' => '关联队列url', 'after' => 'access_token']);
        }
        $table->update();
    }
}

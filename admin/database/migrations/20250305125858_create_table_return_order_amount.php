<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTableReturnOrderAmount extends Migrator
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
        $table = $this->table('return_order_amount', [
            'id' => false,
            'primary_key' => 'id',
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '退款金额表'
        ]);

        if (!$this->hasTable('return_order_amount')) {
            $table->addColumn('id', 'integer', ['signed' => false, 'identity' => true, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '会员充值订单操作日志ID'])
                ->addColumn('operator_name', 'string', ['limit' => 50, 'default' => '', 'comment' => '操作员姓名'])
                ->addColumn('operator_email', 'string', ['limit' => 50, 'default' => '', 'comment' => '操作员电子邮件'])
                ->addColumn('client', 'string', ['limit' => 50, 'default' => '', 'comment' => '客户端信息'])
                ->addColumn('message', 'string', ['limit' => 255, 'default' => '', 'comment' => '消息内容'])
                ->addColumn('action', 'string', ['limit' => 255, 'default' => '', 'comment' => '操作'])
                ->addColumn('data', 'string', ['limit' => 255, 'default' => '', 'comment' => '数据'])
                ->addColumn('recharge_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '充值订单ID'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->create();
        }
    }
}

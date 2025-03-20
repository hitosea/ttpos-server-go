<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRelatedUuidToLlPaymentOrderTable extends Migrator
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
        $table = $this->table('ll_payment_order');
        if (!$table->hasColumn('related_uuid')) {
            $table->addColumn('related_uuid', 'biginteger', [
                'null' => false,
                'default' => 0,
                'comment' => '关联订单ID',
                'after' => 'payment_order_uuid',
            ]);
            $table->addIndex(['related_uuid'], ['name' => 'related_uuid']);
        }
        if (!$table->hasColumn('expired_time')) {
            $table->addColumn('expired_time', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '过期时间',
                'after' => 'll_create_time',
            ]);
        }
        // 修改link_url字段为text类型
        $table->changeColumn('link_url', 'text', ['null' => true, 'default' => null, 'comment' => 'lianlian订单支付链接']);
        // 更新表
        $table->update();
    }
}

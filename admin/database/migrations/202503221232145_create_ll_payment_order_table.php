<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateLlPaymentOrderTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     */
    public function change()
    {
        $table = $this->table('ll_payment_order', [
            'id' => false,
            'primary_key' => 'id',
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => 'lianlian支付订单'
        ]);

        if (!$this->hasTable('ll_payment_order')) {
            $table->addColumn('id', 'integer', ['signed' => false, 'identity' => true, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'UUID'])
                ->addColumn('payment_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '自己系统的支付订单ID'])
                ->addColumn('related_type', 'tinyinteger', ['default' => 0, 'comment' => '关联订单类型：0-销售订单；1-充值订单'])
                ->addColumn('merchant_id', 'string', ['limit' => 255, 'default' => '', 'comment' => 'lianlian商户号'])
                ->addColumn('merchant_order_id', 'string', ['limit' => 255, 'default' => '', 'comment' => '自己系统的为支付生成的订单号'])
                ->addColumn('order_id', 'string', ['limit' => 255, 'default' => '', 'comment' => 'lianlian订单ID'])
                ->addColumn('order_type', 'string', ['limit' => 50, 'default' => '', 'comment' => '订单类型'])
                ->addColumn('order_status', 'string', ['limit' => 50, 'default' => '', 'comment' => 'lianlian订单状态 PI-初始化(未访问支付页操作) WP-等待支付 PS-支付成功 PF-支付失败 PE-支付已过期'])
                ->addColumn('order_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0.00, 'comment' => 'lianlian订单金额'])
                ->addColumn('order_currency', 'string', ['limit' => 50, 'default' => '', 'comment' => 'lianlian订单货币'])
                ->addColumn('full_name', 'string', ['limit' => 50, 'default' => '', 'comment' => '订单人名称'])
                ->addColumn('order_desc', 'string', ['limit' => 50, 'default' => '', 'comment' => '订单描述'])
                ->addColumn('link_url', 'string', ['limit' => 5000, 'default' => '', 'comment' => 'lianlian订单支付链接'])
                ->addColumn('merchant_user_id', 'string', ['limit' => 255, 'default' => '', 'comment' => '自己系统的用户ID'])
                ->addColumn('ll_create_time', 'string', ['limit' => 250, 'default' => '0', 'comment' => 'lianlian订单创建时间'])
                ->addColumn('pay_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '支付时间'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['order_id'], ['name' => 'order_id'])
                ->addIndex(['merchant_order_id'], ['name' => 'merchant_order_id'])
                ->addIndex(['payment_order_uuid'], ['name' => 'payment_order_uuid'])
                ->create();
        }
    }
}
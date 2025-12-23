<?php
/**
 * 为支付方式表添加 erpnext_payment_id 字段
 * 
 * 优化: opt-251223-001-admin-payment-management
 * 需求: 新增支付方式时，name保存在erpnext_payment，新增字段erpnext_payment_id保存PaymentId
 */

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpnextPaymentIdToPaymentMethod extends Migrator
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
        if ($this->hasTable('payment_method')) {
            $table = $this->table('payment_method');
            
            // 检查字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('erpnext_payment_id')) {
                $table->addColumn('erpnext_payment_id', 'string', [
                    'limit' => 255,
                    'default' => '',
                    'null' => false,
                    'comment' => 'ERPNext支付方式ID',
                    'after' => 'erpnext_payment'
                ]);
            }
            
            $table->update();
        }
    }
}


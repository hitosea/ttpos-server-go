<?php
/**
 * 为支付方式表添加 is_show_kiosk 字段
 * 
 * 任务: story-shop-payment-management
 * 需求: 结账显示需要加上自助点餐机
 */

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsShowKioskToPaymentMethod extends Migrator
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
            if (!$table->hasColumn('is_show_kiosk')) {
                $table->addColumn('is_show_kiosk', 'integer', [
                    'limit' => 3,
                    'default' => 0,
                    'null' => false,
                    'comment' => '0-不显示 1-自助点餐机结账显示',
                    'after' => 'is_show_assistant'
                ]);
            }
            
            $table->update();
        }
    }
}

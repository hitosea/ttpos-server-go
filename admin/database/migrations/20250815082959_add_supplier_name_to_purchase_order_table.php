<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSupplierNameToPurchaseOrderTable extends Migrator
{
    /**
     * 添加供应商名称字段到采购订单表
     * Change Method.
     *
     * Rename the method to "change" to make these the default actions.
     * More information on this method is available here:
     * https://book.cakephp.org/phinx/0/en/migrations.html#the-change-method
     * Uncomment this method if you would like to use it.
     *
     * @return void
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('purchase_order')) {
            return;
        }
        
        $table = $this->table('purchase_order');
        
        // 检查字段是否已经存在
        if (!$table->hasColumn('supplier_name')) {
            $table->addColumn('supplier_name', 'string', [
                'limit'   => 100,
                'default' => '',
                'null'    => false,
                'comment' => '供应商名称',
                'after'   => 'order_type'
            ]);
            
            $table->update();
        }
    }
}

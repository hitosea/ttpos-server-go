<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateSaleOrderInvoiceInfoTable extends Migrator
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
        if (!$this->hasTable('sale_order_invoice_info')) {
            $table = $this->table('sale_order_invoice_info', ['engine' => 'InnoDB', 'comment' => '销售订单发票信息表']);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一ID'])
                ->addColumn('sale_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '销售订单ID'])
                ->addColumn('company_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '公司名称'])
                ->addColumn('company_addr', 'string', ['limit' => 255, 'default' => '', 'comment' => '公司地址'])
                ->addColumn('company_tax_number', 'string', ['limit' => 255, 'default' => '', 'comment' => '公司税号'])
                ->addColumn('company_phone', 'string', ['limit' => 255, 'default' => '', 'comment' => '公司电话'])
                ->addColumn('print_num', 'integer', ['signed' => false, 'default' => 0, 'comment' => '打印次数'])
                ->addColumn('create_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['sale_order_uuid'], ['name' => 'idx_sale_order_uuid'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->create();
        }
    }
}

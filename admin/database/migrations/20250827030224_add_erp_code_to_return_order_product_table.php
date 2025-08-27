<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpCodeToReturnOrderProductTable extends Migrator
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
        $table = $this->table('return_order_product');
        // 检查字段是否存在，如果不存在则添加
        if (!$table->hasColumn('erp_code')) {
            $table->addColumn('erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERP系统商品编码', 'after' => 'product_total_amount']);
        }
        $table->update();
    }
}

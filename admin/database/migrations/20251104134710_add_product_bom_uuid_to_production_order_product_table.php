<?php

use think\migration\Migrator;
use think\migration\db\Column;
use Phinx\Db\Adapter\MysqlAdapter;

class AddProductBomUuidToProductionOrderProductTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-change-method
     *
     * === NOTE: ===
     * DANGER! Do NOT use constants in this method in order to avoid the following issue:
     * When you use a constant in the code and then you remove the constant, or change the constant in your code
     * then the migration code will fail due to constant not found.
     * As a result, you will not be able to rollback to the previous version or migrate forward.
     * In that case, you will have to manually edit the migration code to fix the issue.
     *
     * Example:
     * $table = $this->table('users');
     * $table->addColumn('name', 'string')
     *       ->addColumn('created_at', 'datetime')
     *       ->save();
     */
    public function change()
    {

        $table = $this->table('production_order_product');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('product_bom_uuid')) {
            $table->addColumn('product_bom_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '商品BOM ID', 'after' => 'flavor_name'])
                ->save();
        }
    }
}

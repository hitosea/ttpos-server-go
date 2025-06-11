<?php

use think\migration\Migrator;

class AddIndexToWarehouseOutFormItem extends Migrator
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
        try {
            $table = $this->table('warehouse_out_form_item');
            if (!$table->hasIndex('idx_warehouse_out_form_uuid')) {
                $table->addIndex(['warehouse_out_form_uuid'], ['name' => 'idx_warehouse_out_form_uuid']);
            }
            if (!$table->hasIndex('idx_material_uuid')) {
                $table->addIndex(['material_uuid'], ['name' => 'idx_material_uuid']);
            }
            if (!$table->hasIndex('idx_product_bom_uuid')) {
                $table->addIndex(['product_bom_uuid'], ['name' => 'idx_product_bom_uuid']);
            }
            $table->update();

            // 
            $table = $this->table('product_bom');
            if (!$table->hasIndex('idx_product_flavor_uuid')) {
                $table->addIndex(['product_flavor_uuid'], ['name' => 'idx_product_flavor_uuid']);
            }
            if (!$table->hasIndex('idx_product_package_uuid')) {
                $table->addIndex(['product_package_uuid'], ['name' => 'idx_product_package_uuid']);
            }
            $table->update();
        } catch (\Exception $e) {
            trace($e->getMessage(), 'error');
        }
    }
}

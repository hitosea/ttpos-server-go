<?php
/**
 * 添加对方机构相关字段到仓库出入库记录表
 */

use think\migration\Migrator;

class AddOtherOrgFieldsToWarehouseInOutLogTable extends Migrator
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
        if ($this->hasTable('warehouse_in_out_log')) {
            $table = $this->table('warehouse_in_out_log');
            
            // 检查other_org_uuid字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('other_org_uuid')) {
                $table->addColumn('other_org_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '对方机构ID', 'after' => 'order_no']);
            }
            
            // 检查other_org_type字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('other_org_type')) {
                $table->addColumn('other_org_type', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '对方机构类型 0:供应商 1:客户', 'after' => 'other_org_uuid']);
            }
            
            // 检查other_org_name字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('other_org_name')) {
                $table->addColumn('other_org_name', 'string', ['limit' => 255, 'comment' => '对方机构名称', 'after' => 'other_org_type']);
            }
            
            $table->update();
        }
    }
}

<?php
/**
 * 创建同步任务表
 */

use think\migration\Migrator;

class CreateSyncTaskTable extends Migrator
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
        // 创建同步任务主表
        if (!$this->hasTable('sync_task')) {
            $table = $this->table('sync_task', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '同步任务表',
                'id' => false,
                'primary_key' => ['id']
            ]);
            
            $table->addColumn('id', 'integer', ['signed' => false, 'identity' => true, 'comment' => '自增ID'])
                  ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '同步任务UUID'])
                  ->addColumn('status', 'integer', ['signed' => false, 'default' => 0, 'comment' => '同步状态: 0-进行中, 1-已完成, 2-失败'])
                  ->addColumn('total_count', 'integer', ['signed' => false, 'default' => 0, 'comment' => '总任务数'])
                  ->addColumn('success_count', 'integer', ['signed' => false, 'default' => 0, 'comment' => '成功任务数'])
                  ->addColumn('fail_count', 'integer', ['signed' => false, 'default' => 0, 'comment' => '失败任务数'])
                  ->addColumn('panic', 'text', ['null' => true, 'comment' => 'panic错误信息']) 
                  ->addColumn('start_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '开始时间'])
                  ->addColumn('end_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '结束时间'])
                  ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                  ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                  ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                  ->addIndex(['uuid'], ['unique' => true])
                  ->addIndex(['status'])
                  ->addIndex(['create_time'])
                  ->create();
        }

        // 创建同步任务明细表
        if (!$this->hasTable('sync_task_item')) {
            $table = $this->table('sync_task_item', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '同步任务明细表',
                'id' => false,
                'primary_key' => ['id']
            ]);
            
            $table->addColumn('id', 'integer', ['signed' => false, 'identity' => true, 'comment' => '自增ID'])
                  ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '同步任务明细UUID'])
                  ->addColumn('sync_task_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '同步任务UUID'])
                  ->addColumn('task_type', 'string', ['limit' => 50, 'default' => '', 'comment' => '任务类型: product_category-商品分类, material_category-物品分类, tax-税类, unit-单位, warehouse-仓库, material-物品, flavor-规格, attribute-属性, sauce-加料, product-商品, bom_card-成本卡, supplier-供应商, warehouse_stock-仓库物品库存'])
                  ->addColumn('task_name', 'string', ['limit' => 100, 'default' => '', 'comment' => '任务名称'])
                  ->addColumn('status', 'integer', ['signed' => false, 'default' => 0, 'comment' => '任务状态: 0-待执行, 1-执行中, 2-已完成, 3-失败'])
                  ->addColumn('error_message', 'text', ['null' => true, 'comment' => '错误消息'])
                  ->addColumn('start_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '开始时间'])
                  ->addColumn('end_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '结束时间'])
                  ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                  ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                  ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                  ->addIndex(['uuid'], ['unique' => true])
                  ->addIndex(['sync_task_uuid'])
                  ->addIndex(['task_type'])
                  ->addIndex(['status'])
                  ->create();
        }
    }
}


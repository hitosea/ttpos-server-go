<?php
/**
 * 添加对方机构相关字段到仓库出入库记录表
 */

use think\migration\Migrator;

class AddBatchFieldsToProductPackageTable extends Migrator
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
        if ($this->hasTable('product_package')) {
            $table = $this->table('product_package');
            
            // 检查is_batch字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('is_batch')) {
                $table->addColumn('is_batch', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否是分批商品, 0-否 1-是', 'after' => 'actual_sale_num']);
            }
            $table->update();
        }

        if ($this->hasTable('sale_order_product')) {
            $table = $this->table('sale_order_product');
            
            // 检查batch_time字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('is_batch')) {
                $table->addColumn('is_batch', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否是分批商品, 0-否 1-是', 'after' => 'open_member_discount']);
            }
            if (!$table->hasColumn('batch_time')) {
                $table->addColumn('batch_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '分批时间(时间戳)，表示该商品实际送厨到厨房的时间', 'after' => 'is_batch']);
            }
            // 检查batch_tag_uuid字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('batch_tag_uuid')) {
                $table->addColumn('batch_tag_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '分批类型UUID', 'after' => 'batch_time']);
            }
            $table->update();
        }

        if ($this->hasTable('production_order_product')) {
            $table = $this->table('production_order_product');
            
            // 检查is_batch字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('is_batch')) {
                $table->addColumn('is_batch', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否是分批商品, 0-否 1-是', 'after' => 'made_time']);
            }
            if (!$table->hasColumn('batch_time')) {
                $table->addColumn('batch_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '分批时间(时间戳)，表示该商品实际送厨到厨房的时间', 'after' => 'is_batch']);
            }
            if (!$table->hasColumn('batch_tag_uuid')) {
                $table->addColumn('batch_tag_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '分批类型UUID', 'after' => 'batch_time']);
            }
            $table->update();
        }
    }
}

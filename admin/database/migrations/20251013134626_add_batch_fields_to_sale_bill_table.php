<?php
/**
 * 添加对方机构相关字段到仓库出入库记录表
 */

use think\migration\Migrator;

class AddBatchFieldsToSaleBillTable extends Migrator
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
        if ($this->hasTable('sale_bill')) {
            $table = $this->table('sale_bill');
            
            // 检查is_batch字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('batch_tag_uuid')) {
                $table->addColumn('batch_tag_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '分批类型UUID', 'after' => 'reverse_settle_count']);
            }
            $table->update();
        }

    }
}

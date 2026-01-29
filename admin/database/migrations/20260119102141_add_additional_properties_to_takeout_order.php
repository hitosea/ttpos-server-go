<?php
/**
 * 为外卖订单表添加 additional_properties 字段
 * 
 * 用途: 存储订单的额外属性信息
 */

use think\migration\Migrator;

class AddAdditionalPropertiesToTakeoutOrder extends Migrator
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
        if ($this->hasTable('takeout_order')) {
            $table = $this->table('takeout_order');
            
            // 检查 additional_properties 字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('additional_properties')) {
                $table->addColumn('additional_properties', 'string', ['limit' => 1000, 'default' => '', 'null' => false, 'comment' => '订单额外属性信息', 'after' => 'raw_data']);
            }
            
            $table->update();
        }
    }
}

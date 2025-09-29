<?php
/**
 * 修改原料表name字段长度为1000
 */

use think\migration\Migrator;

class ModifyMaterialNameLength extends Migrator
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
        if ($this->hasTable('material')) {
            $table = $this->table('material');
            
            // 检查字段是否存在
            if ($table->hasColumn('name')) {
                // 修改name字段为text类型
                $table->changeColumn('name', 'text', ['comment' => '原料名称']);
                $table->update();
            }
        }
    }
}

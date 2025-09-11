<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsUsedFields extends Migrator
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
        // 成本卡
        $table = $this->table('product_bom_card');
        if (!$table->hasColumn('is_used')) {
            $table->addColumn('is_used', 'integer', ['null' => false, 'default' => 0, 'comment' => '是否被使用, 0-否 1-是', 'after' => 'num']);
        }
        $table->update();
        // 关联材料
        $table = $this->table('related_material');
        if (!$table->hasColumn('is_used')) {
            $table->addColumn('is_used', 'integer', ['null' => false, 'default' => 0, 'comment' => '是否被使用, 0-否 1-是', 'after' => 'base_unit_conversion_rate']);
        }
        $table->update();
    }
}

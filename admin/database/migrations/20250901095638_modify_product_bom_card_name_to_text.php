<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ModifyProductBomCardNameToText extends Migrator
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
        // 修改成本卡表的name字段为text类型
        $table = $this->table('product_bom_card');
        if ($table->hasColumn('name')) {
            $table->changeColumn('name', 'text', ['null' => false, 'comment' => '名称']);
        }
        $table->update();
    }
}

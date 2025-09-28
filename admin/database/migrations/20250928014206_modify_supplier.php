<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ModifySupplier extends Migrator
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
        // TODO 更新后可删除
        $table = $this->table('supplier');
        if (!$table->hasColumn('multi_language_name_uuid')) {
            $table->addColumn('multi_language_name_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '多语言名称UUID', 'after' => 'name']);
        }
        if ($table->hasColumn('name')) {
            $table->changeColumn('name', 'text', ['comment' => '供应商名称']);
        }
        $table->update();
    }
}

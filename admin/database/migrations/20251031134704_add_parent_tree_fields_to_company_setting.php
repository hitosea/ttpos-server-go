<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddParentTreeFieldsToCompanySetting extends Migrator
{
    // 迁移目标
    const TARGET = 'all';

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
        // 添加parent_company_uuids字段
        $table = $this->table('company_setting');
        if (!$table->hasColumn('parent_company_uuids')) {
            $table->addColumn('parent_company_uuids', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '父级公司UUID路径，从根节点到父节点，逗号分隔', 'after' => 'erpnext_admin_email']);
        }
        // 添加has_children字段
        if (!$table->hasColumn('has_children')) {
            $table->addColumn('has_children', 'integer', ['default' => 0, 'comment' => '是否含有子节点: 0-否 1-是', 'after' => 'parent_company_uuids']);
        }
        $table->update();
    }
}

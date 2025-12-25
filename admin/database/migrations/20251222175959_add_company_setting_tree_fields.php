<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddCompanySettingTreeFields extends Migrator
{
    // 迁移目标
    const TARGET = 'main';

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
        $table = $this->table('company_setting');
        
        // 添加父级公司UUID路径字段
        if (!$table->hasColumn('parent_company_uuids')) {
            $table->addColumn('parent_company_uuids', 'string', [
                'limit' => 255,
                'default' => '',
                'comment' => '父级公司UUID路径，从根节点到父节点，逗号分隔',
                'after' => 'headquarter_uuid'
            ]);
        }
        
        // 添加是否含有子节点字段
        if (!$table->hasColumn('has_children')) {
            $table->addColumn('has_children', 'integer', [
                'limit' => 10,
                'default' => 0,
                'comment' => '是否含有子节点: 0-否 1-是',
                'after' => 'parent_company_uuids'
            ]);
        }
        
        $table->update();
    }
}


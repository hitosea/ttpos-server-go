<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ModifyCompanySetting extends Migrator
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
        $table = $this->table('company_setting');
        // 添加erpnext_code字段
        if ($table->hasColumn('erpnext_code')) {
            $table->removeColumn('erpnext_code');
        }
        if ($table->hasColumn('erpnext_name')) {
            $table->removeColumn('erpnext_name');
        }
        // 添加erpnext_site_code字段
        if (!$table->hasColumn('erpnext_site_code')) {
            $table->addColumn('erpnext_site_code', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => 'ERPNext站点编码',
                'after' => 'delivery_config'
            ]);
        }
        // 添加erpnext_company_abbr字段
        if (!$table->hasColumn('erpnext_company_abbr')) {
            $table->addColumn('erpnext_company_abbr', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => 'ERPNext公司缩写',
                'after' => 'erpnext_site_code'
            ]);
        }
        if (!$table->hasColumn('erpnext_branch_name')) {
            $table->addColumn('erpnext_branch_name', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => 'ERPNext分支名称',
                'after' => 'erpnext_company_abbr'
            ]);
        }
        $table->update();
    }
}

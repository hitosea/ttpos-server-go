<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpnextAdminEmailToCompanySetting extends Migrator
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
        if (!$table->hasColumn('erpnext_admin_email')) {
            $table->addColumn('erpnext_admin_email', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => 'ERPNext 管理员邮箱',
                'after' => 'erpnext_pos_profile_name'
            ])->update();
        }
    }
}

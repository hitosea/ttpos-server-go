<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddEnableSmsToCompanySetting extends Migrator
{
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
        if (!$table->hasColumn('sms_quota')) {
            $table->addColumn('sms_quota', 'integer', ['default' => 0, 'null' => false, 'comment' => '短信配额', 'after' => 'is_open_buffet']);
        }
        if (!$table->hasColumn('enable_sms')) {
            $table->addColumn('enable_sms', 'integer', ['default' => 0, 'null' => false, 'comment' => '是否启用短信功能：0-否；1-是', 'after' => 'is_open_buffet']);
        }
        $table->update();
    }
}

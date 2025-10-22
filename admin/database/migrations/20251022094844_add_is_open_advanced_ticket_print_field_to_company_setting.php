<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsOpenAdvancedTicketPrintFieldToCompanySetting extends Migrator
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
        // 添加is_open_advanced_ticket_print字段
        $table = $this->table('company_setting');
        if (!$table->hasColumn('is_open_advanced_ticket_print')) {
            $table->addColumn('is_open_advanced_ticket_print', 'integer', ['default' => 0, 'comment' => '是否开启高级票据打印模板: 0-否 1-是', 'after' => 'is_open_local_print']);
        }
        $table->update();
    }
}

<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddAbbreviationToBatchTagTable extends Migrator
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
        // 为 batch_tag 表增加 abbreviation 字段
        $table = $this->table('batch_tag');
        if (!$table->hasColumn('abbreviation')) {
            $table->addColumn('abbreviation', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => '名称缩写',
                'after' => 'multi_language_name_uuid',
            ])->update();
        }
    }
}


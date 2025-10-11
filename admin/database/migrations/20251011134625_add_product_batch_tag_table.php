<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddProductBatchTagTable extends Migrator
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
        // 创建分批类型表，包含以上字段
        if (!$this->hasTable('batch_tag')) {
            $table = $this->table('batch_tag',  ['comment' => '分批类型']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
                ->addColumn('name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '分批类型名称'])
                ->addColumn('multi_language_name_uuid', 'biginteger', ['default' => 0, 'comment' => '多语言名称ID'])
                ->addColumn('color', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '颜色'])
                ->addColumn('sort', 'integer', ['default' => 0, 'comment' => '排序'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['color'], ['unique' => true, 'name' => 'unique_color'])
                ->create();
        }
    }
}

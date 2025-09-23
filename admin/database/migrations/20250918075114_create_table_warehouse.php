<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTableWarehouse extends Migrator
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
        // 创建仓库表，包含以上字段
        if (!$this->hasTable('warehouse')) {
            $table = $this->table('warehouse',  ['comment' => '仓库']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
                ->addColumn('name', 'text', ['comment' => '名称'])
                ->addColumn('multi_language_name_uuid', 'biginteger', ['default' => 0, 'comment' => '多语言名称UUID'])
                ->addColumn('type', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '仓库类型'])
                ->addColumn('code', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '仓库编码'])
                ->addColumn('status', 'integer', ['null' => false, 'default' => 0, 'comment' => '仓库状态'])
                ->addColumn('contact', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '联系人'])
                ->addColumn('phone', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '联系电话'])
                ->addColumn('address', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '地址'])
                ->addColumn('is_default', 'integer', ['null' => false, 'default' => 0, 'comment' => '是否默认：0-否；1-是'])
                ->addColumn('erp_code', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '关联erpnext'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->create();
        }

        $table = $this->table('supplier');

        if (!$table->hasColumn('code')) {
            $table->addColumn('code', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '编码', 'after' => 'uuid']);
        }
        if (!$table->hasColumn('erp_code')) {
            $table->addColumn('erp_code', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '关联erpnext', 'after' => 'staff_uuid']);
        }
        if (!$table->hasColumn('company_abbr')) {
            $table->addColumn('company_abbr', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '所属公司简称，如果是自己的company_abbr表示自己创建，其他值表示非自己创建', 'after' => 'erp_code']);
        }
        if (!$table->hasColumn('status')) {
            $table->addColumn('status', 'integer', ['null' => false, 'default' => 0, 'comment' => '状态：0-禁用；1-启用', 'after' => 'code']);
        }
        if ($table->hasColumn('name')) {
            $table->changeColumn('name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '供应商名称', 'after' => 'uuid']);
        }
        if ($table->hasColumn('contact_name')) {
            $table->changeColumn('contact_name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '联系人姓名']);
        }
        if ($table->hasColumn('contact_phone')) {
            $table->changeColumn('contact_phone', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '联系人电话']);
        }
        if ($table->hasColumn('address')) {
            $table->changeColumn('address', 'string', ['limit' => 500, 'null' => false, 'default' => '', 'comment' => '地址']);
        }
        $table->update();
    }
}

<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddSortToProductAttributeTables extends Migrator
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
     *    addCustomColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Any other destructive changes will result in an error when trying to
     * rollback the migration.
     *
     * Remember to call "create()" or "update()" to apply your changes! For example:
     *
     * $this->createTable("users")
     *      ->addColumn("username", "string", array("limit" => 20))
     *      ->addColumn("password", "string", array("limit" => 40))
     *      ->addColumn("password_salt", "string", array("limit" => 40))
     *      ->addColumn("email", "string", array("limit" => 100))
     *      ->addColumn("first_name", "string", array("limit" => 30))
     *      ->addColumn("last_name", "string", array("limit" => 30))
     *      ->addColumn("created", "datetime")
     *      ->addColumn("updated", "datetime", array("null" => true))
     *      ->addIndex(array("username", "email"), array("unique" => true))
     *      ->save();
     */
    public function change()
    {
        $table = $this->table('product_attribute_group');

        // 添加sort字段
        if (!$table->hasColumn('sort')) {
            $table->addColumn('sort', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '排序(数字越小越靠前)',
                'after' => 'multi_language_name_uuid'
            ]);
        }

        $table->update();

        $db = Db::connect(Db::getConfig('default'), true);
        $sources = $db->name('product_attribute_group')->where('sort', 0)->order('create_time asc')->select();
        $sort = 1;
        foreach ($sources as $source) {
            $db->name('product_attribute_group')->where('uuid', $source['uuid'])->update(['sort' => $sort]);
            $sort++;
        }


        $table = $this->table('product_attribute');

        // 添加sort字段
        if (!$table->hasColumn('sort')) {
            $table->addColumn('sort', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '排序(数字越小越靠前)',
                'after' => 'multi_language_name_uuid'
            ]);
        }

        $table->update();

        $db = Db::connect(Db::getConfig('default'), true);
        $sources = $db->name('product_attribute')->where('sort', 0)->order('create_time asc')->select();
        $sort = 1;
        foreach ($sources as $source) {
            $db->name('product_attribute')->where('uuid', $source['uuid'])->update(['sort' => $sort]);
            $sort++;
        }
    }
}

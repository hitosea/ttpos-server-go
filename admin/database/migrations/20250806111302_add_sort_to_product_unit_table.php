<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddSortToProductUnitTable extends Migrator
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
        $table = $this->table('product_unit');

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
        $units = $db->name('product_unit')->where('sort', 0)->order('create_time asc')->select();
        $sort = 1;
        foreach ($units as $unit) {
            $db->name('product_unit')->where('uuid', $unit['uuid'])->update(['sort' => $sort]);
            $sort++;
        }
    }
}

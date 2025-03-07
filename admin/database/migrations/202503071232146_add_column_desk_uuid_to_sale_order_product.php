<?php

use think\migration\Migrator;

class AddColumnDeskUuidToSaleOrderProduct extends Migrator
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
        $table = $this->table('sale_order_product');
        if (!$table->hasColumn('desk_uuid')) {
            $table->addColumn('desk_uuid', 'biginteger', ['default' => 0, 'signed' => false, 'comment' => '桌台ID, 默认为0是本台，大于0为合并过来的桌台', 'after' => 'must_plan_uuid'])
              ->update();
        }
    }
}
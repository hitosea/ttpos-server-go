<?php

use think\migration\Migrator;

class AddSortAndCommitPayTimeFromMemberSaleOrderTable extends Migrator
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
        $table = $this->table('member_sale_order');
        if (!$table->hasColumn('sort')) {
            $table->addColumn('sort', 'integer', ['null' => false, 'default' => 0, 'comment' => '排序, 0-其他状态，1-骑手正在赶往商家，2-骑手配送中，降序排序', 'after' => 'product_amount']);
            $table->update();
        }
    }
} 
<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddTakeoutFieldsToStatisticsSale extends Migrator
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
        $table = $this->table('statistics_sale');
        if (!$table->hasColumn('is_takeout')) {
            $table->addColumn('is_takeout', 'integer', ['default' => 0, 'comment' => '是否外送', 'after' => 'is_special']);
            $table->addIndex(['is_takeout'], ['name' => 'idx_is_takeout']);
            $table->update();
        }

        if (!$table->hasColumn('delivery_fee')) {
            $table->addColumn('delivery_fee', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '配送费', 'after' => 'is_takeout'])
                ->update();
        }
    }
}

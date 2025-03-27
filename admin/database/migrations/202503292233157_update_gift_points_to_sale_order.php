<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateGiftPointsToSaleOrder extends Migrator
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
        $table = $this->table('sale_order');
        if ($table->hasColumn('gift_point') && !$table->hasColumn('gift_points')) {
            $table->renameColumn('gift_point', 'gift_points');
            $table->update();
        }
        if ($table->hasColumn('gift_point_rate') && !$table->hasColumn('gift_points_rate')) {
            $table->renameColumn('gift_point_rate', 'gift_points_rate');
            $table->update();
        }
    }
}

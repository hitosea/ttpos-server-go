<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnDelayTimeToSaleOrderBuffetDelayProduct extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     */
    public function change()
    {
        $table = $this->table('sale_order_buffet_delay_product');
        if (!$table->hasColumn('delay_time')) {
            $table->addColumn('delay_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '加钟时间'])->update();
        }
    }
}
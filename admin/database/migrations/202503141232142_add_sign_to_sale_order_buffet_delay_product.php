<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSignToSaleOrderBuffetDelayProduct extends Migrator
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
        $table = $this->table('sale_order_buffet_delay_product');
        if (!$table->hasColumn('sign')) {
                $table->addColumn('sign', 'string', [
                    'null' => false,
                    'limit' => 255,
                    'comment' => '加钟商品签名。生成uuid,用于标识不同拆单中的加钟商品是不是同一次加购的。在同一个子单中相同签名的加钟商品要合并',
                    'after' => 'buffet_delay_uuid',
                ]);
        }
        $table->update();
    }
}

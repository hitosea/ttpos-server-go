<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFieldsToMarketingCoupon extends Migrator
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
        $table = $this->table('marketing_coupon');
        if (!$table->hasColumn('status')) {
            $table->addColumn('status', 'integer', [
                'null' => false,
                'default' => 1,
                'comment' => '状态 1 开启 0 禁用',
                'after' => 'valid_days',
            ]);
        }
        $table->update();
    }
}

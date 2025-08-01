<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnRiderAvatarAndRiderRatingToMemberSaleOrder extends Migrator
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
        if (!$table->hasColumn('rider_avatar')) {
            $table->addColumn('rider_avatar', 'string', ['null' => false, 'default' => '', 'comment' => '骑手头像', 'after' => 'rider_phone']);
            $table->update();
        }
        if (!$table->hasColumn('rider_rating')) {
            $table->addColumn('rider_rating', 'decimal', ['precision' => 12, 'scale' => 2,  'null' => false, 'default' => 0.00, 'comment' => '骑手评分', 'after' => 'rider_avatar']);
            $table->update();
        }
        if ($table->hasColumn('expected_finish_time')) {
            $table->changeColumn('expected_finish_time', 'string', ['null' => false, 'default' => '', 'comment' => '预计送达时间']);
            $table->update();
        }
    }
}

<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateColumnMemberLevel extends Migrator
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
        $table = $this->table('member_level');
        if ($table->hasColumn('discount')) {
            $table->changeColumn('discount', 'decimal', ['precision' => 10, 'scale' => 2, 'comment' => '等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8 '])->update();
        }

        $table = $this->table('member_card');
        if ($table->hasColumn('discount')) {
            $table->changeColumn('discount', 'decimal', ['precision' => 10, 'scale' => 2, 'comment' => '折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段'])->update();
        }

    }
}

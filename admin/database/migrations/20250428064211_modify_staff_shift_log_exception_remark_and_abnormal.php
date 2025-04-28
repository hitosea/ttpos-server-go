<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ModifyStaffShiftLogExceptionRemarkAndAbnormal extends Migrator
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
        $table = $this->table('staff_shift_log');
        if ($table->hasColumn('exception_remark')) {
            $table->changeColumn('exception_remark', 'string', ['limit' => 500, 'null' => false , 'default' => '',  'comment' => '异常报备' ]);
        }
        if ($table->hasColumn('abnormal')) {
            $table->changeColumn('abnormal', 'text',  ['comment' => '异常信息-json字符串', ]);
        }
        $table->update();
    }
}

<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 任务 #39858, #39627: 班次与ERP解耦
 *
 * 为 ttpos_staff_shift_log 表添加 shift_version 字段
 * 用于区分新旧班次版本：
 * - 1: 旧版（有ERP开关帐，创建Opening/Closing Entry）
 * - 2: 新版（无ERP开关帐，班次管理完全在TTPOS内部完成）
 */
class AddShiftVersionToStaffShiftLog extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        if (!$this->hasTable('staff_shift_log')) {
            return;
        }

        $table = $this->table('staff_shift_log');

        if (!$table->hasColumn('shift_version')) {
            $table->addColumn('shift_version', 'integer', [
                'limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY,
                'null' => false,
                'default' => 2,
                'comment' => '班次版本: 1=旧版(有ERP开关帐) 2=新版(无ERP开关帐)',
                'after' => 'opening_payment_methods',
            ])->update();
        }
    }
}

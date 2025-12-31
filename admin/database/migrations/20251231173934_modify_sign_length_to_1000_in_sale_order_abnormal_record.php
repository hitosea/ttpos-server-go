<?php

use think\migration\Migrator;

class ModifySignLengthTo1000InSaleOrderAbnormalRecord extends Migrator
{
    /**
     * 修改 sale_order_abnormal_record 表的 sign 字段长度为 2000
     */
    public function change()
    {
        $table = $this->table('sale_order_abnormal_record');

        if ($table->hasColumn('sign')) {
            $table->changeColumn('sign', 'string', [
                'limit'   => 2000,
                'null'    => false,
                'default' => '',
                'comment' => '操作签名',
            ])->update();
        }
    }
}


<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddHeadquarterUuidToPaymentMethodTable extends Migrator
{
    /**
     * 添加总部ID字段到 ttpos_payment_method 表
     */
    public function change()
    {
        $table = $this->table('payment_method');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '总部ID', 'after' => 'erpnext_payment'])->update();
        }
    }
}


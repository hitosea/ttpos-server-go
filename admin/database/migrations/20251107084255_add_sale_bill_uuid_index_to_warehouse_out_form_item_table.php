<?php

use think\migration\Migrator;

class AddSaleBillUuidIndexToWarehouseOutFormItemTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-change-method
     *
     * @return void
     */
    public function change()
    {
        $table = $this->table('warehouse_out_form_item');

        // 检查索引是否已经存在
        if ($table->hasIndex(['sale_bill_uuid'])) {
            return;
        }

        // 添加sale_bill_uuid索引
        $table->addIndex(['sale_bill_uuid'], ['name' => 'idx_sale_bill_uuid'])
            ->update();
    }
}

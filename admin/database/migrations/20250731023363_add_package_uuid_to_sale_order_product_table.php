<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddPackageUuidToSaleOrderProductTable extends Migrator
{
    public function change()
    {
        $table = $this->table('sale_order_product');
        if (!$table->hasColumn('package_uuid')) {
            $table->addColumn('package_uuid', 'biginteger', [
                'null' => false,
                'default' => 0,
                'comment' => '套餐uuid',
                'after' => 'is_accept_order'
            ])
            ->update();
        }
    }
} 
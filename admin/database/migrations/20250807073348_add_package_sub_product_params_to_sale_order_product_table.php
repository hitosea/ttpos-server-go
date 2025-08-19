<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class AddPackageSubProductParamsToSaleOrderProductTable extends AbstractMigration
{
    /**
     * 为 sale_order_product 表新增字段 package_sub_product_params
     */
    public function up(): void
    {
        $table = $this->table('sale_order_product');

        // 表不存在则不处理
        if (!$table->exists()) {
            return;
        }

        // 字段不存在则新增
        if (!$table->hasColumn('package_sub_product_params')) {
            $table->addColumn('package_sub_product_params', 'text', [
                'null' => false,
                'comment' => '套餐子商品参数',
                'after' => 'product_type',
            ])->update();
        }
    }

    /**
     * 回滚：删除字段
     */
    public function down(): void
    {
        $table = $this->table('sale_order_product');

        if ($table->exists() && $table->hasColumn('package_sub_product_params')) {
            $table->removeColumn('package_sub_product_params')->update();
        }
    }
}



<?php

use think\migration\Migrator;

class AddAvgToProductionOrderProductTable extends Migrator
{
    /**
     * 添加avg_make_duration、avg_send_duration、avg_all_duration字段到生产订单商品表
     */
    public function change()
    {
        $table = $this->table('production_order_product');


        // 添加avg_make_duration、avg_send_duration、avg_all_duration字段
        if (!$table->hasColumn('avg_make_duration')) {
            $table->addColumn('avg_make_duration', 'decimal', ['null' => true, 'default' => null, 'precision' => 22, 'scale' => 4, 'comment' => '制作时长平均值(秒)', 'after' => 'all_duration'])
                ->update();
        }
        if (!$table->hasColumn('avg_send_duration')) {
            $table->addColumn('avg_send_duration', 'decimal', ['null' => true, 'default' => null, 'precision' => 22, 'scale' => 4, 'comment' => '传菜时长平均值(秒)', 'after' => 'avg_make_duration'])
                ->update();
        }
        if (!$table->hasColumn('avg_all_duration')) {
            $table->addColumn('avg_all_duration', 'decimal', ['null' => true, 'default' => null, 'precision' => 22, 'scale' => 4, 'comment' => '总时长平均值(秒)', 'after' => 'avg_send_duration'])
                ->update();
        }
    }
}

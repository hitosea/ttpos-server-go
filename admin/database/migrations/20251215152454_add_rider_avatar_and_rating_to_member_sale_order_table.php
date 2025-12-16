<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRiderAvatarAndRatingToMemberSaleOrderTable extends Migrator
{
    /**
     * 添加骑手头像和评分字段到 ttpos_member_sale_order 表
     */
    public function change()
    {
        $table = $this->table('member_sale_order');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('rider_avatar')) {
            $table->addColumn('rider_avatar', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '骑手头像', 'after' => 'rider_phone'])->update();
        }

        if (!$table->hasColumn('rider_rating')) {
            $table->addColumn('rider_rating', 'decimal', ['precision' => 20, 'scale' => 4, 'null' => false, 'default' => 0.0000, 'comment' => '骑手评分', 'after' => 'rider_avatar'])->update();
        }
    }
}

<?php
/**
 * 创建总部字段覆盖追踪表
 *
 * 用于记录子店对总部来源商品/物品可修改字段的覆盖情况
 * 当子店修改了可控字段时，记录override，总部推送时根据此记录决定是否覆盖
 */

use think\migration\Migrator;

class CreateHqFieldOverride extends Migrator
{
    // 无 TARGET：仅应用到所有商户数据库

    public function change()
    {
        if ($this->hasTable('hq_field_override')) {
            return;
        }

        $table = $this->table('hq_field_override', ['comment' => '总部字段覆盖追踪表']);
        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
            ->addColumn('entity_type', 'string', ['limit' => 50, 'default' => '', 'comment' => '实体类型: product/product_takeout/material'])
            ->addColumn('entity_uuid', 'biginteger', ['default' => 0, 'comment' => '实体UUID'])
            ->addColumn('field_type', 'string', ['limit' => 50, 'default' => '', 'comment' => '字段类型: dine_shelf/takeout_shelf/takeout_price/safety_stock/negative_stock'])
            ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true])
            ->addIndex(['entity_uuid', 'field_type', 'delete_time'], ['unique' => true, 'name' => 'uk_entity_field'])
            ->addIndex(['entity_type'], ['name' => 'idx_entity_type'])
            ->create();
    }
}

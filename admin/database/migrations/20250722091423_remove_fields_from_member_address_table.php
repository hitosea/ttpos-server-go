<?php

use think\migration\Migrator;

class RemoveFieldsFromMemberAddressTable extends Migrator
{
    // 迁移目标
    const TARGET = 'shop_master';
    
    /**
     * 删除会员地址表中的字段
     */
    public function change()
    {
        try {
            $table = $this->table('member_address');
            if (!$table->hasColumn('phone_prefix')) {
                $table->addColumn('phone_prefix', 'string', ['limit' => 10, 'null' => false, 'default' => '+66', 'comment' => '手机区号', 'after' => 'phone']);
                $table->update();
            }
            // 检查字段是否存在，如果存在则删除
            if ($table->hasColumn('gender')) {
                $table->removeColumn('gender');
            }
            if ($table->hasColumn('province')) {
                $table->removeColumn('province');
            }
            if ($table->hasColumn('city')) {
                $table->removeColumn('city');
            }
            if ($table->hasColumn('area')) {
                $table->removeColumn('area');
            }
            if ($table->hasColumn('country')) {
                $table->removeColumn('country');
            }
            $table->save();
        } catch (\Throwable $th) {
            //throw $th;
        }
    }
} 
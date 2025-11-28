<?php

use think\facade\Db;
use think\migration\Migrator;

class UpdateBomCardAccessSort extends Migrator
{
    /**
     * 更新成本卡相关权限的排序和信息
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);
        
        // 更新商品成本卡的权限顺序
        // 原来：添加(sort=1) 编辑(sort=2)
        // 修改为：编辑(sort=1) 删除(sort=2)
        
        // 1. uuid=2858233630720000 编辑 sort改为1
        $db->name('access')
            ->where('uuid', '=', 2858233630720000)
            ->update([
                'sort' => 1,
                'update_time' => time()
            ]);
        
        // 2. uuid=2858216853504000 添加改为删除 sort改为2 path改为bom_card_product_delete
        $db->name('access')
            ->where('uuid', '=', 2858216853504000)
            ->update([
                'name' => '删除',
                'path' => 'bom_card_product_delete',
                'sort' => 2,
                'update_time' => time()
            ]);
        
        // 更新加料成本卡的权限顺序
        // 原来：添加(sort=1) 编辑(sort=2)
        // 修改为：编辑(sort=1) 删除(sort=2)
        
        // 3. uuid=2858296545280000 编辑 sort改为1
        $db->name('access')
            ->where('uuid', '=', 2858296545280000)
            ->update([
                'sort' => 1,
                'update_time' => time()
            ]);
        
        // 4. uuid=2858283962368000 添加改为删除 sort改为2 path改为bom_card_sauce_delete
        $db->name('access')
            ->where('uuid', '=', 2858283962368000)
            ->update([
                'name' => '删除',
                'path' => 'bom_card_sauce_delete',
                'sort' => 2,
                'update_time' => time()
            ]);
    }
}

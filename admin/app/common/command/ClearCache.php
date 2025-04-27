<?php

declare(strict_types=1);

namespace app\common\command;

use think\facade\Cache;
use think\console\Input;
use think\console\Output;
use think\console\Command;
use app\common\model\shop\User;
use app\common\enum\settings\SettingEnum;

// 清空缓存
// ./cmd think clear-cache
class ClearCache extends Command
{
    protected function configure()
    {
        // 指令配置
        $this->setName('clear-cache')->setDescription('清空所有缓存');
    }

    protected function execute(Input $input, Output $output)
    {
        Cache::tag('cache')->clear();
        Cache::tag('firstshop')->clear();
        Cache::tag('common_get_settingLanguages')->clear();
        Cache::set('sync_setting_' . SettingEnum::CLOUD_BASIC, null);
        //
        $shop_supplier_id = User::getShopInfo('shop_supplier_id');
        Cache::tag('category' . $shop_supplier_id . '0' . '0')->clear();
        Cache::tag('category' . $shop_supplier_id . '0' . '1')->clear();
        Cache::tag('category' . $shop_supplier_id . '1' . '0')->clear();
        Cache::tag('category' . $shop_supplier_id . '1' . '1')->clear();
        //
        Cache::set('__SYNC_GET_PUBLICKEY_', 0);
    }
}

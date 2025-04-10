<?php

declare(strict_types=1);

namespace app\common\command;

use think\facade\Db;
use think\facade\Env;
use think\facade\Cache;
use think\console\Input;
use think\facade\Config;
use think\console\Output;
use think\console\Command;
use app\common\enum\settings\SettingEnum;

// 清空所有订单关联数据
// ./cmd think reload-order

class ReloadOrder extends Command
{
    protected function configure()
    {
        // 指令配置
        $this->setName('reload-order')->setDescription('清空所有订单关联数据');
    }

    protected function execute(Input $input, Output $output)
    {
        $output->writeln("请输入商家id");
        $appId = trim($output->ask($input, ''));
        //
        $default = config('database.default');
        $config = config('database');
        $mysql = $config['connections'][$default];
        $mysql['database'] = 'shop'.$appId;
        $mysql['username'] = env('DB_USERNAME');
        $mysql['password'] = env('DB_ROOT_PASSWORD');
        $config['connections'][$default] = $mysql;
        Config::set($config, 'database');
        //
        $db = Db::connect(Db::getConfig('default'), true);
        //
        $supplierName = $db->name('supplier')->value('name');
        $shopSupplierId = $db->name('supplier')->value('shop_supplier_id');
        if (!$supplierName) {
            $output->writeln("商家id不存在!");
        }
        //
        $random = mt_rand(100000, 999999);
        $output->writeln("<info>商家：" . $supplierName . "($appId)</info>");
        $output->writeln("<fg=red>将会清空所有订单,会员余额记录,会员积分记录,操作记录,钱箱金额,确认请输入确认码: {$random}</>");
        $confirmationCode = trim($output->ask($input, ''));
        if ($confirmationCode == $random) {
            Cache::set("__RELOADAYNC__", 1, 600);
            // 清表
            $prefix = Env::get('DB_PREFIX');
            /**
             * 订单相关表
             */
            foreach ([
                'order',
                'order_product',
                'order_product_return',
                'order_address',
                'order_buffet',
                'order_buffet_customer',
                'order_buffet_discount',
                'order_delay',
                'order_extract',
                'order_finance',
                'order_pay_type',
                'order_peak_time',
                'order_settled',
                'order_refund',
                'order_free',
                'order_product_free',
                'order_product_return',
                'order_refund_destination',
                'order_operation_log',
                'order_abnormal_log',
                //
                'user_balance_log',
                'user_recharge_order',
                'user_recharge_order_operation_log',
                'user_recharge_order_pay_type',
                'user_recharge_order_refund',
                'user_recharge_order_refund_destination',
                'shop_opt_log',
                'shop_account_log',
                'printer_log',
                'printer_read_log',
                'user_points_log',
            ] as $tablename) {
                $db->execute("TRUNCATE TABLE `{$prefix}{$tablename}`");
            }
            // 还原桌台状态
            $db->execute("UPDATE `{$prefix}table` SET `status` = 10");
            // 归零钱箱
            $db->execute( "UPDATE `{$prefix}shop_account` SET `amount` = 0");
            //
            Cache::tag('cache')->clear();
            Cache::tag('firstshop')->clear();
            Cache::tag('common_get_settingLanguages')->clear();
            Cache::set('sync_setting_' . SettingEnum::CLOUD_BASIC, null);
            Cache::tag('category' . $shopSupplierId . '0' . '0');
            Cache::tag('category' . $shopSupplierId . '0' . '1');
            Cache::tag('category' . $shopSupplierId . '1' . '0');
            Cache::tag('category' . $shopSupplierId . '1' . '1');
            Cache::set('__SYNC_GET_PUBLICKEY_', 0);
            //
            Cache::set("__RELOADAYNC__", 0);
        } else {
            $output->writeln("错误的确认码。退出指令执行。");
            return;
        }
        //
        $output->writeln('#####完成#####');
    }
}

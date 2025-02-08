<?php

declare(strict_types=1);

namespace app\common\command;

use think\facade\Db;
use think\facade\Env;
use think\console\Input;
use think\facade\Config;
use think\console\Output;
use think\console\Command;

// 修复订单支付类型
// ./cmd think fix-order-pay-type

class FixOrderPayType extends Command
{
    protected function configure()
    {
        // 指令配置
        $this->setName('fix-order-pay-type')->setDescription('修复订单支付类型');
    }

    protected function execute(Input $input, Output $output)
    {
        $output->writeln("请输入商家id");
        $appId = trim($output->ask($input, ''));
        //
        $default = config('database.default');
        $config = config('database');
        $mysql = $config['connections'][$default];
        $mysql['database'] = 'shop' . $appId;
        $mysql['username'] = 'root';
        $mysql['password'] = env('DB_ROOT_PASSWORD');
        $config['connections'][$default] = $mysql;
        Config::set($config, 'database');
        //
        $db = Db::connect(Db::getConfig('default'), true);
        //
        $supplierName = $db->name('supplier')->where('company_uuid', $appId)->value('name');
        if (!$supplierName) {
            $output->writeln("商家id不存在!");
        }
        //
        $random = mt_rand(100000, 999999);
        $output->writeln("<info>商家：" . $supplierName . "($appId)</info>");
        $output->writeln("<fg=red>将会查询所有重复的现金支付记录进行修复 (包含订单,交班单,统计报表); 因业务影响问题“钱箱金额”不会修复, 需自行协商解决</>");
        $output->writeln("<fg=red>确认修复, 请输入确认码: {$random}</>");
        $confirmationCode = trim($output->ask($input, ''));
        if ($confirmationCode == $random) {
            //
            $orderPayTypes = $db->name('order_pay_type')
                ->field('*, count(*) as count_num')
                ->group('order_id,value,price')
                ->having('count(*) > 1')
                ->where('value', 40)
                ->select();
            if (!$orderPayTypes) {
                $output->writeln("不存在重复的现金支付记录");
                return;
            }
            $output->writeln("存在重复的现金支付记录 " . count($orderPayTypes) . " 条, 正在修复...");
            // 开启事务
            $db->startTrans();
            try {
                // Your code that modifies the database goes here
                foreach ($orderPayTypes as $key => $orderPayType) {
                    $index = $key + 1;
                    $price = $orderPayType['price'];
                    //
                    $output->writeln("修复条目：$index  订单ID：" . $orderPayType['order_id'] . " 支付金额：" . $price);
                    //
                    $order = $db->name('order')
                        ->field('order_id,cashier_id,create_time,actual_price,pay_price,change_due,pay_status')
                        ->where('order_id', $orderPayType['order_id'])
                        ->where('pay_status', 20)
                        ->find();
                    //
                    if ($order && $price) {
                        // 修复订单
                        $db->name('order')->where('order_id', $orderPayType['order_id'])->update([
                            'actual_price' => Db::raw('actual_price - ' . $price),
                            'change_due' => Db::raw('actual_price - pay_price'),
                        ]);
                        $newOrder = $db->name('order')
                            ->field('order_id,cashier_id,create_time,actual_price,pay_price,change_due')
                            ->where('order_id', $orderPayType['order_id'])
                            ->find();

                        // 修复交班单
                        $shiftLog = $db->name('shop_user_shift_log')
                            ->where('shift_user_id', $order['cashier_id'])
                            ->where('shift_start_time', '<=', $order['create_time'])
                            ->where('shift_end_time', '>=', $order['create_time'])
                            ->select();
                        if (count($shiftLog) > 1) {
                            $output->writeln("<fg=red>存在多条交班记录，无法进行修复, 请人工处理</>");
                            return;
                        }
                        if ($shiftLog[0]) {
                            // 原记录的收入
                            $oldRecordedIncome = ($orderPayType['price'] - $order['change_due']) * 2;
                            // 应记录的收入
                            $recordedIncome = $orderPayType['price'] - $newOrder['change_due'];
                            // 差值
                            $difference = $recordedIncome - $oldRecordedIncome;
                            //
                            $output->writeln("修复差值：" . $difference);
                            //
                            $cashIncome = false;
                            $incomes = json_decode($shiftLog[0]['incomes'], true);
                            foreach ($incomes as $key => $income) {
                                if ($income['pay_type'] == 40) {
                                    $incomes[$key]['price'] = $cashIncome = $incomes[$key]['price'] + $difference;
                                }
                            }
                            if ($cashIncome) {
                                $db->name('shop_user_shift_log')->where('id', $shiftLog[0]['id'])->update([
                                    'cash_income' => $cashIncome,
                                    'incomes' => json_encode($incomes),
                                ]);
                                // 交班快照
                                $snapshot = $db->name('shop_user_shift_snapshot')->where('shift_log_id', $shiftLog[0]['id'])->find();
                                if ($snapshot) {
                                    $content = json_decode($snapshot['content'], true);
                                    $content['incomes'] = $incomes;
                                    $content['cash_income'] = $cashIncome;
                                    $content['order']['incomes'] = $incomes;
                                    $db->name('shop_user_shift_snapshot')->where('shift_log_id', $shiftLog[0]['id'])->update([
                                        'content' => json_encode($content),
                                    ]);
                                }
                            }
                        }

                        // 修复订单支付类型
                        $db->name('order_pay_type')
                            ->where('id', $orderPayType['id'])
                            ->where('order_id', $orderPayType['order_id'])
                            ->delete();
                    }
                }
                // 提交事务
                $db->commit();
            } catch (\Exception $e) {
                // 回滚事务
                $db->rollback();
                $output->writeln("修复失败: " . $e->getMessage());
            }
        } else {
            $output->writeln("错误的确认码。退出指令执行。");
            return;
        }
        //
        $output->writeln('#####完成#####');
    }
}

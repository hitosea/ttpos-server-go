<?php

namespace app\common\model\shop;

use help\QueueHelp;
use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use app\common\exception\BaseException;

/**
 * 店铺金额表
 */
class Account extends BaseModel
{
    protected $name = 'shop_account';
    protected $pk = 'id';

    /**
     * 获取金额
     * 
     * @param User|null $shopUser
     * @param float $current_cash_income
     * @return float
     * @throws BaseException
     */
    public static function getAmount(User $shopUser = null, $current_cash_income = 0)
    {
        $amount = 0;
        if ($account = Account::order('id', 'desc')->find()) {
            $amount = $account->amount;
        } else {
            $userShiftLog = UserShiftLog::order('id', 'desc')->where('status', 1)->find();
            if (!$userShiftLog && !$shopUser) {
                throw new BaseException(['msg' => '获取金额错误', 'code' => 0]);
            }
            if ($userShiftLog) {
                $amount = helper::bcadd($userShiftLog->cash_left ?: 0, $current_cash_income);
                $shopSupplierId = $userShiftLog->shop_supplier_id;
                $appId = $userShiftLog->app_id;
            } else {
                $amount = 0;
                $shopSupplierId = $shopUser->shop_supplier_id;
                $appId = $shopUser->app_id;
            }
            Account::addAccount($amount, $shopSupplierId, $appId);
        }
        return $amount;
    }

    /**
     * 更新金额
     * 
     * @param int $pm
     * @param float $amount
     * @param int $dutyNo
     * @param int $cashierId
     * @param int $shopSupplierId
     * @param int $appId
     * @param string $source
     * @return Account
     */
    public static function updateAmount(int $pm, float $amount, string|int $dutyNo, int $cashierId, int $shopSupplierId, int $appId, $source = 'shift')
    {
        if ($amount <= 0) {
            return;
        }
        // 禁止并发操作
        $queue = new QueueHelp('ACCOUNTLOG-APPLOG' . $shopSupplierId);
        $queue->while();
        //
        $account = self::order('id', 'desc')->find();
        $currentAmount = $account->amount ?: 0;
        $preAmount = $currentAmount;
        $afterAmount = 0;
        if ($pm == 0) {
            $afterAmount = $currentAmount = helper::bcsub($currentAmount, $amount);
        } else {
            $afterAmount = $currentAmount = helper::bcadd($currentAmount, $amount);
        }
        // 更新账户余额
        if ($account) {
            $account->amount = $currentAmount;
            $account->save();
        } else {
            $account = self::addAccount($currentAmount, $shopSupplierId, $appId);
        }
        // 记录变更日志
        AccountLog::create([
            'pm' => $pm,
            'amount' => $amount,
            'pre_amount' => $preAmount,
            'after_amount' => $afterAmount,
            'duty_no' => $dutyNo,
            'cashier_id' => $cashierId,
            'shop_supplier_id' => $shopSupplierId,
            'app_id' => $appId,
            'source' => $source,
        ]);
        //
        $queue->release();
        //
        return $account;
    }

    /**
     * 添加账号
     */
    public static function addAccount($amount, $shopSupplierId, $appId)
    {
        return self::create([
            'amount' => $amount,
            'shop_supplier_id' => $shopSupplierId,
            'app_id' => $appId,
            'create_time' => time(),
            'update_time' => time()
        ]);
    }
}

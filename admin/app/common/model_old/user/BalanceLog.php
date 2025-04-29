<?php

namespace app\common\model_old\user;

use app\common\enum\settings\SettingEnum;
use app\common\library\helper;
use app\common\model_old\BaseModel;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum;
use app\common\model_old\settings\Setting as SettingModel;
use think\facade\Log;

/**
 * 用户余额变动明细模型
 */
class BalanceLog extends BaseModel
{
    protected $name = 'user_balance_log';
    protected $updateTime = false;

    const VERSION = '1.0.8';        // 当前版本 - 区分数据来源版本，根据需要修改

    const VERSION_107 = '1.0.7';    // 区分数据来源版本，根据需要修改

    /**
     * 获取当前模型属性
     */
    public static function getAttributes()
    {
        return [
            // 充值方式
            'scene' => BalanceLogSceneEnum::data(),
        ];
    }

    /**
     * 关联会员记录表
     */
    public function user()
    {
        $module = self::getCalledModule() ?: 'common';
        return $this->belongsTo("app\\{$module}\\model\\user\\User");
    }

    /**
     * 余额变动场景
     */
    public function getSceneAttr($value)
    {
        return ['text' => BalanceLogSceneEnum::data()[$value]['name'], 'value' => $value];
    }

    /**
     * 新增记录
     * @param $scene  // 余额变动场景 10-用户充值 20-用户消费 30-管理员操作 40-订单退款 60-反结账
     * @param $data
     * @param $describeParam
     * @return void
     */
    public static function add($scene, $data, $describeParam, $custom_dec = false)
    {
        //
        $model = new static;
        $setting = SettingModel::getSupplierItem(SettingEnum::POINTS, $model::$app_id, $model::$app_id);
       
        //
        $order_id = isset($data['order_id'])  ? $data['order_id'] : 0;
        $user_id = isset($data['user_id'])  ? $data['user_id'] : 0;
        $money = isset($data['money'])  ? $data['money'] : 0;
        $gift_money = isset($data['gift_money'])  ? $data['gift_money'] : 0;
        // 记录余额变更
        $user = User::where('user_id', $user_id)->find();
        if ($user) {
            $user_balance = max($user['balance'], 0);
            $user_gift_balance = max($user['gift_balance'], 0);
            $before_money = $user['balance'] + $user['gift_balance'];
            if ($money < 0 || ($scene == BalanceLogSceneEnum::DEDUCT && $gift_money < 0)) {
                $abs_money = abs($money);
                /**
                 * 减少
                 */
                $after_money = helper::bcadd($before_money, $money); //
                // 扣除类型
                $deduct_money_type = isset($setting['deduct_order']) ? $setting['deduct_order'] : 1;
                if (($scene == BalanceLogSceneEnum::ADMIN || $scene == BalanceLogSceneEnum::DEDUCT) && $gift_money < 0) {
                    $user->where('user_id', $user_id)->dec('gift_balance', abs($gift_money))->update();
                    $data['gift_money'] = -abs($gift_money); // 余额日志记录赠送金额
                }
                // 用户充值订单退款（退款时不扣减赠送金额跟赠送积分，前去商家后台处理）
                else if ($scene == BalanceLogSceneEnum::RECHARGE_REFUND) {
                    $user->where('user_id', $user_id)->dec('balance', $abs_money)->update();
                }
                // 用户充值订单反结账
                else if ($scene == BalanceLogSceneEnum::RECHARGE_REVERSE) {
                    // 反结账 abs_money 包含赠送金额，所以直接扣除
                    $giftMoney = abs($gift_money);
                    $absBalance = $abs_money - $giftMoney; // 余额
                    $user->where('user_id', $user_id)->dec('balance', $absBalance)->dec('gift_balance', $giftMoney)->update();
                }
                // 订单日志操作
                else {
                    // 先赠后主
                    if ($deduct_money_type == 2) {
                        // 赠送余额
                        $remaining_money = $user_gift_balance - $abs_money;
                        // 赠送余额
                        if ($remaining_money >= 0) {
                            $user->where('user_id', $user_id)->dec('gift_balance', $abs_money)->update();
                            $data['gift_money'] = -abs($abs_money); // 余额日志记录赠送金额
                        } else {
                            // 先用赠送余额扣除
                            $user->where('user_id', $user_id)->dec('gift_balance', abs($user_gift_balance))->update();
                            $data['gift_money'] = -abs($user_gift_balance); // 余额日志记录赠送金额
                            // 再用主余额扣除剩下的
                            $user->where('user_id', $user_id)->dec('balance', abs($remaining_money))->update();
                        }
                    }
                    // 主赠按比例
                    elseif ($deduct_money_type == 3) {
                        $main_ratio = round($setting['deduct_ratio_main'] / 100, 4);
                        $main_pay_money = helper::bcmul($abs_money, $main_ratio, 5);
                        // 主赠余额应扣
                        $main_pay_money = round($main_pay_money, 2);
                        $gift_pay_money = $abs_money - $main_pay_money;
                        //
                        $main_remaining_money = $user_balance - $main_pay_money; //
                        $gift_remaining_money = $user_gift_balance - $gift_pay_money; //
                        // 主赠余额都足够
                        if ($main_remaining_money >= 0 && $gift_remaining_money >= 0) {
                            $user->where('user_id', $user_id)->dec('balance', abs($main_pay_money))->update();
                            $user->where('user_id', $user_id)->dec('gift_balance', abs($gift_pay_money))->update();
                            $data['gift_money'] = -abs($gift_pay_money); // 余额日志记录赠送金额
                        }
                        // 主够赠不够
                        else if ($main_remaining_money >= 0 && $gift_remaining_money < 0) {
                            // 先把赠的扣完
                            $user->where('user_id', $user_id)->dec('gift_balance', abs($user_gift_balance))->update();
                            $data['gift_money'] = -abs($user_gift_balance); // 余额日志记录赠送金额
                            // 再用主扣除所有
                            $user->where('user_id', $user_id)->dec('balance', $main_pay_money + abs($gift_remaining_money))->update();
                        }
                        // 赠够主不够
                        else if ($main_remaining_money < 0 && $gift_remaining_money >= 0) {
                            // 先把主的扣完
                            $user->where('user_id', $user_id)->dec('balance', abs($user_balance))->update();
                            // 再用赠扣除所有
                            $change_gift_balance = $gift_pay_money + abs($main_remaining_money);
                            $user->where('user_id', $user_id)->dec('gift_balance', $change_gift_balance)->update();
                            $data['gift_money'] = -abs($change_gift_balance); // 余额日志记录赠送金额
                        } else {
                            $user->where('user_id', $user_id)->dec('balance', $abs_money)->update();
                        }
                    }
                    // 先主后赠
                    else {
                        // 主余额
                        $remaining_money = $user_balance - $abs_money;
                        // 主余额是否足够
                        if ($remaining_money >= 0) {
                            $user->where('user_id', $user_id)->dec('balance', $abs_money)->update();
                        } else {
                            // 先用主余额扣除
                            $user->where('user_id', $user_id)->dec('balance', abs($user_balance))->update();
                            // 再用赠送余额扣除剩下的
                            $user->where('user_id', $user_id)->dec('gift_balance', abs($remaining_money))->update();
                            $data['gift_money'] = -abs($remaining_money); // 余额日志记录赠送金额
                        }
                    }
                }
            } else {
                /**
                 * 增加
                 */
                $after_money = helper::bcadd($before_money, $money);
                // 订单余额退款都退回赠送余额
                if ($scene == BalanceLogSceneEnum::REFUND) {
                    // 赠送余额
                    $user->where('user_id', $user_id)->inc('gift_balance', abs($money))->update();
                    $data['gift_money'] = abs($money); // 余额日志记录赠送金额
                }
                // 反结账余额原路返回还
                else if ($scene == BalanceLogSceneEnum::REVERSE) {
                    $log = BalanceLog::where('scene', BalanceLogSceneEnum::CONSUME)->where('order_id', $order_id)->order(['log_id' => 'desc'])->find();
                    if ($log) {
                        // 主余额
                        $user->where('user_id', $user_id)->inc('balance', abs($log->money - $log->gift_money))->update();
                        // 赠送余额
                        $user->where('user_id', $user_id)->inc('gift_balance', abs($log->gift_money))->update();
                        $data['gift_money'] = abs($log->gift_money); // 余额日志记录赠送金额
                    } else {
                        // 主余额
                        $user->where('user_id', $user_id)->inc('balance', abs($money - $gift_money))->update();
                        // 赠送余额
                        $user->where('user_id', $user_id)->inc('gift_balance', abs($gift_money))->update();
                        $data['gift_money'] = abs($gift_money); // 余额日志记录赠送金额
                    }
                }
                //
                else {
                    // 主余额
                    $user->where('user_id', $user_id)->inc('balance', abs($money - $gift_money))->update();
                    // 赠送余额
                    $user->where('user_id', $user_id)->inc('gift_balance', abs($gift_money))->update();
                }
            }
            $data['before_money'] = $before_money;
            $data['after_money'] = $after_money;
        }
       
        $model->save(array_merge([
            'scene' => $scene,
            'describe' => $custom_dec ? $describeParam : vsprintf(BalanceLogSceneEnum::data()[$scene]['describe'], $describeParam),
            'app_id' => $model::$app_id,
            'version' => self::VERSION
        ], $data));
    }
}

<?php

namespace app\job\event;

use help\HttpHelp;
use app\job\model\user\User as UserModel;

class UserRecharge
{
    /**
     * 执行函数
     */
    public function handle($data)
    {
        $uuid = $data['uuid'];
        $balance = $data['balance'];
        $points = $data['points'];
        if ($balance['mode'] != 'inc') {
            return true;
        }
        $bonusPoints = 0;
        if ($points['mode'] == 'inc') {
            $bonusPoints = $points['value'];
        }

        // 查询当前会员余额和积分
        $user = UserModel::where('uuid', $uuid)->find();
        $curBalance = $user['balance'];
        $curPoints = $user['points'];

        $parems = [
            'company' => '', // 公司名称
            'phone' => $user['phone'], // 会员手机号
            'recharge' => floatval($balance['money']), // 充值金额
            'bonus_money' => floatval($balance['gift_balance']), // 赠送金额
            'bonus_points' => floatval($bonusPoints), // 赠送积分
            'balance' => floatval($curBalance), // 会员当前余额
            'points_balance' => floatval($curPoints), // 会员当前积分
        ];

        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/sms/member-recharge', json_encode($parems), [
            'Authorization: Bearer ' . $data['token'],
            'Accept-Language: ' . $data['language'],
            'Content-Type: application/json; charset=utf-8',
        ]);

        if (!$res) {
            $parems['error'] = '请求失败';
            $this->dologs('sms/member-recharge', $parems);
            return true;
        }
        $res = json_decode($res, true);
        if (($res['code'] ?? -1) != 0) {
            $parems['error'] = $res['message'];
            $this->dologs('sms/member-recharge', $parems);
            return true;
        }

        return true;
    }

    /**
     * 记录日志
     */
    private function dologs($method, $params = [])
    {
        $value = 'UserRecharge --' . $method;
        foreach ($params as $key => $val)
            $value .= ' --' . $key . ' ' . $val;
        return log_write($value, 'task');
    }
}
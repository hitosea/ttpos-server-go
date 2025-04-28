<?php

namespace app\cashier\model\cashier;

use app\common\model\shop\UserShiftLog;
use app\common\model\shop\User as UserModel;
use app\common\model\shop\Access as AccessModel;
use app\common\model\shop\BindRecord as BindRecordModel;

/**
 * 收银员登录模型
 */
class User extends UserModel
{
    /**
     *检查登录
     */
    public function checkLogin($user)
    {
        $device_id = ($user['key'] ?? '') ?: ($user['device_id'] ?? '');
        $brand = $user['brand'] ?? '';
        //
        $username = $user['user_name'] ?? '';
        $password = $user['password'] ?? '';
        //
        $user = self::setAppId(0)->withoutGlobalScope()->where(function ($q) use ($username) {
            $q->whereRaw('BINARY user_name = :user_name', ['user_name' => $username])->whereOr('phone', $username);
        })->where('password', $password)->order('shop_user_id', 'desc')->with(['app', 'supplier'])->find();
        if (!$user) {
            $this->error = '账号或密码错误';
            return false;
        }
        if ($user['is_delete'] == 1) {
            $this->error = '账号被删除，请联系管理员';
            return false;
        }
        if ($user['is_status'] == 1) {
            $this->error = '账号被禁用，请联系管理员';
            return false;
        }
        if (empty($user['app'])) {
            $this->error = '未找到绑定的商家，请确认登录信息';
            return false;
        }
        if ($user['app']['is_delete']) {
            $this->error = '未找到绑定的商家，请确认登录信息';
            return false;
        }
        if ($user['app']['is_recycle']) {
            $this->error = '商家账号异常，请联系管理员';
            return false;
        }
        //
        request()->appId = $user['app_id'];
        request()->shopSupplierId = $user['shop_supplier_id'];
        // 验证权限
        $permission = (new AccessModel([], $user->app_id))->getPermission(AccessModel::CASHIER_ROUTE_NAME, $user, $user['supplier']);
        if (empty($permission)) {
            $this->error = '当前无权限，请联系管理员';
            return false;
        }
        // 检查是否有未交班的收银员 - 当前收银机有收银员未交班，不能登录别的收银账号
        $curCashierUser = self::setAppId($user->app_id)->where(['cashier_online' => 1])->where(['bind_key' => $device_id])->find();
        if ($curCashierUser && $curCashierUser['shop_user_id'] != $user['shop_user_id']) {
            $this->error = __('当前收银机上有未交班的账号，请联系') . '[' . $curCashierUser['real_name'] . ']' . __('完成交班后再登录');
            return false;
        }
        // 是否是首次接班 1是 0否
        $user['is_first_login'] = $user['cashier_online'] == 0 ? 1 : 0;
        // 是否已在其他收银机登录 - 同一个收银员未交班是不能同时登陆两个收银机
        $model = new BindRecordModel;
        $bindInfo = $model->where(['source' => BindRecordModel::SOURCE_CASHIER, 'finally_login_id' => $user['shop_user_id']])->where('key', '<>', $device_id)->find();
        if (($user['cashier_online'] == 1 && $user['bind_key'] != $device_id) || $bindInfo) {
            $this->error = __('收银员') . '[' . ($user['real_name'] ?: $user['user_name']) . ']' . __('已在其他收银机登录未交班，请先完成交班操作');
            return false;
        }
        // 权限
        request()->licenses = $user->license = $user['app']->getLicense();
        // 绑定设备 先检测是否能进行绑定，避免先更新用户表信息再弹出绑定错误
        $data = [
            'key' => $device_id,
            'brand' => $brand,
            'source' => BindRecordModel::SOURCE_CASHIER,
            'finally_login_id' => $user['shop_user_id'],
            'finally_login_time' => time(),
            'app_id' => $user['app_id'],
            'shop_supplier_id' => $user['shop_supplier_id'],
        ];
        if (!$model->add($data, $user->license)) {
            $this->error = $model->getError() ?: __('绑定失败');
            return false;
        }
        // 更新收银员信息 - 主平台和商户
        $userData = ['cashier_online' => 1, 'bind_key' => $device_id];
        if ($user['cashier_login_time'] == 0 || $user['cashier_online'] == 0) {
            $workingLog = (new UserShiftLog)->createWorkingLog($user);
            $userData['cashier_login_time'] = $workingLog->shift_start_time;
            $userData['duty_no'] = $workingLog->shift_no;
        }
        $user->save($userData);
        (new self)->where('shop_user_id', $user->shop_user_id)->find()?->save($userData);
        // 保存登录状态
        $user['token'] = signToken($user['shop_user_id'], 'cashier', $device_id, md5($user->password));
        //
        return $user;
    }


    /*
    * 修改密码
    */
    public function editPass($data, $user)
    {
        $user_info = User::detail($user['cashier_id']);
        if (!$user_info) {
            $this->error = '用户不存在';
            return false;
        }
        if ($data['password'] != $data['confirmPass']) {
            $this->error = '新密码输入不一致';
            return false;
        }
        if ($user_info['password'] != salt_hash($data['oldpass'])) {
            $this->error = '原始密码错误';
            return false;
        }
        $date['password'] = salt_hash($data['password']);
        $user_info->save($date);
        //
        (new User([], 0))->where('shop_user_id', $user_info->shop_user_id)->find()?->save($date);
        //
        return true;
    }

    /**
     * 获取用户信息
     */
    public static function getUser($data)
    {
        return self::where(['shop_user_id' => $data['uid']])->with(['app'])->find();
    }
}

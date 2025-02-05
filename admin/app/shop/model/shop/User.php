<?php

namespace app\shop\model\shop;

use help\ValidateHelp;
use app\common\model\shop\User as UserModel;
use app\common\model\shop\Access as AccessModel;
use app\common\model\shop\LoginLog as LoginLogModel;
use app\shop\model\settings\Setting as SettingModel;

/**
 * 后台管理员登录模型
 */
class User extends UserModel
{
    /**
     *检查登录
     */
    public function checkLogin($user)
    {
        $username = $user['username'] ?? '';
        $password = $user['password'] ?? '';
        $user = $this->where(function ($q) use ($username) {
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
        //
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
        request()->license = $user->license = $user['app']->getLicense();
        // 验证权限
        $permission = (new AccessModel)->getPermission(AccessModel::SHOP_ROUTE_NAME, $user, $user['supplier']);
        if (empty($permission)) {
            $this->error = '当前无权限，请联系管理员';
            return false;
        }
        // 保存登录状态
        $user['token'] = signToken($user['shop_user_id'], 'shop', '', md5($user->password));
        // 货币信息
        $user['currency'] = SettingModel::getCurrency($user['shop_supplier_id'], $user['app_id']);
        // 写入登录日志
        LoginLogModel::add($username, \request()->ip(), '登录成功', $user['app']['app_id']);
        return $user;
    }


    /*
    * 修改密码
    */
    public function editPass($data, $user)
    {
        if (!ValidateHelp::validateAlphaPassword($data['password'])) {
            $this->error = '不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种';
            return false;
        }
        $user_info = User::detail($user['shop_user_id']);
        if ($user_info['password'] != salt_hash($data['oldpass'])) {
            $this->error = '原密码错误';
            return false;
        }
        if ($data['password'] != $data['confirmPass']) {
            $this->error = '两次密码输入不一致';
            return false;
        }
        $date['password'] = salt_hash($data['password']);
        $user_info->save($date);
        //
        (new User([], 0))->where('shop_user_id', $user_info->shop_user_id)->find()?->save($date);
        //
        return true;
    }

    /*
    * SAAS-首次修改密码
    */
    public function saasEditPassword($data, $user)
    {
        $oldPassword = $data['old_password'] ?? '';
        $newPassword = $data['password'] ?? '';
        $confirmPassword = $data['confirm_password'] ?? '';
        if (!ValidateHelp::validateAlphaPassword($data['password'])) {
            $this->error = '不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种';
            return false;
        }
        if ($newPassword === $oldPassword) {
            $this->error = '新密码与旧密码不能相同';
            return false;
        }
        //
        $userInfo = User::detail($user['shop_user_id']);
        if ($userInfo['password'] != salt_hash($oldPassword)) {
            $this->error = '原密码错误';
            return false;
        }
        if ($userInfo['password_change'] > 0) {
            $this->error = '已经修改过密码';
            return false;
        }
        if ($newPassword != $confirmPassword) {
            $this->error = '两次密码输入不一致';
            return false;
        }
        $date['password_change'] = $userInfo['password_change'] + 1;
        $date['password'] = salt_hash($newPassword);
        $userInfo->save($date);
        //
        (new User([], 0))->where('shop_user_id', $userInfo->shop_user_id)->find()?->save($date);
        //
        return true;
    }

    /**
     * 获取用户信息
     */
    public static function getUser($data)
    {
        return (new static())->where('shop_user_id', '=', $data['uid'])->with(['app'])->find();
    }
}

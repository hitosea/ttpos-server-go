<?php

namespace app\shop\model\shop;

use help\ValidateHelp;
use app\common\model\shop\User as UserModel;
use app\common\model\shop\Access as AccessModel;
use app\common\model\shop\LoginLog as LoginLogModel;
use app\common\model\shop\UserShiftLog;
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

        // 先查询用户（不验证密码）
        $user = $this->with(['app', 'supplier'])
            ->where(function ($q) use ($username) {
                $q->whereRaw('BINARY username = :username', ['username' => $username]);
                $q->whereOr('phone', $username);
            })
            ->find();
        if (!$user) {
            $this->error = '账号或密码错误';
            return false;
        }

        // 验证密码（支持 MD5 和 bcrypt）
        list($isValid, $needUpgrade) = verify_password($password, $user['password']);
        if (!$isValid) {
            $this->error = '账号或密码错误';
            return false;
        }

        // 如果需要升级，异步升级为 bcrypt
        if ($needUpgrade) {
            upgrade_password_async($user['uuid'], 'ttpos_staff', 'uuid', 'password', $password, $user['company_uuid']);
        }
        if ($user['delete_time'] > 0) {
            $this->error = '账号已删除，请联系管理员';
            return false;
        }
        if ($user['is_disable'] == 1) {
            $this->error = '账号被禁用，请联系管理员';
            return false;
        }
        //
        if (empty($user['app'])) {
            $this->error = '未找到绑定的商家，请确认登录信息';
            return false;
        }
        if ($user['app']['delete_time']) {
            $this->error = '未找到绑定的商家，请确认登录信息';
            return false;
        }
        if ($user['app']['is_recycle']) {
            $this->error = '商家账号异常，请联系管理员';
            return false;
        }
        //
        request()->appId = $user['company_uuid'];
        request()->license = $user->license = $user['app']->getLicense();
        // 验证权限
        $permission = (new AccessModel)->getPermission(AccessModel::SHOP_ROUTE_NAME, $user, $user['supplier']);
        if (empty($permission)) {
            $this->error = '当前无权限，请联系管理员';
            return false;
        }
        // 保存登录状态
        $user['token'] = signToken($user['uuid'], 'shop', '', md5($user->password), $user['company_uuid']);
        // 货币信息
        $user['currency'] = SettingModel::getCurrency(0, $user['company_uuid']);
        // 写入登录日志
        LoginLogModel::add($username, \request()->ip(), '登录成功', $user['uuid']);
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
        $user_info = User::detail($user['uuid']);
        
        // 验证旧密码（支持 MD5 和 bcrypt）
        list($isValid, $needUpgrade) = verify_password($data['oldpass'], $user_info['password']);
        if (!$isValid) {
            $this->error = '原密码错误';
            return false;
        }
        
        if ($data['password'] != $data['confirmPass']) {
            $this->error = '两次密码输入不一致';
            return false;
        }
        
        // 使用 bcrypt 加密新密码
        $date['password'] = hash_password_bcrypt($data['password']);
        $user_info->save($date);
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
        $userInfo = User::detail($user['uuid']);
        
        // 验证旧密码（支持 MD5 和 bcrypt）
        list($isValid, $needUpgrade) = verify_password($oldPassword, $userInfo['password']);
        if (!$isValid) {
            $this->error = '原密码错误';
            return false;
        }
        
        if ($userInfo['password_change_count'] > 0) {
            $this->error = '已经修改过密码';
            return false;
        }
        if ($newPassword != $confirmPassword) {
            $this->error = '两次密码输入不一致';
            return false;
        }
        
        // 使用 bcrypt 加密新密码
        $date['password_change_count'] = $userInfo['password_change_count'] + 1;
        $date['password'] = hash_password_bcrypt($newPassword);
        $userInfo->save($date);
        //
        return true;
    }

    /**
     * 获取用户信息
     */
    public static function getUser($data)
    {
        return (new static())->where('uuid', '=', $data['uid'])->with(['app'])->find();
    }

    /**
     * 获取用户token
     */
    public static function getUserTokenByDutyNo($dutyNo)
    {
        $shiftLog = UserShiftLog::where('shift_no', $dutyNo)->find();
        if (!$shiftLog) {
            return '';
        }
        $user = self::where('uuid', $shiftLog['staff_uuid'])->find();
        if ($shiftLog['status'] == 0 && $user) {
            return signToken($user['uuid'], 'shop', '', md5($user->password), $user['company_uuid']);
        }
        if ($user && $user['bind_key']) {
            $user = self::where([
                ['bind_key', '=', $user['bind_key']],
                ['cashier_online', '=', 1],
                ['cashier_login_time', '>', 0],
                ['duty_no', '<>', ''],
            ])->find();
            if ($user) {
                return signToken($user['uuid'], 'shop', '', md5($user->password), $user['company_uuid']);
            }
        }
        
        return '';
    }
}

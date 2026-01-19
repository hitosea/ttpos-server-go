<?php

namespace app\admin\model;

use think\Model;
use app\admin\model\admin\Staff as SaasStaffModel;
use app\common\model\BaseModel;
use app\common\exception\BaseException;

class CompanyStaff extends BaseModel
{
    protected $name = 'company_staff';
    protected $pk = 'id';

    /**
     * 关联应用表
     */
    public function app()
    {
        return $this->belongsTo('app\\common\\model\\app\\App', 'company_uuid', 'uuid');
    }

    /**
     * 关联门店表
     */
    public function supplier()
    {
        return $this->belongsTo('app\\common\\model\\supplier\\Supplier', 'company_uuid', 'company_uuid');
    }

    /**
     * 检测账号是否存在集团中
     * @param string $username 用户名或手机号
     * @return mixed 存在返回company_uuid,不存在返回false
     */
    public static function checkExistInCompany($username)
    {
        return static::where(function ($query) use ($username) {
            $query->whereRaw('BINARY username = :username', ['username' => $username])
                ->whereOr('phone', $username);
        })->value('company_uuid') ?: false;
    }

    /**
     * 验证用户名是否重复
     */
    public static function checkExist($user_name)
    {
        // 区分字母大小写
        return !!static::withoutGlobalScope()->alias('user')
            ->leftJoin('company_setting su', "su.company_uuid = user.company_uuid")
            ->whereRaw('BINARY user.username = :username', ['username' => $user_name])
            ->value('user.uuid');
    }

    /**
     * 用户手机号是否存在
     */
    public static function checkPhoneExist($phone)
    {
        return !!static::withoutGlobalScope()->alias('user')
            ->leftJoin('company_setting su', "su.company_uuid = user.company_uuid")
            ->where('user.phone', '=', $phone)
            ->where('user.phone', '<>', '')
            ->value('user.uuid');
    }

    /**
     * 根据用户ID是否存在
     */
    public static function checkUserExist($shopUserId)
    {
        return !!static::withoutGlobalScope()->alias('user')
            ->leftJoin('company_setting su', "su.company_uuid = user.company_uuid")
            ->where('user.uuid', '=', $shopUserId)
            ->value('user.uuid');
    }

    /**
     * 新增商家用户记录
     */
    public function add($company_uuid, $data)
    {
        if (self::checkExist($data['user_name'])) {
            $this->error = '商家用户名已存在';
            return false;
        }

        $uuid = createUuid();
        $password = hash_password_bcrypt($data['password']);

        // 同时保存到ttpos_staff表中
        $saasStaff = new SaasStaffModel(); 
        if (!$saasStaff->save([
            'uuid' => $uuid,
            'email' => $data['user_name'],
            'real_name' => $data['user_name'],
            'phone' => $data['phone'],
            'password' => $password, 
            'company_uuid' => $company_uuid, 
        ])) {
            $this->error = $saasStaff->getError();
            return false;
        }

        return $this->save([
            'uuid' => $uuid,
            'username' => $data['user_name'],
            'phone' => $data['phone'],
            'password' => $password,
            'company_uuid' => $company_uuid,
            'is_super' => 1
        ]);
    }

    /**
     * 商家用户登录
     */
    public function login($app_id)
    {
        // 验证用户名密码是否正确
        $user = self::detail(['app_id' => $app_id], ['app']);
        if (empty($user)) {
            throw new BaseException(['msg' => '超级管理员用户信息不存在']);
        }
        $this->loginState($user);
    }
}

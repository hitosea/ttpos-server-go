<?php

namespace app\admin\model;

use think\Model;
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
        return $this->save([
            'uuid' => createUuid(),
            'username' => $data['user_name'],
            'phone' => $data['phone'],
            'password' => salt_hash($data['password']),
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

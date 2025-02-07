<?php

namespace app\common\model\shop;

use think\Model;
use think\facade\Cache;
use app\common\model\BaseModel;

/**
 * 商家用户模型
 */
class User extends BaseModel
{
    protected $name = 'company_staff';
    protected $pk = 'id';

    // 写入后同步数据
    public static function onAfterUpdate(Model $model)
    {
        if (app('http')->getName() == 'admin') {
            if ($model->getConnection() == 'mysql') {
                try {
                    (new self([], $model->company_uuid))->where('uuid', $model->uuid)->find()?->save($model);
                } catch (\Exception $e) {
                    //
                }
            }
        }
    }

    /**
     * 当real_name不存在时返回username
     * @return array
     */
    public static function getRealNameAttr($v, $data)
    {
        if (!$v) {
            return $data['username'] ?? '';
        }
        return $v;
    }

    /**
     * 关联应用表
     */
    public function app()
    {
        return $this->belongsTo('app\\common\\model\\app\\App', 'company_uuid', 'uuid');
    }

    /**
     * 关联用户角色表表
     */
    public function role()
    {
        return $this->belongsToMany('app\\common\\model\\auth\\Role', 'app\\common\\model\\auth\\UserRole');
    }

    public function userRole()
    {
        return $this->hasMany('app\\common\\model\\shop\\UserRole', 'staff_uuid', 'uuid');
    }

    /**
     * 关联门店表
     */
    public function supplier()
    {
        return $this->belongsTo('app\\common\\model\\supplier\\Supplier', 'company_uuid', 'company_uuid');
    }

    /**
     * 关联设备信息
     */
    public function device()
    {
        return $this->hasOne('app\\common\\model\\shop\\BindRecord', 'key', 'bind_key');
    }

    /**
     * 关联当班信息
     */
    public function working()
    {
        return $this->hasOne('app\\common\\model\\shop\\UserShiftLog', 'shift_user_id', 'shop_user_id')->where('shift_no', $this->duty_no)->where('status', 0);
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
     * 商家用户详情
     */
    public static function detail($where, $with = [])
    {
        !is_array($where) && $where = ['uuid' => (int)$where];
        return static::where(array_merge(['is_delete' => 0], $where))->with($with)->find()?->hidden(['password']);
    }

    /**
     * 保存登录状态
     */
    public function loginState($user)
    {
        $app = $user['app'];
        // 保存登录状态
        $session = array(
            'user' => [
                'shop_user_id' => $user['shop_user_id'],
                'user_name' => $user['user_name'],
                'shop_supplier_id' => $user['shop_supplier_id'],
                'user_type' => $user['user_type'],
            ],
            'supplier' => [
                'name' => isset($user['supplier']) && $user['supplier'] ? $user['supplier']['name'] : '',
                'category_set' => isset($user['supplier']) && $user['supplier'] ? $user['supplier']['category_set'] : 10,
                'is_main' => isset($user['supplier']) && $user['supplier'] ? $user['supplier']['is_main'] : 1,
            ],
            'app' => $app->toArray(),
            'is_login' => true,
        );
        session('jjjshop_store', $session);
    }

    /**
     * 获取店铺信息-绑定使用
     */
    public static function getShopInfo($key = '', $isClearCache = false)
    {
        if ((!$data = Cache::get('first_shop_info')) || $isClearCache) {
            $userModel = new static;
            $info = (new static)->withoutGlobalScope()->where('company_uuid', '>', 0)->field('company_uuid')->find();
            $company_uuid = $info?->company_uuid ?: 0;
            $data = compact('company_uuid');
            if ($company_uuid) {
                Cache::tag('firstshop')->set('first_shop_info', $data);
            }
        }
        return $key ? ($data[$key] ?? 0) : $data;
    }

    /**
     * 门店人员
     * @return array|\think\Collection
     */
    public function getList()
    {
        return $this->with(['userRole.role', 'supplier'])
            ->where('is_delete', '=', 0)
            ->order(['create_time' => 'desc'])
            ->hidden(['password'])
            ->select();
    }
}

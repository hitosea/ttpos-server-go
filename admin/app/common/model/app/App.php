<?php

namespace app\common\model\app;

use think\Model;
use think\facade\Cache;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use app\common\enum\http\StatusCode;
use app\common\exception\BaseException;
use app\common\model\supplier\Supplier;

/**
 * 应用模型
 */
class App extends BaseModel
{
    protected $name = 'company';
    protected $pk = 'id';

    // 写入后同步数据
    public static function onAfterUpdate(Model $model)
    {
        if (app('http')->getName() == 'admin') {
            if ($model->getConnection() == 'mysql' && $model->supplier?->deploy_mode == 1) {
                try {
                    (new self([], $model->id))->where('id', $model->id)->find()?->save($model);
                } catch (\Exception $e) {
                    //
                }
            }
        }
    }

    /**
     * 追加字段
     */
    protected $append = [
        'expire_time_text'
    ];

    /**
     *  过期时间
     */
    public function getExpireTimeTextAttr($value, $data)
    {
        // 发货状态
        if ($data['expire_time'] == 0) {
            return __('永不过期');
        }
        return date('Y-m-d H:i:s', $data['expire_time']);
    }

    /**
     *  开始时间
     */
    public function getAuthStartTimeAttr($value)
    {
        return $value ? date('Y-m-d H:i:s', $value) : 0;
    }

    /**
     * 关联子表
     */
    public function supplier()
    {
        return $this->hasOne(Supplier::class, 'company_id', 'id');
    }

    /**
     * 关联子表
     */
    public function suppliers()
    {
        return $this->hasMany(Supplier::class, 'parent_id', 'id');
    }

    /**
     * 关联店铺表
     */
    public function shopUsers()
    {
        return $this->hasMany(User::class, 'company_id', 'id');
    }

    /**
     * 获取应用信息
     */
    public static function detail($app_id = null): self
    {
        if (is_null($app_id)) {
            $self = new static();
            $app_id = $self::$app_id;
        }
        $detail = (new static())->find($app_id);
        if (!empty($detail['pay_type'])) {
            $detail['pay_type'] = json_decode($detail['pay_type'], true);
        }
        return $detail;
    }

    /**
     * 从缓存中获取app信息
     */
    public static function getAppCache($app_id = null)
    {
        if (is_null($app_id)) {
            $self = new static();
            $app_id = $self::$app_id;
        }
        if (!$data = Cache::get('app_' . $app_id)) {
            $data = self::detail($app_id);
            if (empty($data)) throw new BaseException(['msg' => '未找到当前应用信息']);
            Cache::tag('cache')->set('app_' . $app_id, $data);
        }
        return $data;
    }

    /**
     * 启用商城
     * @return bool
     */
    public function updateStatus()
    {
        $res = $this->allowField(['status', 'is_recycle'])->save([
            'is_recycle' => $this['is_recycle'] == 1 ? 0 : 1,
            'status' => $this['status'] == 1 ? 0 : 1,
        ]);
        return $res;
    }

    /**
     * 连锁开启关闭
     */
    public function updateChain()
    {
        return $this->allowField(['is_chain'])->save([
            'is_chain' => !$this['is_chain'],
        ]);
    }

    /**
     * 所有商城
     */
    public static function getAll()
    {
        return (new self())->where('is_delete', '=', 0)
            ->where('is_recycle', '=', 0)
            ->select();
    }

    /**
     * 所有商城
     */
    public static function getBySerial($serial_no)
    {
        return (new self())->withoutGlobalScope()->where('serial_no', '=', $serial_no)->find();
    }

    /**
     * 获取授权码
     */
    public function getLicense()
    {
        if ($this->expire_time > 0 && $this->expire_time < time()) {
            throw new BaseException(['msg' => '店铺状态已到期，如需继续使用，请联系销售代表', 'code' => StatusCode::EXPIRE_ERROR]);
        }
        if ($this->status == 0 || $this->is_delete == 1) {
            throw new BaseException(['msg' => '店铺状态异常，如需继续使用，请联系销售代表', 'code' => StatusCode::USER_ERROR]);
        }
        //
        $data = [
            'app_id' => $this->id,
            'c_n' => $this->supplier?->chain_number,                // 连锁编号
            'name' => $this->supplier?->name,                       // 真实姓名
            'sale' => $this->supplier?->sale_stock,                 // 进销存: 0不开启, 1开启
            'reserve' => $this->supplier?->reserve,                 // 预订: 0不开启, 1开启
            'c_l' => $this->supplier?->cash_limit,                  // 收银机上限
            'k_l' => $this->supplier?->kitchen_limit,               // 厨显上限
            't_l' => $this->supplier?->tablet_limit,                // 平板上限
            'a_l' => $this->supplier?->assistant_limit,             // 点餐助手上限
            'z_l' => $this->supplier?->table_limit,                 // 桌台上限
            'p_l' => $this->supplier?->printer_limit,               // 打印机上限
            'c_set' => $this->supplier?->category_set,              // 商品分类设置10同步主店20分店创建
            's_type' => $this->supplier?->store_type,               // 店铺类型10加盟20自营
            'level' => $this->supplier?->level,                     // 商家等级: 1开始
            'phone' => $this->supplier?->link_phone,                // 联系电话
            'addr' => $this->supplier?->address,                    // 商家地址
            'dm' => $this->supplier?->deploy_mode,                  // 部署方式 0局域网部署, 1云部署
            'domain' => request()->domain(),                        // request
            'day' =>  $this->auth_day,                              // 有效期天数
            'c_time' => strtotime($this->auth_start_time),          // 有效期开始时间
            '_mac' => $this->supplier?->mac_addr ?? '',
            '_uuid' => $this->supplier?->serial_number ?? '',
            '_ctime' => time(),
            'logo' => $this->supplier?->logo,                       // 商家logo
            'is_open_member' => $this->supplier?->is_open_member,
            'is_open_tablet' => $this->supplier?->is_open_tablet,
            'is_open_scan' => $this->supplier?->is_open_scan,
            'is_open_assistant' => $this->supplier?->is_open_assistant,
            'is_open_kitchen_kds' => $this->supplier?->is_open_kitchen_kds,
            'is_open_buffet' => $this->supplier?->is_open_buffet,
            'is_accept_scan_order' => $this->supplier?->is_accept_scan_order, // 是否开启扫码点餐接单
            'is_open_local_print' => $this->supplier?->is_open_local_print, // 是否开启本地打印服务
        ];
        //
        return $data;
    }
}

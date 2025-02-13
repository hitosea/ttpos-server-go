<?php

namespace app\common\model\shop;

use help\SystemHelp;
use app\common\model\BaseModel;
use app\common\model\order\Order;
use app\common\model\store\Table;
use app\common\enum\http\StatusCode;
use app\common\enum\settings\SettingEnum;
use app\shop\model\settings\Setting as SettingModel;
use app\common\model\shop\UserShiftLog as UserShiftLogModel;

/**
 * 设备绑定模型
 */
class BindRecord extends BaseModel
{
    protected $name = 'device';
    protected $pk = 'id';

    // 品牌
    const BRAND_A1_1500 = 'A1-1500';       //不带打印机的 compax收银机器
    const BRAND_A1_1510P = 'A2-1510P';     //自带打印机的 compax收银机器
    const BRAND_T2 = 'T2';                 //自带打印机的 商米收银机器
    const BRAND_T2S = 'T2s';               //自带打印机的 商米收银机器
    const BRAND_D2S_PLUS = 'D2s_PLUS';     //自带打印机的 商米收银机器
    const BRAND_T2MINI_S = 'T2mini_s';     //自带打印机的 商米小屏收银机器
    const BRANDS_ALL = [self::BRAND_A1_1500, self::BRAND_A1_1510P, self::BRAND_T2S, self::BRAND_T2MINI_S, self::BRAND_T2, self::BRAND_D2S_PLUS];    // 所有的机器
    const BRANDS_PRINTS = [self::BRAND_A1_1510P, self::BRAND_T2S, self::BRAND_T2MINI_S, self::BRAND_T2, self::BRAND_D2S_PLUS];                      // 所有带打印机的机器
    const BRANDS_SUNMI_ALL_PRINTS = [self::BRAND_T2, self::BRAND_T2S, self::BRAND_T2MINI_S, self::BRAND_D2S_PLUS];                                  // 商米所有带打印机的机器

    // 来源
    const SOURCE_CASHIER = 'cashier'; // 收银机
    const SOURCE_TABLET = 'tablet'; // 平板端
    const SOURCE_KITCHEN = 'kitchen'; // 厨显端
    const SOURCE_ASSISTANT = 'assistant'; // 点餐助手

    /**
     * 收银员
     */
    public function shopUser()
    {
        return $this->hasOne('app\\common\\model\\shop\\User', 'shop_user_id', 'finally_login_id')->hidden(['update_time', 'password']);
    }

    /**
     * 新增绑定记录
     */
    public function add($data, $licenseInfo = [])
    {
        $source = isset($data['source']) ? $data['source'] : '';
        $finally_login_id = isset($data['finally_login_id']) ? $data['finally_login_id'] : 0;
        $finally_login_time = isset($data['finally_login_time']) ? $data['finally_login_time'] : 0;
        $key = isset($data['key']) ? $data['key'] : '';
        $print_port_id = $data['print_port_id'] ?? 0;
        //
        $app_id = request()->appId;
        $shop_supplier_id = isset($data['shop_supplier_id']) ? $data['shop_supplier_id'] : 0;
        if (empty($key)) {
            $key = request()->header('deviceid') ?? '';
        }
        $brand = $data['brand'] ?? '';
        $platform = SystemHelp::getPlatform();
        $userAgent = request()->header('User-Agent') ?? '';
        $device_ip = $data['device_ip'] ?? '';
        if (!in_array($source, [self::SOURCE_CASHIER, self::SOURCE_TABLET, self::SOURCE_KITCHEN, self::SOURCE_ASSISTANT]) || !$shop_supplier_id || !$key) {
            $this->error = "来源设备错误";
            return false;
        }

        // 授权信息
        $licenseInfo = $licenseInfo ?: request()->licenses ?: [];
        $cash_limit = isset($licenseInfo['c_l']) ? $licenseInfo['c_l'] : 0;
        $kitchen_limit = isset($licenseInfo['k_l']) ? $licenseInfo['k_l'] : 0;
        $tablet_limit = isset($licenseInfo['t_l']) ? $licenseInfo['t_l'] : 0;
        $assistant_limit = isset($licenseInfo['a_l']) ? $licenseInfo['a_l'] : 0;

        // 兼容 device_remark 与 remark
        if (isset($data['remark'])) {
            $data['device_remark'] = $data['remark'];
        }

        // 已存在记录，因为前端sid永远慢一步的原因，暂时不需要sid条件查询并去掉全局查询范围
        $record = $this->withoutGlobalScope()->where('source', $source)->where('key', $key)->find();
        if ($record) {
            $record->withoutGlobalScope()->save([
                'id' => $record['id'],
                'print_port_id' => $print_port_id == 0 ? $record->print_port_id : $print_port_id,
                'remark' => ($data['device_remark'] ?? '') ?: $record->remark,
                'brand' => $brand,
                'platform' => $platform,
                'user_agent' => $userAgent,
                'device_ip' => $device_ip,
                'finally_login_id' => $finally_login_id,
                'finally_login_time' => $finally_login_time == 0 ? $record->finally_login_time : $finally_login_time,
            ]);
            return true;
        }

        //
        if ($source == self::SOURCE_CASHIER) {
            $bindCount =  $this->withoutGlobalScope()->where('finally_login_id', '>', 0)->where('source', $source)->count();
            if ($bindCount >= $cash_limit) {
                $this->error = "收银机登录设备已达上限，请在其他设备上退出登录或联系销售代表";
                return false;
            }
        }

        if ($source == self::SOURCE_TABLET) {
            $bindCount =  $this->withoutGlobalScope()->where('finally_login_id', '>', 0)->where('source', $source)->count();
            if ($bindCount >= $tablet_limit) {
                $this->errorCode = StatusCode::DEVICE_ERROR;
                $this->error = "平板登录设备已达上限，请在其他设备上退出登录或联系销售代表";
                return false;
            }
        }

        if ($source == self::SOURCE_KITCHEN) {
            $bindCount =  $this->withoutGlobalScope()->where('finally_login_id', '>', 0)->where('source', $source)->count();
            if ($bindCount >= $kitchen_limit) {
                $this->errorCode = StatusCode::DEVICE_ERROR;
                $this->error = "厨显登录设备已达上限，请在其他设备上退出登录或联系销售代表";
                return false;
            }
        }

        if ($source == self::SOURCE_ASSISTANT) {
            $bindCount =  $this->withoutGlobalScope()->where('finally_login_id', '>', 0)->where('source', $source)->count();
            if ($bindCount >= $assistant_limit) {
                $this->error = "点餐助手登录设备已达上限，请在其他设备上退出登录或联系销售代表";
                return false;
            }
        }

        // 绑定品牌，如果自带打印，默认更新收银打印配置
        if (in_array($brand, self::BRANDS_PRINTS)) {
            //
            $settingModel = new SettingModel();
            //
            $settingModel->setAppId($app_id);
            //
            $printerSettings = $settingModel::getSupplierItem(SettingEnum::PRINTER, $shop_supplier_id);
            if ($printerSettings && isset($printerSettings['cashier_open']) && $printerSettings['cashier_open'] == 1) {
                $setDefaultPrinter[] = [
                    'key' => $key,
                    'printer_id' => $key,
                ];
                $printerSettings['cashier_printer'] = $printerSettings['cashier_printer'] ?? [];
                $printerSettings['cashier_printer'] = array_merge($printerSettings['cashier_printer'], $setDefaultPrinter);
                // 去重
                $printerSettings['cashier_printer'] = array_values(array_reduce($printerSettings['cashier_printer'], function ($carry, $item) {
                    $carry[$item['key']] = $item;
                    return $carry;
                }, []));
                if (!$settingModel->edit(SettingEnum::PRINTER, $printerSettings, $shop_supplier_id, $app_id)) {
                    $this->error = "设置默认打印机失败";
                    return false;
                }
            }
        }

        $this->save([
            "uuid"=> createUuid(),
            'source' => $source,
            'key' => $key,
            'address' => $data['address'] ?? '',
            'port' => $data['port'] ?? '',
            'remark' => $data['device_remark'] ?? '',
            'brand' => $brand,
            'platform' => $platform,
            'user_agent' => $userAgent,
            'device_ip' => $device_ip,
            'finally_login_id' => $finally_login_id,
            'finally_login_time' => $finally_login_time,
        ]);
        return true;
    }

    /**
     * 更新设备备注
     */
    public function updateRemark($data)
    {
        $source = isset($data['source']) ? $data['source'] : '';
        $key = isset($data['key']) ? $data['key'] : '';
        if (empty($key)) {
            $key = request()->header('deviceid') ?? '';
        }
        if (!in_array($source, [self::SOURCE_CASHIER, self::SOURCE_TABLET, self::SOURCE_KITCHEN, self::SOURCE_ASSISTANT]) || !$key) {
            $this->error = "来源设备错误";
            return false;
        }

        // 兼容 device_remark 与 remark
        if (isset($data['remark'])) {
            $data['device_remark'] = $data['remark'];
        }

        // 已存在记录，因为前端sid永远慢一步的原因，暂时不需要sid条件查询并去掉全局查询范围
        $record = $this->withoutGlobalScope()->where('source', $source)->where('key', $key)->find();

        if ($record) {
            $record->withoutGlobalScope()->save([
                'id' => $record['id'],
                'remark' => $data['device_remark'] ?? '',
            ]);
            return true;
        }
        return false;
    }

    /**
     * 获取各端绑定记录
     */
    public function getBindList($source)
    {
        $query = $this;

        if ($source !== 'all') {
            if (!in_array($source, [self::SOURCE_CASHIER, self::SOURCE_TABLET, self::SOURCE_KITCHEN, self::SOURCE_ASSISTANT])) {
                $this->error = "来源设备错误";
                return false;
            }
            $query = $query->where('source', '=', $source);
        }

        $list = $query->with([
            'shopUser' => function ($query) {
                $query->field(['shop_user_id', 'IF(real_name = "", username, real_name) as real_name', 'cashier_login_time']);
            }
        ])->order(['create_time' => 'desc'])->select();
        // 处理收银机是否有交班  is_cashier_shift 1-已交班 0-未交班
        foreach ($list as &$item) {
            if ($item['source'] == self::SOURCE_CASHIER) {
                $item['is_cashier_shift'] = 1;
                $user = User::where(['bind_key' => $item['key']])->find();
                if ($user) {
                    $item['is_cashier_shift'] = $user->cashier_online ? 0 : 1;
                }
            }
            // 格式化时间
            $item['finally_login_time'] = date('Y-m-d H:i:s', $item['finally_login_time']);
        }
        return $list;
    }

    /**
     * 解绑设备
     */
    public function unbind($id)
    {
        $record = $this->where('id', '=', $id)->find();
        //
        $app_id = $record['app_id'] ?? 0;
        $shop_supplier_id = $record['shop_supplier_id'] ?? 0;
        if (!$record) {
            $this->error = "设备不存在";
            return false;
        }
        // 绑定数量
        $curNum = $this->where('source', '=', $record['source'])->count();
        $this->startTrans();
        try {
            // 收银机 （收银机没交班时：自动生成交班操作，预留金额为0，清用户设备绑定key）
            if ($record['source'] == self::SOURCE_CASHIER) {
                $user = User::where(['bind_key' => $record['key']])->where(['cashier_online' => 1])->find();
                if ($user) {
                    $data['shop_user_id'] = $user->shop_user_id;
                    $data['is_cash_taken_out_all'] = $curNum <= 1 ? 1 : 0;      // 解绑最后一个收银机时再取出所有钱，预留金额为0
                    $shift = (new UserShiftLogModel);
                    if ($shift->shiftLog($data, $record['key']) || !$shift->getError()) {
                        $user->save(['bind_key' => '']);
                    } else {
                        $this->error = "交班失败:" . $shift->getError();
                        return false;
                    }
                    // 取消收银台上的订单和挂单
                    Order::delStayOrder();
                }
                // 所有设备被解绑，桌台订单都取消
                if ($curNum <= 1) {
                    Order::delStayTableOrder();
                }

                // 解绑收银机设备时，去除该收银机的打印设置
                $settingModel = new SettingModel();
                $printerSettings = $settingModel::getSupplierItem(SettingEnum::PRINTER, $shop_supplier_id);
                if ($printerSettings && isset($printerSettings['cashier_printer'])) {
                    $printerSettings['cashier_printer'] = $printerSettings['cashier_printer'] ?? [];
                    $newPrinter = array_filter($printerSettings['cashier_printer'], function ($item) use ($record) {
                        return $item['key'] != $record['key'];
                    });
                    $printerSettings['cashier_printer'] = $newPrinter;
                    if (!$settingModel->edit(SettingEnum::PRINTER, $printerSettings, $shop_supplier_id, $app_id)) {
                        $this->error = "设置默认打印机失败";
                        return false;
                    }
                }
            }
            // 平板端（清桌台设备绑定key）
            if ($record['source'] == self::SOURCE_TABLET) {
                $table = Table::where('bind_info', $record['key'])->find();
                if ($table) {
                    $table->save(['is_bind' => 0, 'bind_info' => '']);
                }
            }
            // 厨显端（暂无处理）
            if ($record['source'] == self::SOURCE_KITCHEN) {
            }
            // 点餐助手（暂无处理）
            if ($record['source'] == self::SOURCE_ASSISTANT) {
            }
            $record->delete();
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }

    /**
     * 根据来源和key解绑
     */
    public function unbindByKey($source, $key)
    {
        $record = $this->where('source', '=', $source)->where('key', '=', $key)->find();
        if (!$record) {
            $this->error = "设备不存在";
            return false;
        }
        $record->delete();
        return true;
    }

    /**
     * 绑定设备数量
     */
    public function getBindCount($shopSupplierId)
    {
        $counts = $this->field('source, COUNT(*) as count')
            ->group('source')
            ->select()
            ->toArray();

        $counts = array_column($counts, 'count', 'source');

        $cashierCount = isset($counts[self::SOURCE_CASHIER]) ? $counts[self::SOURCE_CASHIER] : 0;
        $tabletCount = isset($counts[self::SOURCE_TABLET]) ? $counts[self::SOURCE_TABLET] : 0;
        $kitchenCount = isset($counts[self::SOURCE_KITCHEN]) ? $counts[self::SOURCE_KITCHEN] : 0;
        $assistantCount = isset($counts[self::SOURCE_ASSISTANT]) ? $counts[self::SOURCE_ASSISTANT] : 0;
        return compact('cashierCount', 'tabletCount', 'kitchenCount', 'assistantCount');
    }

    /**
     * 获取品牌
     */
    public static function getBrand(string $deviceId = '')
    {
        return self::withoutGlobalScope()->where('key', $deviceId)->value('brand') ?: '';
    }
}

<?php

namespace app\common\model\supplier;

use app\admin\model\app\App;
use think\Model;
use help\ImgHelp;
use help\DateHelp;
use help\ImageHelp;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\store\TakeOrder;
use app\common\enum\settings\SettingEnum;
use app\common\enum\settings\LanguageEnum;
use app\common\model\shop\User;
use app\shop\model\settings\Setting as SettingModel;
use help\HttpHelp;
use think\facade\Log;

/**
 * 商家供应商模型
 */
class Supplier extends BaseModel
{
    use SoftDelete;
    protected $name = 'company_setting';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    // 更新后同步数据
    public static function onAfterUpdate(Model $model)
    {
        if (app('http')->getName() == 'admin') {
            if ($model->getConnection() == 'mysql') {
                try {
                    (new self([], $model->company_uuid))->where('company_uuid', $model->company_uuid)->find()?->save($model);
                } catch (\Exception $e) {
                    //
                }
            }
        }
    }

    // 写入后同步数据
    public static function onAfterWrite(Model $data)
    {
        self::synchronousSetting($data);
    }

    /**
     * 属性
     */
    public function getDeliveryTimeAttr($value)
    {
        return $value ? json_decode($value, true) : '';
    }

    /**
     * 属性
     */
    public function setLanguagesAttr($value)
    {
        return json_encode($value ?: [], true);
    }

    /**
     * 属性
     */
    public function getLanguagesAttr($value)
    {
        return $value ? json_decode($value, true) : [];
    }

    /**
     * 属性
     */
    public function getPickTimeAttr($value)
    {
        return $value ? json_decode($value, true) : '';
    }

    /**
     * 属性
     */
    public function getDeliverySetAttr($value)
    {
        return $value ? json_decode($value, true) : [];
    }

    /**
     * 属性
     */
    public function setDeliverySetAttr($value)
    {
        if (!$value) {
            return [];
        }
        if (is_array($value)) {
            return json_encode($value) ?: [];
        }
        return $value;
    }

    /**
     * 属性
     */
    public function getStoreSetAttr($value)
    {
        return $value ? json_decode($value, true) : [];
    }

    /**
     * 属性
     */
    public function setStoreSetAttr($value)
    {
        if (!$value) {
            return [];
        }
        if (is_array($value)) {
            return json_encode($value) ?: [];
        }
        return $value;
    }

    /**
     * 属性
     */
    public function getStoreTimeAttr($value)
    {
        return $value ? json_decode($value, true) : '';
    }

    /**
     * 关联应用表
     */
    public function app()
    {
        return $this->belongsTo('app\\common\\model\\app\\App', 'company_uuid', 'uuid');
    }

    /**
     * 关联品牌类型
     */
    public function category()
    {
        return $this->hasOne('app\\common\\model\\supplier\\Category', 'category_id', 'category_id');
    }

    /**
     * 关联business
     */
    public function business()
    {
        return $this->hasOne('app\\common\\model\\file\\UploadFile', 'file_id', 'business_id');
    }

    /**
     * 关联超管
     */
    public function superUser()
    {
        return $this->hasOne('app\\common\\model\\shop\\User', 'company_uuid', 'company_uuid')
            ->where('is_super', '=', 1);
    }

    /**
     * 关联用户表
     */
    public function user()
    {
        return $this->belongsTo('app\\common\\model\\user\\User', 'user_id', 'user_id');
    }

    /**
     * 关联支付
     */
    public function paymentApp()
    {
        return $this->belongsTo('app\\common\\model\\pay\\PaymentApp', 'company_uuid', 'company_uuid');
    }

    /**
     * 详情
     */
    public static function detail($shop_supplier_id, $with = [])
    {
        return static::with($with)->where('id', $shop_supplier_id)->find();
    }

    /**
     * 累积供应商结算金额 (批量)
     */
    public function onBatchIncSupplierMoney($data)
    {
        foreach ($data as $supplierId => $supplierMoney) {
            $this->inc('total_money', $supplierMoney)->inc('money', $supplierMoney);
        }
        return true;
    }

    /**
     * 资金冻结
     */
    public function freezeMoney($money)
    {
        return $this->save([
            'money' => $this['money'] - $money,
            'freeze_money' => $this['freeze_money'] + $money,
        ]);
    }

    /**
     * 获得白色背景和黑色文本徽标路径
     */
    public static function getWhiteBackgroundWithBlackTextLogoPath($appId, $logoUrl = '')
    {
        // 可能是带query的url，需要去掉query
        $logoUrl = parse_url($logoUrl, PHP_URL_PATH);
        $savePath = public_path("uploads/shop$appId") . 'white_background_text_' . basename($logoUrl);
        // 判断图片是否存在
        if (!file_exists($savePath) && $logoUrl) {
            ImageHelp::whiteBackgroundWithBlackText('http://nginx/' . $logoUrl, $savePath);
        }
        //
        return $savePath;
    }

    /**
     * 获取列表数据
     */
    public static function getAllList()
    {
        $model = new static();
        // 查询列表数据
        return $model->order(['create_time' => 'desc'])->select();
    }

    /**
     * 启用进销存
     * @return bool
     */
    public function updateSaleStock()
    {
        return $this->allowField(['sale_stock'])->save([
            'sale_stock' => $this['sale_stock'] == 1 ? 0 : 1,
        ]);
    }

    /**
     * 启用预订
     * @return bool
     */
    public function updateReserve()
    {
        return $this->allowField(['reserve'])->save([
            'reserve' => $this['reserve'] == 1 ? 0 : 1,
        ]);
    }

    /**
     * 获取可用语言列表
     * @return bool
     */
    public function getLanguageList()
    {
        if (!$languages = $this->languages) {
            return [];
        }
        $languages = is_array($languages) ? $languages : json_decode($languages, true);
        $languages = is_array($languages) ? $languages : json_decode($languages, true);
        return array_values(array_filter(LanguageEnum::data(), function ($language) use ($languages) {
            return in_array($language['name'], $languages);
        }));
    }

    /**
     * 同步设置
     * @return bool
     */
    public static function synchronousSetting(self $data, $type = '')
    {
        $oidAppId = request()->appId ?: 0;
        request()->appId = $data->company_uuid;
        $companyUuid = $data->company_uuid;
        try {
            $languages = $data['languages'] ?? [];
            $languages = !is_array($languages) ? json_decode($languages, true) : $languages;
            $languages = !is_array($languages) ? json_decode($languages, true) : $languages;
            //
            if ($type == 'initShopBaseData') {
                $key = 0;
                $settingLanguages = [];
                foreach (LanguageEnum::data() as $language) {
                    if (in_array($language['name'], $languages)) {
                        $settingLanguages[] = [
                            'key' => $key = $key + 1,
                            'name' => $language['name'],
                            'value' => $language['value'],
                        ];
                    }
                }
            } else {
                $setting = SettingModel::detail(SettingEnum::STORE, $data->shop_supplier_id);
                $settingLanguages = ($setting->values['language'] ?? []);
                foreach ($settingLanguages as &$language) {
                    if (!in_array($language['name'], $languages)) {
                        $language['name'] = '';
                        $language['value'] = '-';
                    }
                }
            }
            $company = App::where('uuid', $companyUuid)->find();
            $logo = $company->logo ?: '';
            (new SettingModel)->edit(SettingEnum::STORE, [
                'name' => $company['name'],
                'logoUrl' => ImgHelp::removeImageDomain($logo),
                'time_zone' => $data['timezone'] ?? '',
                'phone' => $data['link_phone'] ?? '',
                'address' => $data['address'] ?? '',
                'coordinates' => $data['coordinates'] ?? '',
                'chain_number' => $data['chain_number'] ?? '',
                'tax_number' => $data['tax_number'] ?? $setting->values['tax_number'] ?? '', // 兼容后台修改商家授权信息，重置税号
                'language' => $settingLanguages
            ], $data->shop_supplier_id, $companyUuid, 1);
            // 拒单
            if (isset($data['is_open_h5_order']) && $data['is_open_h5_order'] == 0) {
                $user = User::where('company_uuid', $companyUuid)->find();
                if (!empty($user)) {
                    $token = signToken($user['uuid'], 'shop', '', md5($user->password), $companyUuid);
                    $res = HttpHelp::postRequest('http://nginx/api/v1/shop/order/reject_all', [], [
                        'Authorization: Bearer ' . $token,
                        'Accept-Language: ' . request()->header('language') ?: 'zh',
                    ]);
                    if (!$res) {
                        return false;
                    } 
                    Log::error($res);
                    $result = json_decode($res, true);
                    if (($result['code'] ?? -1) != 0) {
                        return false;
                    }
                }
            }
        } catch (\Exception $e) {
        }
        request()->appId = $oidAppId;
    }

    // 平板端获取基础信息
    public static function getTabletBaseInfo()
    {
        $detail = (new self)->withoutGlobalScope()->find();
        //
        $shopSupplierId = $detail['shop_supplier_id'] ?? 0;
        $appId = $detail['app_id'] ?? 0;
        $languageList = SettingModel::getSupplierLanguage($shopSupplierId, $appId);
        $settingData = SettingModel::getAll($appId, $shopSupplierId, $languageList);
        // 货币信息
        $currency = $settingData[SettingEnum::CURRENCY]['values'] ?? [];
        $detail['currency'] = [
            'unit' => $currency['unit'],
            'is_open' => $currency['is_open'],
            'unit_position' => $currency['unit_position'],
            'vices' => [
                'vice_unit' => $currency['vice_unit'],
                'vice_unit_position' => $currency['vice_unit_position'],
                'unit_rate' => $currency['unit_rate'],
            ],
        ];
        // 平板端设置
        $tablet = $settingData[SettingEnum::TABLET]['values'] ?? [];
        $languageList = $tablet['language_list'];
        unset($tablet['advanced_password']);
        unset($tablet['language_list']);
        $tablet['language_list'] = [];
        foreach ($languageList as $language) {
            if (in_array($language['key'], $tablet['language'])) {
                $tablet['language_list'][] = $language;
            }
        }
        $detail['tablet'] = $tablet;
        // 替换成商家后台设置的名称和logo
        $shop = $settingData[SettingEnum::STORE]['values'] ?? [];
        $detail['name'] = $shop['name'];
        $detail['logo'] = str_starts_with($shop['logoUrl'], 'http') ? $shop['logoUrl'] : base_url() . $shop['logoUrl'];
        //
        $business = $settingData[SettingEnum::BUSINESS]['values'] ?? [];
        $detail['no_clear_table'] = $business['no_clear_table'] ?? 0;   // 结账后不清台 0-清台 1-不清台
        // 自助餐设置
        $buffet = $settingData[SettingEnum::BUFFET]['values'] ?? [];
        $detail['buffet'] = $buffet;
        // 收银机设置
        $cashier = $settingData[SettingEnum::CASHIER]['values'] ?? [];
        $cashierDetail = $detail['cashier'];
        $cashierDetail['order_method'] = $cashier['order_method'];
        $detail['cashier'] = $cashierDetail;
        // 授权信息
        $detail['license_remaining_days'] = DateHelp::getLicenseRemainingDays(request()->licenses);
        // 云端基础信息
        $cloud_basic = SettingModel::getCloudBasic();
        $detail['cloud_basic'] = $cloud_basic;
        // 厨显设置
        $kitchen = $settingData[SettingEnum::KITCHEN]['values'] ?? [];
        $kitchenDetail = [];
        $kitchenDetail['is_open'] = $kitchen['is_open'] ?? 0;
        $detail['kitchen'] = $kitchenDetail;
        // 是否是云端部署 true-云端部署 false-本地部署
        $detail['is_cloud_deploy'] = env('IS_CLOUD_DEPLOY', false);
        // 订单设置
        return $detail;
    }

    // 扫码h5获取基础信息
    public static function getScanBaseInfo()
    {
        $detail = [];
        $shop_detail = (new self)->withoutGlobalScope()->find();
        //
        $shopSupplierId = $shop_detail['shop_supplier_id'] ?? 0;
        $appId = $shop_detail['app_id'] ?? 0;
        $languageList = SettingModel::getSupplierLanguage($shopSupplierId, $appId);
        $settingData = SettingModel::getAll($appId, $shopSupplierId, $languageList);
        $detail['shop'] = $shop_detail;
        // 货币信息
        $currency = $settingData[SettingEnum::CURRENCY]['values'] ?? [];
        $detail['currency'] = [
            'unit' => $currency['unit'],
            'is_open' => $currency['is_open'],
            'unit_position' => $currency['unit_position'],
            'vices' => [
                'vice_unit' => $currency['vice_unit'],
                'vice_unit_position' => $currency['vice_unit_position'],
                'unit_rate' => $currency['unit_rate'],
            ],
        ];
        // 扫码H5设置
        $h5 = $settingData[SettingEnum::H5]['values'] ?? [];
        $languageList = $h5['language_list'];
        unset($h5['language_list']);
        $h5['language_list'] = [];
        foreach ($languageList as $language) {
            if (in_array($language['key'], $h5['language'])) {
                $h5['language_list'][] = $language;
            }
        }
        $detail['h5'] = $h5;
        // 替换成商家后台设置的名称和logo
        $shop = $settingData[SettingEnum::STORE]['values'] ?? [];
        $detail['name'] = $shop['name'];
        $detail['logo'] = str_starts_with($shop['logoUrl'], 'http') ? $shop['logoUrl'] : base_url() . $shop['logoUrl'];
        // 自助餐设置
        $buffet = $settingData[SettingEnum::BUFFET]['values'] ?? [];
        $detail['buffet'] = $buffet;
        // 授权信息
        $detail['license_remaining_days'] = DateHelp::getLicenseRemainingDays(request()->licenses);
        $detail['is_open_h5_order'] = isset(request()->licenses['is_open_h5_order']) ? request()->licenses['is_open_h5_order'] : 0;
        // 云端基础信息
        $cloud_basic = SettingModel::getCloudBasic();
        $detail['cloud_basic'] = $cloud_basic;
        // 厨显设置
        $kitchen = $settingData[SettingEnum::KITCHEN]['values'] ?? [];
        $kitchenDetail = [];
        $kitchenDetail['is_open'] = $kitchen['is_open'] ?? 0;
        $detail['kitchen'] = $kitchenDetail;
        // 是否是云端部署 true-云端部署 false-本地部署
        $detail['is_cloud_deploy'] = env('IS_CLOUD_DEPLOY', false);
        // 订单设置
        return $detail;
    }

    // 电子菜单获取基础信息
    public static function getMenuBaseInfo()
    {
        $detail = [];
        $shop_detail = (new self)->withoutGlobalScope()->find();
        //
        $shopSupplierId = $shop_detail['shop_supplier_id'] ?? 0;
        $appId = $shop_detail['app_id'] ?? 0;
        $languageList = SettingModel::getSupplierLanguage($shopSupplierId, $appId);
        $settingData = SettingModel::getAll($appId, $shopSupplierId, $languageList);
        // 收银机设置
        $cashier = $settingData[SettingEnum::CASHIER]['values'] ?? [];
        // 货币信息
        $currency = $settingData[SettingEnum::CURRENCY]['values'] ?? [];
        $detail['currency'] = [
            'unit' => $currency['unit'],
            'is_open' => $currency['is_open'],
            'unit_position' => $currency['unit_position'],
            'vices' => [
                'vice_unit' => $currency['vice_unit'],
                'vice_unit_position' => $currency['vice_unit_position'],
                'unit_rate' => $currency['unit_rate'],
            ],
        ];
        // 电子菜单设置
        $menu['language_list'] = $languageList;
        $menu['menu_show_sold_out'] = $cashier['menu_show_sold_out'];
        $detail['menu'] = $menu;
        // 替换成商家后台设置的名称和logo
        $shop = $settingData[SettingEnum::STORE]['values'] ?? [];
        $detail['name'] = $shop['name'];
        $detail['logo'] = str_starts_with($shop['logoUrl'], 'http') ? $shop['logoUrl'] : base_url() . $shop['logoUrl'];
        // 云端基础信息
        $cloud_basic = SettingModel::getCloudBasic();
        $detail['cloud_basic'] = $cloud_basic;
        return $detail;
    }
}

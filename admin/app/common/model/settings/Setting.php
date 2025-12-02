<?php

namespace app\common\model\settings;

use app\common\model\user\Grade;
use help\IpHelp;
use think\Model;
use help\ImgHelp;
use help\JsonHelp;
use help\ArrayHelp;
use think\facade\Env;
use think\facade\Cache;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use app\common\model\shop\BindRecord;
use app\common\service\sync\SyncService;
use app\common\enum\settings\SettingEnum;
use app\common\enum\settings\BusinessEnum;
use app\common\enum\settings\LanguageEnum;
use app\common\enum\settings\TimeZoneEnum;
use app\common\model\settings\Printer as PrinterModel;

/**
 * 系统设置模型
 */
class Setting extends BaseModel
{
    protected $name = 'setting';
    protected $pk = 'key';

    /**
     * 获取器: 转义数组格式
     */
    public function getValuesAttr($value)
    {
        return $value ? json_decode($value, true) : [];
    }

    /**
     * 修改器: 转义成json格式
     */
    public function setValuesAttr($value)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 获取指定项设置
     */
    public static function getItem($key, $app_id = null)
    {
        $data = self::getAll($app_id);
        $data_key = $data[$key];
        if (isset($data_key)) {
            $data_key = $data[$key]['values'];
            JsonHelp::jsonRecursive($data_key);
        } else {
            $data_key = [];
        }
        return $data_key;
    }

    /**
     * 获取系统配置
     */
    public static function getSysConfig($key = SettingEnum::SYS_CONFIG)
    {
        $model = new static;
        $result = $model->withoutGlobalScope()->where('key', '=', $key)->value('values');
        if (!$result) {
            $appName = app('http')->getName();
            $languageList = self::getSupplierLanguage($appName != 'admin' ? User::getShopInfo('company_uuid') : 0)['language'] ?? [];
            $result = $model->defaultData(null, $languageList)[$key]['values'];
        } else {
            $result = json_decode($result, true);
        }
        return $result;
    }

    /**
     * 获取指定项语言设置
     */
    public static function getSupplierLanguage($company_uuid, $app_id = null)
    {
        $data = self::getAll($company_uuid);
        $data_key = $data[SettingEnum::STORE]['values'] ?? [];
        if (isset($data_key['language'])) {
            $data_key['language'] = $data_key['language'] ?: [];
        } else {
            $data_key['language'] = [];
        }
        // v1.0.8 语言数据兼容处理
        foreach ($data_key['language'] as $key => $language) {
            $data_key['language'][$key]['key'] = $language['name'];
            $data_key['language'][$key]['i'] = $language['key'] ?? '';
        }
        //
        return $data_key['language'];
    }

    /**
     * 获取指定项设置
     */
    public static function getSupplierItem($key, $shop_supplier_id, $app_id = null)
    {
        $languageList = $key != SettingEnum::STORE ? self::getSupplierLanguage($shop_supplier_id, $app_id) : [];
        $data = self::getAll($app_id, $shop_supplier_id, $languageList);
        $data_key = $data[$key];
        if (isset($data_key)) {
            $data_key = $data[$key]['values'];
            JsonHelp::jsonRecursive($data_key);
        } else {
            $data_key = [];
        }
        return $data_key;
    }

    /**
     * 获取设置项信息
     */
    public static function detail($key, $shop_supplier_id = 0)
    {
        return self::where('key', '=', $key)->find() ?? new self;
    }

    /**
     * 全局缓存: 系统设置
     */
    public static function getAll($app_id = null, $shop_supplier_id = 0, $languageList = [])
    {
        $static = new static;
        is_null($app_id) && $app_id = $static::$app_id;
        is_null($shop_supplier_id) && $shop_supplier_id = $static::$app_id;
        $shop_supplier_id == 0 && $shop_supplier_id = $static::$app_id;
        $setting = $static->select();
        $data = empty($setting) ? [] : array_column($static->collection($setting)->toArray(), null, 'key');

        $data = $static->getMergeData($data, $languageList);

        if ($shop_supplier_id > 0) {
            // 购物送积分规则处理会员
            $grades = (new Grade)->getLists();
            $rateMemberLevels = [];
            $quantityMemberLevels = [];
            foreach ($grades as $grade) {
                $rateMemberLevels[] = [
                    "uuid" => $grade['uuid'],
                    "name" => $grade['name'],
                    "value" => $grade['points_rate'],
                ];
                $quantityMemberLevels[] = [
                    "uuid" => $grade['uuid'],
                    "name" => $grade['name'],
                    "value" => $grade['points_quantity'],
                ];
            }
            foreach ($data[SettingEnum::POINTS]["values"]["shopping_gift_rules"] as $k => $shoppingGiftRule) {
                switch ($shoppingGiftRule["type"]) {
                    case "payment_amount":
                        $data[SettingEnum::POINTS]["values"]["shopping_gift_rules"][$k]["member_levels"] = $rateMemberLevels;
                        break;
                    case "desk":
                        $data[SettingEnum::POINTS]["values"]["shopping_gift_rules"][$k]["member_levels"] = $quantityMemberLevels;
                        break;
                }
            }
        }

        return $data;
    }

    /**
     * 数组转换为数据集对象
     */
    public function collection($resultSet)
    {
        $item = current(get_mangled_object_vars($resultSet));
        if ($item instanceof Model) {
            return \think\model\Collection::make($resultSet);
        } else {
            return \think\Collection::make($resultSet);
        }
    }


    /**
     * 合并用户设置与默认数据
     */
    private function getMergeData($userData, $languageList = [])
    {
        $defaultData = $this->defaultData(null, $languageList);
        if (isset($userData['store']['values']['checkedPay'])) {
            unset($defaultData['store']['values']['checkedPay']);
        }
        if (isset($userData['store']['values']['ip_white_list'])) {
            $userData['store']['values']['ip_white_list'] = env('PAY_SERVICE_IP', '');
        }
        // 时区列表拿默认列表
        if (isset($userData['store']['values']['time_zone_list'])) {
            unset($userData['store']['values']['time_zone_list']);
        }
        // 接单语音，设备本地处理，不需要合并
        if (isset($userData['cashier']['values']['is_auto_voice'])) {
            unset($userData['cashier']['values']['is_auto_voice']);
        }
        // 语言 不需要合并
        if (isset($userData['cashier']['values']['language'])) {
            unset($defaultData['cashier']['values']['language']);
        }
        if (isset($userData['tablet']['values']['language'])) {
            unset($defaultData['tablet']['values']['language']);
        }
        if (isset($userData['kitchen']['values']['language'])) {
            unset($defaultData['kitchen']['values']['language']);
        }
        // 过滤佛历
        if (isset($userData['printer']['values']['calendar_list'])) {
            unset($userData['printer']['values']['calendar_list']);
        }
        // 过滤打印方式
        if (isset($userData['printer']['values']['print_list'])) {
            unset($userData['printer']['values']['print_list']);
        }
        // 门店业务-过滤列表
        if (isset($userData['business']['values']['zeroing_method_list'])) {
            unset($userData['business']['values']['zeroing_method_list']);
        }
        if (isset($userData['business']['values']['checkout_zeroing_method_list'])) {
            unset($userData['business']['values']['checkout_zeroing_method_list']);
        }
        if (isset($userData['business']['values']['gift_method_list'])) {
            unset($userData['business']['values']['gift_method_list']);
        }
        if (isset($userData['business']['values']['free_method_list'])) {
            unset($userData['business']['values']['free_method_list']);
        }
        // 过滤助手支持功能
        if (isset($userData['assistant']['values']['support_function_list'])) {
            unset($userData['assistant']['values']['support_function_list']);
        }
        // 收银机、平板端图片列表不为空是，处理图片路径
        if (isset($userData['cashier']['values']['carousel']) && !empty($userData['cashier']['values']['carousel'])) {
            foreach ($userData['cashier']['values']['carousel'] as $key => &$item) {
                $item['file_path'] = ImgHelp::addImageDomain($item['file_path']);
            }
        }
        if (isset($userData['tablet']['values']['carousel']) && !empty($userData['tablet']['values']['carousel'])) {
            foreach ($userData['tablet']['values']['carousel'] as $key => &$item) {
                $item['file_path'] = ImgHelp::addImageDomain($item['file_path']);
            }
        }
        // 收银机设置是否显示售罄商品
        if (isset($defaultData['cashier']['values']['is_show_scan_sold_out']) || isset($userData['cashier']['values']['is_show_scan_sold_out'])) {
            $is_show_scan_sold_out = array_key_exists('is_show_scan_sold_out', $userData['cashier']['values'] ?? [])
                ? $userData['cashier']['values']['is_show_scan_sold_out']
                : $defaultData['cashier']['values']['is_show_scan_sold_out'];
            $defaultData['h5']['values']['is_show_scan_sold_out'] = $is_show_scan_sold_out;
        }
        if (isset($defaultData['cashier']['values']['is_show_assistant_sold_out']) || isset($userData['cashier']['values']['is_show_assistant_sold_out'])) {
            $is_show_assistant_sold_out = array_key_exists('is_show_assistant_sold_out', $userData['cashier']['values'] ?? [])
                ? $userData['cashier']['values']['is_show_assistant_sold_out']
                : $defaultData['cashier']['values']['is_show_assistant_sold_out'];
            $defaultData['assistant']['values']['is_show_assistant_sold_out'] = $is_show_assistant_sold_out;
            $userData['assistant']['values']['is_show_assistant_sold_out'] = $is_show_assistant_sold_out;
        }
        // 处理云端基础信息图片路径
        if (isset($userData['cloud_basic']['values']['base']) && !empty($userData['cloud_basic']['values']['base'])) {
            $base = &$userData['cloud_basic']['values']['base'];
            $base['brand_logo'] = ImgHelp::addImageDomain($base['brand_logo']);
            $base['brand_logo_long'] = ImgHelp::addImageDomain($base['brand_logo_long']);
        }
        // 处理商家信息图片路径
        if (isset($userData['store']['values']['logoUrl']) && !empty($userData['store']['values']['logoUrl'])) {
            $userData['store']['values']['logoUrl'] = ImgHelp::addImageDomain($userData['store']['values']['logoUrl']);
        }
        // 总权限 - 不开启自助餐 -（v1.1.1需要兼容初始化设置没有键值，但是商家端默认是开启的情况）
        $licenses = request()->licenses;
        if (($licenses['is_open_buffet'] ?? 0) == 0) {
            if (isset($userData['buffet']['values'])) {
                $userData['buffet']['values']['is_open'] = '0';
            }
            if (isset($defaultData['buffet']['values'])) {
                $defaultData['buffet']['values']['is_open'] = '0';
            }
            foreach ($defaultData["points"]['values']['shopping_gift_rules'] as $key => $pointsRule) {
                $newMealType = array_values(array_filter($pointsRule["meal_type"], function ($item) {
                    return $item != "buffet";
                }));
                $defaultData["points"]['values']['shopping_gift_rules'][$key]["meal_type"] = $newMealType;
            }
        }
        // 总权限 - 不开启厨显 -（v1.1.1需要兼容初始化设置没有键值，但是商家端默认是开启的情况）
        if (($licenses['is_open_kitchen_kds'] ?? 0) == 0) {
            if (isset($userData['kitchen']['values'])) {
                $userData['kitchen']['values']['is_open'] = '0';
            }
            if (isset($defaultData['kitchen']['values'])) {
                $defaultData['kitchen']['values']['is_open'] = '0';
            }
        }
        // 总权限 - 不开启会员
        if (($licenses['is_open_member'] ?? 0) == 0) {
            // 会员关闭时 门店管理 支付方式 余额这个开关要关了
            if (isset($userData['payment']['values'])) {
                $userData['payment']['values']['is_balance'] = '0';
            }
            // 会员关闭时 各端设置-点餐助手-支持功能-添加会员 去掉这个
            if (isset($userData['assistant']['values'])) {
                $supportFunction = $userData['assistant']['values']['support_function'] ?? '';
                if (is_array($supportFunction)) {
                    $userData['assistant']['values']['support_function'] = array_values(array_diff($supportFunction, ['add_member']));
                }
                //
                $supportFunctionList = $defaultData['assistant']['values']['support_function_list'] ?? '';
                if (is_array($supportFunctionList)) {
                    $defaultData['assistant']['values']['support_function_list'] = array_values(array_filter($supportFunctionList, function ($element) {
                        return $element['key'] != 'add_member';
                    }));
                }
            }
        }

        // 处理打印语言
        $userData['printer']['values']['language'] = $userData['printer']['values']['language'] ?? [];

        // v1.0.8 语言数据兼容处理
        $result = ArrayHelp::arrayMergeMultiple($defaultData, $userData);

        //
        return $result;
    }

    /**
     * 更新设置
     */
    public static function updateSetting(string $key, array $values, int $shop_supplier_id = 0): bool
    {
        $model = self::detail($key, $shop_supplier_id);
        //
        $model = $model->save(
            [
                'key' => $key,
                'describe' => SettingEnum::data()[$key]['describe'],
                'values' => $values,
            ]
        );
        // 删除系统设置缓存
        Cache::set(sprintf("setting:company_id:%d", self::$app_id), null);
        Cache::tag('common_get_settingLanguages')->clear();
        Cache::tag('cashier')->clear();
        //
        return $model !== null;
    }

    /**
     * 获取云数据
     */
    public static function getCloudBasic()
    {
        $cloudBasic = Cache::get(SettingEnum::CLOUD_BASIC) ?: [];
        if (!$cloudBasic || !isset($cloudBasic['base']) || empty($cloudBasic['base'])) {
            $cloudBasic = (new SyncService)->syncBaseInfo();
            if (!$cloudBasic) {
                return [];
            }
        }
        $base = &$cloudBasic['base'];
        $base['brand_logo'] = ImgHelp::addImageDomain($base['brand_logo'] ?? '', true);
        $base['brand_logo_long'] = ImgHelp::addImageDomain($base['brand_logo_long'] ?? '', true);
        if (!empty($base['browser_logo'])) {
            $base['browser_logo'] = ImgHelp::addImageDomain($base['browser_logo'], true);
        }
        //
        return $cloudBasic;
    }

    /**
     * 保存云数据
     */
    public static function saveCloudBasic($data, $shopSupplierId = 0)
    {
        $appId = self::$app_id ?? 0;
        $values = [
            'base' => [
                'brand_name' => $data['setting']['name'],
                'brand_logo' => $data['setting']['logo'],
                'brand_logo_long' => $data['setting']['logo_long'],
                'browser_logo' => $data['setting']['browser_logo'] ?? '',
                'browser_title' => $data['setting']['browser_title'] ?? '',
                'expiration_reminder' => $data['setting']['reminder'],
            ],
            'shop' => [
                'expire_time_text' => "",
                'name' => $data['name'],
                'chain_number' => $data['c_n'] ?? '',
                'sale_stock' => $data['sale'],
                'reserve' => $data['reserve'],
                'cash_limit' => $data['c_l'],
                'kitchen_limit' => $data['k_l'],
                'tablet_limit' => $data['t_l'],
                'logo' => $data['logo'],
                'address' => $data['addr'],
                'link_phone' => $data['phone'],
            ],
        ];
        $model = self::detail(SettingEnum::CLOUD_BASIC, $shopSupplierId);
        $model->save([
            'key' => SettingEnum::CLOUD_BASIC,
            'describe' => SettingEnum::data()[SettingEnum::CLOUD_BASIC]['describe'],
            'values' => json_encode($values),
        ]);
        // 删除系统设置缓存
        Cache::set(sprintf("setting:company_id:%d", $appId), null);
        Cache::tag('common_get_settingLanguages')->clear();
        Cache::set('sync_setting_' . SettingEnum::CLOUD_BASIC, $data);
    }

    /**
     * 获取打印信息
     */
    public static function getPrinterInfo($printerConfig, $deviceId)
    {
        $printer = null;
        $printerId = 0;
        $cashierBindKey = "";
        $isCashierPrinter = false;
        $isCashierOpen = false;
        if (($printerConfig['cashier_open'] ?? '')) {
            $isCashierOpen = true;
            foreach ($printerConfig['cashier_printer'] as $config) {
                if ($deviceId == $config['key']) {
                    $printerId = $config['printer_id'] ?? 0;
                }
            }
            if (strlen($printerId) > 12 || !preg_match('/^-?\d+$/', $printerId)) {
                $printer = BindRecord::getBrand($printerId);
                $cashierBindKey = $printerId;
                $printerId = 0;
                $isCashierPrinter = true;
            } else if ($printerId != 0) {
                $printer = PrinterModel::detail($printerId);
            }
            //
            $printerId = (int)$printerId;
        }
        //
        return compact('printer', 'printerId', 'isCashierPrinter', 'isCashierOpen', 'cashierBindKey');
    }

    /**
     * 获取可用语言列表
     * @return array
     */
    public function getLanguageList($languages = [])
    {
        if (!$languages) {
            $languages = self::getCloudBasic()['shop']['languages'] ?? [];
        }
        return array_values(array_filter(LanguageEnum::data(), function ($language) use ($languages) {
            return in_array($language['name'], $languages);
        }));
    }

    /**
     * 获取可用语言列表
     * @param array $languages 云平台授权语言
     * @param array $settingLanguages 设置语言列表
     * @return array
     */
    public function getSyncLanguageList($syncLanguages, $settingLanguages)
    {
        $syncLanguages = !is_array($syncLanguages) ? json_decode($syncLanguages, true) : ($syncLanguages ?: []);
        $settingLanguages = !is_array($settingLanguages) ? json_decode($settingLanguages, true) : $settingLanguages;

        //
        if (!empty($settingLanguages)) {
            foreach ($settingLanguages as &$language) {
                if (!in_array($language['name'], $syncLanguages)) {
                    $language['name'] = '';
                    $language['value'] = '-';
                }
            }
        } else {
            // 默认授权语言
            $authLanguages = $this->getLanguageList($syncLanguages);
            foreach ($authLanguages as $key => $language) {
                $settingLanguages[] = [
                    'key' => $key + 1,
                    'name' => $language['name'],
                    'value' => $language['value'],
                ];
            }
        }

        return array_values($settingLanguages);
    }

    /**
     * 默认配置
     */
    public function defaultData($storeName = null, $languageList = [])
    {

        foreach ($languageList as $key => $language) {
            $languageList[$key]['index'] = $language['key'];
            unset($languageList[$key]['name']);
        }

        $defaultLanguage = $languageList[0]['key'] ?? 'en';

        return [
            SettingEnum::STORE => [
                'key' => 'store',
                'describe' => '商城设置',
                'values' => [
                    // 商城名称
                    'name' => $storeName ?: 'Shop',
                    // 是否开启短信验证
                    'sms_open' => true,
                    // 是否记录日志
                    'is_get_log' => true,
                    // 是否开启微信授权
                    'wx_open' => true,
                    //默认头像
                    'avatarUrl' => base_url() . 'image/user/avatarUrl.png',
                    //商城logo
                    'logoUrl' => base_url() . 'image/diy/logo.png',
                    // 抹零方式
                    'zeroing_method' => '0', // 0-不抹零 1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元
                    // ip白名单
                    'ip_white_list' => '',
                    // 时区
                    'time_zone' => '',
                    // 结账后不清台 0-清台 1-不清台
                    'no_clear_table' => '0',
                    // 时区列表
                    'time_zone_list' => TimeZoneEnum::data(2),
                    // 公司名称
                    'company' => '',
                    // 地址
                    'address' => '',
                    // 联系电话
                    'phone' => '',
                    // 税号
                    'tax_number' => '',
                    // 连锁编号
                    'chain_number' => '',
                    // 系统语言
                    'language' => '', // 语言列表
                    // 授权语言
                    'auth_language' => '',
                ],
            ],
            SettingEnum::TRADE => [
                'key' => 'trade',
                'describe' => '交易设置',
                'values' => [
                    'order' => [
                        'close_days' => '3',
                        'receive_days' => '30',
                        'points_days' => '7'
                    ],
                ]
            ],
            SettingEnum::STORAGE => [
                'key' => 'storage',
                'describe' => '上传设置',
                'values' => [
                    'default' => env('STORAGE_DRIVER', 'local'),
                    'engine' => [
                        'local' => [],
                        'qiniu' => [
                            'bucket' => '',
                            'access_key' => '',
                            'secret_key' => '',
                            'domain' => 'http://'
                        ],
                        'aliyun' => [
                            'bucket' => '',
                            'access_key_id' => '',
                            'access_key_secret' => '',
                            'domain' => 'http://'
                        ],
                        'qcloud' => [
                            'bucket' => '',
                            'region' => '',
                            'secret_id' => '',
                            'secret_key' => '',
                            'domain' => 'http://'
                        ],
                        'google' => [
                            'credentials_file' => env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME'),
                            'bucket' => $bucket = env('GOOGLE_APPLICATION_UPLOADS_BUCKET_NAME'),
                            'uploads_catalogue' => $catalogue = env('GOOGLE_APPLICATION_UPLOADS_CATALOGUE_NAME'),
                            'domain' => "https://storage.googleapis.com/$bucket/$catalogue"
                        ],
                    ]
                ],
            ],
            SettingEnum::SMS => [
                'key' => 'sms',
                'describe' => '短信通知',
                'values' => [
                    'default' => 'aliyun',
                    'engine' => [
                        'aliyun' => [
                            'AccessKeyId' => '',
                            'AccessKeySecret' => '',
                            'sign' => '',
                            'template_code' => ''
                        ],
                        'qcloud' => [
                            'AccessKeyId' => '',
                            'AccessKeySecret' => '',
                            'sign' => '',
                            'template_code' => ''
                        ],
                        'hwcloud' => [
                            'AccessKeyId' => '',
                            'AccessKeySecret' => '',
                            'sign' => '',
                            'sender' => '',
                            'template_code' => '',
                            'url' => ''
                        ],
                    ],
                ],
            ],
            SettingEnum::TPL_MSG => [
                'key' => 'tplMsg',
                'describe' => '模板消息',
                'values' => [
                    'payment' => [
                        'is_enable' => '0',
                        'template_id' => '',
                    ],
                    'delivery' => [
                        'is_enable' => '0',
                        'template_id' => '',
                    ],
                    'refund' => [
                        'is_enable' => '0',
                        'template_id' => '',
                    ],
                ],
            ],
            SettingEnum::PRINTER => [
                'key' => 'printer',
                'describe' => '小票打印机设置',
                'values' => [
                    'cashier_open' => '1',   // 是否开启打印
                    'cashier_printer_id' => '-1', // 打印机id
                    'cashier_printer' => [],
                    'language_list' => $languageList, // 语言列表
                    'language_method' => '1', // 语言方式（收银） 1-单语言 2-多语言
                    'default_language' => $defaultLanguage, // 打印语言（收银）
                    'print_method' => 1, // 打印方式（收银） 1-文本打印 2-图片打印
                    'kitchen_language' => $defaultLanguage, // 打印语言（送厨）
                    'kitchen_print_method' => 1, // 打印方式（送厨） 1-文本打印 2-图片打印
                    'consumption_tax' => '1', // 消费税 1显示全部类型 2仅显示商品已含税 3仅显示商品未含税 4全部不显示
                    'buffet_sign_open' => '1', // 自助餐标识设置（默认开启）
                    'monetary_unit_open' => '1', // 货币单位（默认开启）
                    // 日历列表 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
                    'calendar_list' => [
                        [
                            'key' => '1',
                            'name' => __('公历'),
                        ],
                        [
                            'key' => '3',
                            'name' => __('佛历'),
                        ],
                    ],
                    // 打印方式列表 （1-文本打印 2-图片打印 ）
                    'print_list' => [
                        [
                            'key' => '1',
                            'name' => __('文本打印'),
                        ],
                        [
                            'key' => '2',
                            'name' => __('图片打印'),
                        ],
                    ],
                    'default_calendar' => '1', // 日历类型 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
                ],
            ],
            SettingEnum::FULL_FREE => [
                'key' => 'full_free',
                'describe' => '满额包邮设置',
                'values' => [
                    'is_open' => '0',   // 是否开启满额包邮
                    'money' => '',      // 单笔订单额度
                ],
            ],
            SettingEnum::RECHARGE => [
                'key' => 'recharge',
                'describe' => '用户充值设置',
                'values' => [
                    'is_entrance' => '1',   // 是否允许用户充值
                    'is_custom' => '1',   // 是否允许自定义金额
                    'is_match_plan' => '1',   // 自定义金额是否自动匹配合适的套餐
                    'describe' => "1. 账户充值仅限微信在线方式支付，充值金额实时到账；\n" .
                        "2. 账户充值套餐赠送的金额即时到账；\n" .
                        "3. 账户余额有效期：自充值日起至用完即止；\n" .
                        "4. 若有其它疑问，可拨打客服电话400-000-1234",     // 充值说明
                ],
            ],
            SettingEnum::POINTS => [
                'key' => 'points',
                'describe' => '积分设置',
                'values' => [
                    'deduction_order' => '1',  // 扣款顺序 1-先主账户后赠送账户 2-先赠送账户后主账户 3-按比例
                    'deduct_ratio_main' => '100',  // 主账户扣款比例0-100
                    'deduct_ratio_gift' => '0',  // 赠送账户扣款比例0-100
                    'points_name' => '积分',         // 积分名称自定义
                    'is_shopping_gift' => '0',      // 是否开启购物送积分
                    'gift_ratio' => '100',            // 是否开启购物送积分
                    'is_shopping_discount' => '0',    // 是否允许下单使用积分抵扣
                    'discount' => [     // 积分抵扣
                        'discount_ratio' => '0.01',       // 积分抵扣比例
                        'full_order_price' => '100.00',       // 订单满[?]元
                        'max_money_ratio' => '10',             // 最高可抵扣订单额百分比
                    ],
                    // 充值说明
                    'describe' => "a) 积分不可兑现、不可转让,仅可在本平台使用;\n" .
                        "b) 您在本平台参加特定活动也可使用积分,详细使用规则以具体活动时的规则为准;\n" .
                        "c) 积分的数值精确到个位(小数点后全部舍弃,不进行四舍五入)\n" .
                        "d) 买家在完成该笔交易(订单状态为“已签收”)后才能得到此笔交易的相应积分,如购买商品参加店铺其他优惠,则优惠的金额部分不享受积分获取;",
                    'shopping_gift_rules' => [
                        [
                            'type' => 'payment_amount', // 按付款金额比例赠送
                            'is_open' => '0', // 是否开启: "1" - 开启; "0" - 关闭
                            'is_member_level_related' => '0', // 是否按会员等级赠送: "1" - 是; "0" - 否
                            'value' => '', // 积分比例
                            'payment_amount_requirement' => '', // 付款金额要求
                            'meal_type' => [ // 就餐类型: "buffet" - 自助餐; "non-buffet" - 非自助餐
                            ],
                            'balance_payment_get_points' => '1', // 会员余额支付是否赠送: "1" - 是; "0" - 否
                            'refund_return_points' => '1', // 退款自动扣积分: "1" - 是; "0" - 否
                            'member_levels' => [ // 会员等级
                            ],
                        ],
                        [
                            'type' => 'desk', // 按桌台人数赠送
                            'is_open' => '0', // 是否开启: "1" - 开启; "0" - 关闭
                            'is_member_level_related' => '0',  // 是否按会员等级赠送: "1" - 是; "0" - 否
                            'value' => '', // 积分数量
                            'payment_amount_requirement' => '', // 付款金额要求
                            'meal_type' => [ // 就餐类型: "buffet" - 自助餐; "non-buffet" - 非自助餐
                                "buffet"
                            ],
                            'balance_payment_get_points' => '1', // 会员余额支付是否赠送: "1" - 是; "0" - 否
                            'refund_return_points' => '0', // 退款自动扣积分: "1" - 是; "0" - 否
                            'member_levels' => [ // 会员等级
                            ],
                        ],
                    ],
                    'exchange' => [
                        'open_points_exchange' => '0', // 会员积分抵扣订单金额
                        'points_exchange_rate' => '', // 每积分抵扣应付金额
                        'auto_points_exchange' => '0', // 是否自动抵扣
                    ],
                ],
            ],
            SettingEnum::OFFICIA => [
                'key' => 'officia',
                'describe' => '公众号关注',
                'values' => [
                    'status' => 0
                ],
            ],
            SettingEnum::COLLECTION => [
                'key' => 'collection',
                'describe' => '引导收藏',
                'values' => [
                    'status' => 0
                ],
            ],
            SettingEnum::HOMEPUSH => [
                'key' => 'homepush',
                'describe' => '首页推送',
                'values' => [
                    // 是否开启
                    'is_open' => 0,
                ]
            ],
            SettingEnum::SIGN => [
                'key' => 'sign',
                'describe' => '签到有礼',
                'values' => [
                    // 是否开启
                    'is_open' => false,
                    // 签到规则
                    'content' => ''
                ]
            ],
            SettingEnum::GETPHOME => [
                'key' => 'getPhone',
                'describe' => '获取手机号设置',
                'values' => [
                    // 显示区域
                    'area_type' => [],
                    // 不再提示天数
                    'send_day' => 7
                ],
            ],
            SettingEnum::SYS_ADMIN_CONFIG => [
                'key' => 'sys_admin_config',
                'describe' => '系统设置',
                'values' => [
                    'brand_name' => 'Shop', // 商城名称
                    'brand_logo' => '/image/logo/ttpos_64_64.png', // 商城背景图
                    'brand_logo_long' => '/image/logo/ttpos_146_40.png', // 商城logo
                    'browser_logo' => '/image/logo/ttpos_64_64.png', // 浏览器LOGO
                    'browser_title' => 'Shop', // 浏览器标题
                    'expiration_reminder' => 0, // 商城logo
                ]
            ],
            SettingEnum::SYS_CONFIG => [
                'key' => 'sys_config',
                'describe' => '系统设置',
                'values' => [
                    'shop_name' => 'Shop', // 商城名称
                    'shop_bg_img' => '',        // 商城背景图
                    'shop_logo_img' => '',      // 商城logo
                    'cashier_name' => '收银台',  // 收银台名称
                    'cashier_bg_img' => '',     // 收银台背景图
                ]
            ],
            SettingEnum::BALANCE => [
                'key' => 'balance',
                'describe' => '充值设置',
                'values' => [
                    // 是否开启
                    'is_open' => '0',
                    // 是否可以自定义
                    'is_plan' => '1',
                    // 最低充值金额
                    'min_money' => 1,
                    // 充值说明
                    'describe' => "a) 账户充值仅限在线方式支付，充值金额实时到账；\n" .
                        "b) 有问题请联系客服;\n",
                ]
            ],
            SettingEnum::THEME => [
                'key' => 'theme',
                'describe' => '主题设置',
                'values' => [
                    'theme' => '0', //主题设置
                ],
            ],
            SettingEnum::DELIVER => [
                'key' => 'deliver',
                'describe' => '配送设置',
                'values' => [
                    'default' => 'local',
                    'engine' => [
                        'local' => [
                            'name' => '商家配送',
                            'time' => 0,
                        ],
                        'dada' => [
                            'name' => '达达配送',
                            'app_key' => '',
                            'app_secret' => '',
                            'source_id' => '', //商户编号
                            'shop_no' => '', //门店编号
                            'online' => '0', //0测试环境1正式环境
                        ],
                        'driver' => [
                            'name' => '配送员配送',
                        ],
                        'meituan' => [
                            'name' => '美团配送',
                            'app_key' => '',
                            'app_secret' => '',
                            'shop_no' => '',
                            'call_back' => '域名/index.php/job/notify/meituan_notify',
                        ],
                        'uu' => [
                            'name' => 'UU跑腿',
                            'app_id' => '',
                            'app_key' => '',
                            'openid' => '',
                            'city_name' => '',
                            'online' => '0', //0测试环境1正式环境
                        ],
                    ]
                ],
            ],
            SettingEnum::GROUP => [
                'key' => 'group',
                'describe' => '团购设置',
                'values' => [
                    // 团购保障
                    'explain' => "随时退，过期自动退",
                    // 未支付订单关闭时间，默认5分钟
                    'close_time' => '5',
                ]
            ],
            SettingEnum::BALANCE_CASH => [
                'key' => 'balance_cash',
                'describe' => '余额提现设置',
                'values' => [
                    // 是否开启
                    'is_open' => '0',
                    // 最低提现金额
                    'min_money' => 1,
                    // 提现比例
                    'cash_ratio' => 100,
                ]
            ],
            SettingEnum::CURRENCY => [
                'key' => 'currency',
                'describe' => '货币单位',
                'values' => [
                    'unit' => '฿', // 货币单位，默认泰铢
                    'print_unit' => '฿', // 货币单位 - 打印专用，默认泰铢
                    // 主货币显示位置 0-金额前 1-金额后
                    'unit_position' => '0',
                    'is_open' => '0', // 副货币单位开关
                    'vice_unit' => '', // 副货币单位
                    // 副货币显示位置 0-金额前 1-金额后
                    'vice_unit_position' => '0',
                    'unit_rate' => '', // 单位汇率
                ],
            ],
            SettingEnum::TAX_RATE => [
                'key' => 'tax_rate',
                'describe' => '税率管理',
                'values' => [
                    'is_open' => '0', // 是否开启 0关闭 1开启
                    'tax_rate' => '', // 税率（v.1.0.4.1废弃）
                    'calc_type' => '1',  // 计算类型 商品已含税价-1 商品未含税价-2
                    'add_tax_category' => [], // 增值税分类
                ],
            ],
            SettingEnum::SERVICE_CHARGE => [
                'key' => 'service_charge',
                'describe' => '服务费',
                'values' => [
                    'is_open' => '0', // 是否开启 0关闭 1开启
                    'charge_type' => '1', // 服务费类型 1-固定金额 2-百分比
                    'service_charge' => '', // 服务费金额
                    'service_charge_rate' => '', // 服务费率
                    'is_open_tax' => '0', // 税费 1-收取税费 0-不收取税费
                    'apply_scope' => '1', // 适用范围 1-全部 2-部分
                    'apply_scope_ordering' => '0', // 适用范围-点餐 0-关闭 1-开启
                    'apply_scope_table' => '0', // 适用范围-桌台 0-关闭 1-开启
                    'apply_scope_table_list' => [], // 适用范围-桌台id列表,
                    'service_fee_base' => '0', // 服务费基础 "0"-商品惠后价 "1"-商品价格合计
                ],
            ],
            SettingEnum::CASHIER => [
                'key' => 'cashier',
                'describe' => '收银机设置',
                'values' => [
                    // 上传后的轮播内容url（图片 + 视频）
                    'carousel' => [],
                    'is_auto_send' => '0', // 收银结账自动送厨房 0-关闭 1-开启
                    // 用餐方式 收银-is_cashier_order（0-关闭 1-开启） 桌台-is_table_order（0-关闭 1-开启）
                    'order_method' => [
                        'is_cashier_order' => '1',
                        'is_table_order' => '1',
                    ],
                    // 收银机服务器连接
                    'server' => [
                        'ip' => str_replace('addr:', '', Env::get('HARDWARE_SERVER_URL', IpHelp::getLanIp())),
                        'port' => Env::get('HARDWARE_SERVER_PORT', '8080'),
                    ],
                    'is_remain_color' => '1', // 是否开启剩余时长颜色 0-关闭 1-开启
                    // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
                    'remain_color' => ["#E50028", "#F2A000"],
                    // 高级设置密码
                    'advanced_password' => '666888',
                    'is_open_cashier_password' => '1', // 是否开启钱箱密码 0-关闭 1-开启
                    'cashier_password' => '666888', // 钱箱密码
                    'lock_password' => '666888', // 锁屏密码
                    // 是否开启自动锁屏 0-关闭 1-开启
                    'is_auto_lock_screen' => '1',
                    // 自动锁屏（秒），默认5分钟
                    'auto_lock_screen' => '300',
                    // 扫码点餐是否显示售罄商品 0-不显示 1-显示
                    'is_show_scan_sold_out' => '1',
                    // 点餐助手是否显示售罄商品 0-不显示 1-显示
                    'is_show_assistant_sold_out' => '1',
                    // 语言列表
                    'language_list' => $languageList,
                    // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
                    'language' => [$defaultLanguage],
                    // 默认语言
                    'default_language' => $defaultLanguage,
                    // 是否自动接单
                    'is_auto_order' => '0',
                    // 自动接单金额上限
                    'auto_order_limit' => '1000',
                    // 是否开启自动接单语音播报
                    'is_auto_voice' => '0',
                    // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
                    'menu_show_sold_out' => '1',
                    // 未点餐时轮播间隔(秒)
                    'no_order_carousel_interval' => '10',
                    // 点餐时展示模式 carousel/order/order_carousel
                    'order_display_mode' => 'carousel',
                    // 点餐时轮播间隔(秒)
                    'order_carousel_interval' => '10',
                ],
            ],
            SettingEnum::TABLET => [
                'key' => 'tablet',
                'describe' => '平板端设置',
                'values' => [
                    // 上传后的轮播内容url（图片 + 视频）
                    'carousel' => [],
                    'is_call_service' => '1', // 是否开启呼叫服务员 0-关闭 1-开启
                    'is_customer_order' => '1', // 是否开启顾客自助下单 0-关闭 1-开启
                    'is_voice_remind' => '1', // 是否开启声音提醒 0-关闭 1-开启
                    'is_show_sold_out' => '0', // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
                    // 自助餐下单限制
                    'is_buffet_order_limit' => '0', // 是否开启自助餐下单限制 0-关闭 1-开启
                    'buffet_order_limit' => [
                        'is_limit_time' => '1', // 是否开启时间限制 0-关闭 1-开启
                        'limit_time' => '0', // 时间限制（分钟）
                        'is_limit_num' => '1', // 是否开启数量限制 0-关闭 1-开启
                        'limit_num' => '0', // 数量限制
                    ],
                    // 非自助餐下单限制
                    'is_order_limit' => '0', // 是否开启非自助餐下单限制 0-关闭 1-开启
                    'order_limit' => [
                        'is_limit_time' => '1', // 是否开启时间限制 0-关闭 1-开启
                        'limit_time' => '0', // 时间限制（分钟）
                        'is_limit_num' => '1', // 是否开启数量限制 0-关闭 1-开启
                        'limit_num' => '0', // 数量限制
                    ],
                    // 平板服务器连接
                    'server' => [
                        'ip' => str_replace('addr:', '', Env::get('HARDWARE_SERVER_URL', IpHelp::getLanIp())),
                        'port' => Env::get('HARDWARE_SERVER_PORT', '8080'),
                    ],
                    // 高级设置密码
                    'advanced_password' => '666888',
                    // 语言列表
                    'language_list' => $languageList,
                    // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
                    'language' => [$defaultLanguage],
                    'default_language' => $defaultLanguage, // 默认语言
                ],
            ],
            SettingEnum::H5 => [
                'key' => 'h5',
                'describe' => '扫码H5设置',
                'values' => [
                    'is_call_service' => '1', // 是否开启呼叫服务员 0-关闭 1-开启
                    'is_customer_order' => '1', // 是否开启顾客自助下单 0-关闭 1-开启
                    'is_voice_remind' => '1', // 是否开启声音提醒 0-关闭 1-开启
                    'is_show_sold_out' => '0', // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
                    // 自助餐下单限制
                    'is_buffet_order_limit' => '0', // 是否开启自助餐下单限制 0-关闭 1-开启
                    'buffet_order_limit' => [
                        'is_limit_time' => '1', // 是否开启时间限制 0-关闭 1-开启
                        'limit_time' => '0', // 时间限制（分钟）
                        'is_limit_num' => '1', // 是否开启数量限制 0-关闭 1-开启
                        'limit_num' => '0', // 数量限制
                    ],
                    // 非自助餐下单限制
                    'is_order_limit' => '0', // 是否开启非自助餐下单限制 0-关闭 1-开启
                    'order_limit' => [
                        'is_limit_time' => '1', // 是否开启时间限制 0-关闭 1-开启
                        'limit_time' => '0', // 时间限制（分钟）
                        'is_limit_num' => '1', // 是否开启数量限制 0-关闭 1-开启
                        'limit_num' => '0', // 数量限制
                    ],
                    // 语言列表
                    'language_list' => $languageList,
                    // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
                    'language' => [$defaultLanguage],
                    'default_language' => $defaultLanguage, // 默认语言
                ],
            ],
            SettingEnum::KITCHEN => [
                'key' => 'kitchen',
                'describe' => '厨显设置',
                'values' => [
                    'is_open' => '1', // 是否开启厨显功能 0关闭 1开启
                    'is_come_dish' => '1', // 是否开启来菜提醒 0-关闭 1-开启
                    'is_call_service' => '1', // 是否开启顾客呼叫提醒 0-关闭 1-开启
                    // 厨显服务器连接
                    'server' => [
                        'ip' => str_replace('addr:', '', Env::get('HARDWARE_SERVER_URL', IpHelp::getLanIp())),
                        'port' => Env::get('HARDWARE_SERVER_PORT', '8080'),
                    ],
                    // 高级设置密码
                    'advanced_password' => '666888',
                    'is_wait_color' => '0', // 是否开启等待时长颜色 0-关闭 1-开启
                    // 时长颜色 10分钟-黄色#ffff00 20分钟-红色#ff0000
                    'wait_color' => [],
                    // 语言列表
                    'language_list' => $languageList,
                    // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
                    'language' => [$defaultLanguage],
                    'default_language' => $defaultLanguage, // 默认语言
                    'is_smart_kitchen' => '0' // 是否开启智能后厨 0-关闭 1-开启
                ],
            ],
            SettingEnum::ASSISTANT => [
                'key' => 'assistant',
                'describe' => '点餐助手设置',
                'values' => [
                    // 服务器连接
                    'server' => [
                        'ip' => str_replace('addr:', '', Env::get('HARDWARE_SERVER_URL', IpHelp::getLanIp())),
                        'port' => Env::get('HARDWARE_SERVER_PORT', '8080'),
                    ],
                    'is_remain_color' => '1', // 是否开启剩余时长颜色 0-关闭 1-开启
                    // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
                    'remain_color' => ["#E50028", "#F2A000"],
                    'advanced_password' => '666888', // 高级设置密码
                    'lock_password' => '666888', // 锁屏密码
                    'default_mode' => '0', // 默认模式 0-服务员模式 1-顾客模式
                    // 支持功能 （添加会员 人数 调整自助餐 转台 清台 并台 优惠折扣 价格 退菜 备注 结账）
                    'support_function_list' => [
                        [
                            'key' => 'add_member',
                            'name' => __('添加会员'),
                        ],
                        [
                            'key' => 'people',
                            'name' => __('人数'),
                        ],
                        [
                            'key' => 'adjust_buffet',
                            'name' => __('调整自助餐'),
                        ],
                        [
                            'key' => 'turn_table',
                            'name' => __('转台'),
                        ],
                        [
                            'key' => 'clear_table',
                            'name' => __('清台'),
                        ],
                        [
                            'key' => 'merge_table',
                            'name' => __('合单结账'), // 并台=>合单结账（1.0.8）
                        ],
                        [
                            'key' => 'discount',
                            'name' => __('优惠折扣'),
                        ],
                        [
                            'key' => 'price',
                            'name' => __('价格'),
                        ],
                        [
                            'key' => 'return_dish',
                            'name' => __('退菜'),
                        ],
                        [
                            'key' => 'remark',
                            'name' => __('备注'),
                        ],
                        [
                            'key' => 'settle',
                            'name' => __('结账'),
                        ],
                        [
                            'key' => 'transfer_dish',
                            'name' => __('转菜'),
                        ],
                        [
                            'key' => 'gift_dish',
                            'name' => __('赠菜'),
                        ],
                    ],
                    'support_function' => [],
                    // 是否开启自动锁屏 0-关闭 1-开启
                    'is_auto_lock_screen' => '1',
                    // 自动锁屏（秒），默认5分钟
                    'auto_lock_screen' => '300',
                    // 语言列表
                    'language_list' => $languageList,
                    // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
                    'language' => [$defaultLanguage],
                    // 默认语言
                    'default_language' => $defaultLanguage,
                    'is_check_order' => '0', // 下单校验高级密码不能为空 '0'-关闭 '1'-开启
                ],
            ],
            SettingEnum::BUFFET => [
                'key' => 'buffet',
                'describe' => '自助餐设置',
                'values' => [
                    // 是否开启自助餐 0-关闭 1-开启
                    'is_open' => '1', // 是否是云上部署
                    // 平板结束时间提醒（分）
                    'tablet_end_time' => '5',
                    // 剩余xx分不可继续点餐开关 0-关闭 1-开启
                    'is_remain_continue' => '0',
                    // 剩余xx分不可继续点餐
                    'remain_continue_time' => '10',
                    // 剩余xx分提醒不可继续点餐
                    'remain_continue_notice_time' => '10',
                    // 非自助餐商品到时是否能继续选购 0-关闭 1-开启
                    'is_buy_continue' => '0',
                    // 是否开启加钟 0-关闭 1-开启
                    'is_add_clock' => '0',
                    // 是否开启自助餐优惠折扣 0-关闭 1-开启
                    'is_buffet_discount' => '0',
                    // 名称 - 加钟时间（分）- 价格
                    'add_clock' => [],
                ],
            ],
            SettingEnum::PAYMENT => [
                'key' => 'payment',
                'describe' => '支付方式',
                'values' => [
                    // 是否开启现金支付 0-关闭 1-开启
                    'is_cash' => '0', // v1.0.9 不需默认，改可编辑
                    // 是否开启余额支付 0-关闭 1-开启
                    'is_balance' => '0', // v1.0.9 不需默认，改可编辑
                    // 是否开启其他方式支付 0-关闭 1-开启
                    'is_other' => '1',
                ],
            ],
            SettingEnum::BUSINESS => [
                'key' => 'business',
                'describe' => '门店业务设置',
                'values' => [
                    // 优惠折扣自动抹零方式列表
                    'zeroing_method_list' => BusinessEnum::zeroingMethodDefault(),
                    // 结账自动抹零方式列表
                    'checkout_zeroing_method_list' => BusinessEnum::checkoutZeroingMethodDefault(),
                    // 优惠折扣自动抹零方式
                    'zeroing_method' => '0',
                    // 结账自动抹零方式
                    'checkout_zeroing_method' => '0',
                    // 赠菜计算方式列表
                    'gift_method_list' => BusinessEnum::giftMethodDefault(),
                    // 赠菜计算方式
                    'gift_method' => '10',
                    // 免单计算方式列表
                    'free_method_list' => BusinessEnum::freeMethodDefault(),
                    // 免单计算方式
                    'free_method' => '10',
                    // 折扣计算方式 10-按百分比 20-直接减免
                    'discount_method' => '10',
                    // 电子菜单二维码校验失效值，6位数数字
                    'qr_code' => '123456',
                    // 结账后不清台 0-清台 1-不清台
                    'no_clear_table' => '0',
                    // 取消订单/退菜 0-无需密码 1-需要密码
                    'is_need_password' => '1',
                    // 菜品卡片样式 0-无图模式 1-图片模式
                    'dish_card_style' => '0',
                    // 菜品卡片样式最后更新时间
                    'dish_card_style_time' => '0',
                    // 开票信息 0-不需要填写 1-需要填写
                    'is_invoice' => '0',
                    // 营业时间
                    'opening_hours' => '',
                    // 外送商品价格比例
                    'delivery_price_ratio' => 100,
                    'is_batch' => '0', // 是否是分批商品 0-否 1-是
                    'batch_product_uuids' => [], // 分批商品UUID列表
                    'batch_tag_num' => 0,  // 分批类型数量

                    // 外卖功能开关 0-关闭 1-开启
                    'enable_order_source' => '0',
                    // 国籍功能开关 0-关闭 1-开启
                    'enable_nationality' => '0',
                ],
            ],
        ];
    }
}

<?php

namespace app\shop\model\settings;

use help\ImgHelp;
use help\ImageHelp;
use think\facade\Cache;
use app\common\model\app\App;
use app\common\model\shop\BindRecord;
use app\common\model\supplier\Supplier;
use app\common\enum\settings\SettingEnum;
use app\common\service\websocket\Websocket;
use app\common\model\settings\Setting as SettingModel;

class Setting extends SettingModel
{
    /**
     * 更新系统设置
     */
    public function edit($key, $values, $shopSupplierId = 0, $appId = 0, $source = 0)
    {
        $model = self::detail($key, $shopSupplierId) ?: $this;
        $appId = $appId > 0 ? $appId : (self::$app_id ?: 0);
        // 各端设置密码字段要保留
        $terminalArr = [SettingEnum::CASHIER, SettingEnum::TABLET, SettingEnum::KITCHEN, SettingEnum::ASSISTANT];
        if (in_array($key, $terminalArr)) {
            if (isset($model['values']['advanced_password']) && !isset($values['advanced_password'])) {
                $values['advanced_password'] = $model['values']['advanced_password'];
            }
            if (isset($model['values']['cashier_password']) && !isset($values['cashier_password'])) {
                $values['cashier_password'] = $model['values']['cashier_password'];
            }
            if (isset($model['values']['lock_password']) && !isset($values['lock_password'])) {
                $values['lock_password'] = $model['values']['lock_password'];
            }
        }
        //
        $data = [
            'key' => $key,
            'describe' => SettingEnum::data()[$key]['describe'],
            'values' => $values = array_merge($model->values, $values),
        ];
        $res = $model->save($data) !== false;
        // 向平台主表同步数据
        if ($source === 0 && $res && ($key == SettingEnum::SYS_CONFIG || $key == SettingEnum::STORE)) {
            $ayncData = [
                'name' => $values['name'],
                'logo' => ImgHelp::removeImageDomain($values['logoUrl'] ?? ''),
                'timezone' => $values['time_zone'] ?? '',
                'link_phone' => $values['phone'] ?? '',
                'address' => $values['address'] ?? '',
                'coordinates' => $values['coordinates'] ?? '',
                'chain_number' => $values['chain_number'] ?? '',
                'tax_number' => $values['tax_number'] ?? '',
            ];
            (new App)->where('uuid', $shopSupplierId)->find()?->save($ayncData);
            (new Supplier)->where('company_uuid', $shopSupplierId)->find()?->save($ayncData);
            (new App([], 0))->where('uuid', $shopSupplierId)->find()?->save($ayncData);
            (new Supplier([], 0))->where('company_uuid', $shopSupplierId)->find()?->save($ayncData);
            // 保存打印用的白底黑字的图片
            ImageHelp::whiteBackgroundWithBlackText('http://nginx/' . $values['logoUrl'], Supplier::getWhiteBackgroundWithBlackTextLogoPath($appId, $values['logoUrl']));
        }
        // 更新其他语言
        if ($key == SettingEnum::STORE && isset($values['language']) && !empty($values['language'])) {
            $names = array_column($values['language'], 'name');
            foreach ([SettingEnum::CASHIER, SettingEnum::TABLET, SettingEnum::H5, SettingEnum::KITCHEN, SettingEnum::ASSISTANT, SettingEnum::PRINTER] as $enum) {
                $setting = self::detail($enum, $shopSupplierId);
                $settingDefaultLanguage = $setting['values']['default_language'] ?? '';
                $settingKitchenLanguage = $setting['values']['kitchen_language'] ?? '';
                $settingValues = $setting['values'] ?? [];
                // 对比并删除不匹配的值
                if (isset($settingValues['language'])) {
                    foreach ($settingValues['language'] as $key => $value) {
                        if (!in_array($value, $names)) {
                            unset($settingValues['language'][$key]);
                        }
                    }

                    if (!in_array($names[0], $settingValues['language'])) {
                        $settingValues['language'][] = $names[0];
                    }
                    $settingValues['language'] = array_values($settingValues['language']);
                }
                //
                if (!in_array($settingDefaultLanguage, $names)) {
                    $settingValues['default_language'] = $names[0];
                }
                // 兼容打印设置厨显语言打印
                if (!in_array($settingKitchenLanguage, $names)) {
                    $settingValues['kitchen_language'] = $names[0];
                }
                //
                if (isset($settingValues['language_list'])) {
                    unset($settingValues['language_list']);
                }
                //
                if (!$setting->isEmpty()) {
                    $setting->values = $settingValues;
                    $setting->save();
                } else {
                    $settingValues['language'] = $names;
                    $setting->save([
                        'key' => $enum,
                        'values' => $settingValues,
                    ]);
                }
            }
        }
        // 更新主收银机
        if ($key == SettingEnum::CASHIER && isset($values['main_cashier_uuid'])) {
            (new BindRecord)->setMainCashier($values['main_cashier_uuid']);
        }
        // 更新厨显
        if ($key == SettingEnum::KITCHEN && isset($values['bind_list'])) {
            if (!is_array($values['bind_list'])) {
                $values['bind_list'] = [];
            }
            foreach ($values['bind_list'] as $relatedPrinter) {
                (new BindRecord)->setKitchenRelatedPrinterUuid($relatedPrinter['uuid'] ?? 0, $relatedPrinter['related_printer_uuid'] ?? 0);
            }
        }
        // 删除系统设置缓存
        $key = sprintf("setting:company_id:%d", $appId);
        if (Cache::has($key)) {
            Cache::delete($key);
        }
        // 删除全局缓存
        Cache::tag('common_get_settingLanguages')->clear();
        if ($appId && Cache::has('{common_get_settingLanguages}_' . 'common_setting_languages' . $appId)) {
            Cache::delete('{common_get_settingLanguages}_' . 'common_setting_languages' . $appId);
        }
        // 删除收银机缓存
        Cache::tag('cashier')->clear();

        // 推送配置更新
        Websocket::pushClient(
            $appId, 
            Websocket::SOURCE_All, 
            Websocket::SOURCE_All, 
            Websocket::UPDATE_CONFIG, 
            0,
            ['update_time' => time()]
        );

        //
        return $res;
    }

    /**
     * 数据验证
     */
    private function validValues($key, $values)
    {
        $callback = [
            'store' => function ($values) {
                return $this->validStore($values);
            },
            'printer' => function ($values) {
                return $this->validPrinter($values);
            },
        ];
        // 验证商城设置
        return isset($callback[$key]) ? $callback[$key]($values) : true;
    }

    /**
     * 验证商城设置
     */
    private function validStore($values)
    {
        if (!isset($values['delivery_type']) || empty($values['delivery_type'])) {
            $this->error = '配送方式至少选择一个';
            return false;
        }
        return true;
    }

    /**
     * 验证小票打印机设置
     */
    private function validPrinter($values)
    {
        if ($values['is_open'] == false) {
            return true;
        }
        if (!$values['printer_id']) {
            $this->error = '请选择订单打印机';
            return false;
        }
        if (empty($values['order_status'])) {
            $this->error = '请选择订单打印方式';
            return false;
        }
        return true;
    }

    /**
     * 获取货币信息
     */
    public static function getCurrency($shop_supplier_id = 0, $app_id = null)
    {
        $currency = static::getSupplierItem('currency', $shop_supplier_id, $app_id);
        return [
            'unit' => $currency['unit'],
            'is_open' => $currency['is_open'],
            // 主货币显示位置 0-金额前 1-金额后
            'unit_position' => $currency['unit_position'],
            'vices' => [
                'vice_unit' => $currency['vice_unit'],
                // 副货币显示位置 0-金额前 1-金额后
                'vice_unit_position' => $currency['vice_unit_position'],
                'unit_rate' => $currency['unit_rate'],
            ],
        ];
    }
}

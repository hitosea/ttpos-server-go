<?php

namespace app\common\template;

use app\common\library\helper;
use app\common\enum\settings\SettingEnum;

/**
 * 模版基类
 * Interface BaseService
 * @package app\common\model
 */
class BaseTemplate
{
    protected $lang;
    protected $setting;
    protected $defaultCalendar = 1;        // 日历： 1公历，3佛历
    protected $currencyUnit = "$";         // 金额单位
    protected $currencyUnitPosition = 0;   // 金额单位位置
    protected $consumptionTax = 0;         // 消费税类型
    protected $allSourceProductList;       // 所有源产品列表
    protected $isSunmi = false;            // 是否商米打印

    /**
     * 构造方法
     * @param $setting 系统设置
     * @param $allSourceProductList 所有源产品列表
     */
    public function __construct($setting = null, $allSourceProductList = null, $isSunmi = false)
    {
        $this->setting = $setting;
        $this->allSourceProductList = $allSourceProductList;
        $this->isSunmi = $isSunmi;
        //
        if ($this->setting) {
            $currency = $setting[SettingEnum::CURRENCY]['values'] ?? [];
            $settingPrinterConfig = $setting[SettingEnum::PRINTER]['values'] ?? [];
            //
            $this->defaultCalendar = $settingPrinterConfig['default_calendar'] ?? 1;
            $this->consumptionTax = $settingPrinterConfig['consumption_tax'] ?? 1;
            $this->currencyUnitPosition = intval($currency['unit_position'] ?? '0');
            $this->currencyUnit = !intval($settingPrinterConfig['monetary_unit_open'] ?? 1) ? '' : ($currency['print_unit'] ?? $this->currencyUnit);
        }
        //
        $this->lang = checkDetect();
    }

    /**
     * 获取价格和单位
     * @param $price 价格
     * @return string
     */
    public function getPriceAndUnit($price)
    {
        $price = helper::amountPermillage($price);
        //
        if ($this->currencyUnitPosition == 1) {
            return $price . $this->currencyUnit;
        }
        return $this->currencyUnit . $price;
    }

    /**
     * 是否缅甸语
     * @param $text 是否文本
     * @return string
     */
    public function isMy($text)
    {
        return ($this->lang == 'my' || preg_match('/[\x{1000}-\x{109F}]/u', $text));
    }

    /**
     * 是否缅甸语Text
     * @param $text 是否文本
     * @return string
     */
    public function isMyText($text)
    {
        return preg_match('/[\x{1000}-\x{109F}]/u', $text);
    }

    /**
     * 是否泰语
     * @param $text 是否文本
     * @return string
     */
    public function isThText($text)
    {
        return preg_match('/[\x{0E00}-\x{0E7F}]/u', $text) || ($text == '฿');
    }

    
}

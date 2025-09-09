<?php

namespace app\common\enum\settings;

use MyCLabs\Enum\Enum;

/**
 * 小票打印机类型 枚举类
 */
class PrinterTypeEnum extends Enum
{
    // 飞鹅打印机
    const FEI_E_YUN = 'FEI_E_YUN';

    // 飞鹅标签打印机
    const FEI_E_YUN_TAG = 'FEI_E_YUN_TAG';

    // 365云打印
    const PRINT_CENTER = 'PRINT_CENTER';

    // 商米 局域网内打印
    const SUNMI_LAN = 'SUNMI_LAN';

    // 商米 云打印
    const SUNMI_CLOUD = 'SUNMI_CLOUD';

    // 芯烨-有线
    const XPRINTER_LAN = 'XPRINTER_LAN';

    // 芯烨-WIFI
    const XPRINTER_WIFI = 'XPRINTER_WIFI';

    // Compax 收银打印机 80mm 自带
    const CASHIER_COMPAX = 'CASHIER_COMPAX';  // -1

    // SUNMI 商米 收银打印机 80mm 自带
    const CASHIER_SUNMI = 'CASHIER_SUNMI';   // 0
    
    // Codesoft（网口）80mm 
    const CODESOFT_LAN = 'CODESOFT_LAN';   

    //Codesoft（WIFI）80mm 
    const CODESOFT_WIFI = 'CODESOFT_WIFI';  

    // GP_CLOUD 
    const GP_CLOUD = 'GP_CLOUD';  

    // 获取打印机类型名称
    public static function getTypeName()
    {
        return [
            self::SUNMI_LAN => __('商米打印机（局域网）') . '80mm',
            self::SUNMI_CLOUD => __('商米打印机（云打印）') . '80mm',
            self::XPRINTER_LAN => __('芯烨打印机（有线）') . '80mm',
            self::XPRINTER_WIFI => __('芯烨打印机（WIFI）') . '80mm',
            self::CODESOFT_LAN => __('Codesoft（网口）') . '80mm',
            self::CODESOFT_WIFI => __('Codesoft（WIFI）') . '80mm',
            self::GP_CLOUD => __('佳博（云打印）') . '80mm'
        ];
    }

    // 获取打印机类型名称
    public static function getTypeNames()
    {
        return [
            self::FEI_E_YUN_TAG => __('飞鹅标签打印机') . ' (58mm)',
            self::PRINT_CENTER => __('365云打印') . ' (58mm)',
            self::FEI_E_YUN => __('飞鹅打印机') . ' 58mm',
            self::SUNMI_LAN => __('商米打印机（局域网）') . '80mm',
            self::XPRINTER_LAN => __('芯烨打印机（有线）') . '80mm',
            self::XPRINTER_WIFI => __('芯烨打印机（WIFI）') . '80mm',
            self::CODESOFT_LAN => __('Codesoft（网口）') . '80mm',
            self::CODESOFT_WIFI => __('Codesoft（WIFI）') . '80mm',
            self::GP_CLOUD => __('佳博（云打印）') . '80mm',
            0 => __('SUNMI 商米 收银打印机') . '80mm',
            -1 => __('Compax 收银打印机') . '80mm',
        ];
    }
}

<?php

namespace app\common\model\settings;

use app\common\model\BaseModel;
use app\common\enum\settings\SettingEnum;

/**
 * 打印机定制
 */
class PrinterCustomize extends BaseModel
{
    protected $name = 'printer_customize';
    protected $pk = 'id';
}

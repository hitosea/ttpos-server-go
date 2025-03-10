<?php

namespace app\common\model\settings;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

class PrinterType extends BaseModel 
{
    use SoftDelete;

    protected $name = 'printer_type';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $autoWriteTimestamp = true;

    protected $append = ['name_text'];

    /**
     * 获取名称文本
     */
    protected function getNameTextAttr($value, $data)
    {
        return extractLanguage($data['name']);
    }

    /**
     * 获取打印机类型
     */
    public static function getPrinterType()
    {
        $data = [];
        $list = self::select();
        foreach ($list as $item) {
            $data[$item['key']] = $item['name_text'];
        }

        return $data;
    }

    /**
     * 根据打印机类型key查询数据
     */
    public static function getPrinterTypeByKey($key)
    {
        return self::where('key', $key)->find();
    }
}
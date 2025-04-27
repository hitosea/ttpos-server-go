<?php

namespace app\common\model_old\delay;

use app\common\model_old\BaseModel;

/**
 *
 */
class Delay extends BaseModel
{
    protected $name = 'delay';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'name_text',
    ];

    /**
     * 获取名称
     */
    public function getNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name']);
    }

    public static function getList()
    {
        return (new self())->where('status', '=', 1)
            ->select();
    }
}

<?php

namespace app\common\model_old\client;

use app\common\model_old\BaseModel;

/**
 *  客户端模型
 */
class ClientVersion extends BaseModel
{
    protected $name = 'client_version';

     /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'update_log_text',
    ];

    /**
     * 结果兼容多语言
     */
    public function getApkdataAttr($value)
    {
        return $value ? json_decode($value, true) : "";
    }

    /**
     * 结果兼容多语言
     */
    public function getUpdateLogTextAttr($value, $data)
    {
        return extractLanguage($data['update_log'] ?? '');
    }
}

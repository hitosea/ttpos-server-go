<?php

namespace app\common\model\product;

use app\common\model\BaseModel;
use app\common\model\store\MultiLanguageName;

/**
 * 原料信息表
 */
class Material extends BaseModel
{
    protected $name = 'material';
    protected $pk = 'id';

    /**
     * 追加字段
     */
    protected $append = [];

    /**
     * 添加原料
     * @return bool
     */
    public function add($data)
    {
        $data['create_time'] = time();
        $data['update_time'] = time();
        $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($data['name']);
        return $this->save($data);
    }

    /**
     * 更新原料
     * @return bool
     */
    public function edit($data)
    {
        $data['create_time'] = time();
        $data['update_time'] = time();
        (new MultiLanguageName)->saveNames($data['name'], $this['multi_language_name_uuid']);
        return $this->save($data);
    }
}

<?php

namespace app\common\model\product;

use app\common\model\BaseModel;
use app\common\model\store\MultiLanguageName;
use think\model\concern\SoftDelete;

/**
 * 原料分类模型
 */
class MaterialCategory extends BaseModel
{
    use SoftDelete;

    protected $name = 'material_category';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $autoWriteTimestamp = true;

    /**
     * 关联多语言名称
     */
    public function multiLanguageName()
    {
        return $this->hasOne(MultiLanguageName::class, 'uuid', 'multi_language_name_uuid');
    }

    /**
     * 获取所有列表
     */
    public static function getAllList() 
    {
        $list = [];
        $categories = self::with([
            'multiLanguageName',
        ])->select();
        foreach ($categories as $category) {
            $name = (new MultiLanguageName())->getNames($category->multi_language_name_uuid);
            $list[] = [
                'uuid' => $category->uuid,
                'category_id' => $category->uuid,
                'name' => $name,
                'name_text' => extractLanguage($name),
                'child' => [],
            ];
        }
        return $list;
    }
}
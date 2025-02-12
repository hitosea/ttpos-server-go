<?php

namespace app\common\model\product;

use app\common\model\BaseModel;

/**
 * 打印标签模型
 */
class Label extends BaseModel
{
    protected $name = 'product_print_label';
    protected $pk = 'label_id';

    /**
     * 处理多语言
     */
    protected $append = ['label_name_text'];

    /**
     * 标签名称
     */
    public function getLabelNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['label_name']);
    }

    /**
     * 关联产品ids
     */
    public function getProductIdsAttr($value, $data = [])
    {
        return $this->product()->column('product_id');
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->hasMany('app\\common\\model\\product\\Product', 'label_id', 'label_id');
    }

    /**
     * 更新标签
     * @param mixed $data
     * @return void
     */
    public function updateLabel($data)
    {
        if ($data) {
            $addData = [];
            foreach ($data as $item) {
                $isExit = $this->where('label_name', '=', $item['label_name'])->count();
                if ($isExit == 0) {
                    $addData[] = [
                        'label_name' => $item['label_name'],
                    ];
                }
            }
            $addData && $this->saveAll($addData);
        }
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        return $this->order(['sort' => 'asc', 'create_time' => 'desc'])->select()?->append(['product_ids'], true);
    }

    /**
     * 详情
     */
    public static function detail($label_id)
    {
        return self::find($label_id);
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($label_id)
    {
        return Product::where('label_id', 'in', $label_id)->count() > 0;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null)
    {
        $filter = [
            'label_name' => $name,
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['label_id', '<>', $id];
        }
        return static::where($filter)->value('label_id') ? true : false;
    }
}

<?php

namespace app\common\model\buffet;

use app\common\model\BaseModel;
use app\common\model\buffet\BuffetCustomer;

/**
 *
 */
class CustomerType extends BaseModel
{
    protected $name = 'customer_type';

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

    /**
     * 列表
     */
    public static function getList()
    {
        return (new self())->where('status', '=', 1)->select();
    }

    /**
     * 删除
     */
    public function setDelete($id)
    {
        $this->startTrans();
        try {
            $find = self::where('id', $id)->find();
            if ($find) {
                $find->status = 0;
                $find->save();
            }
            // 删除自助餐顾客类型关联
            foreach (BuffetCustomer::where('customer_type_id', $id)->select() as $buffet) {
                $buffet->delete();
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            return false;
        }
    }
}

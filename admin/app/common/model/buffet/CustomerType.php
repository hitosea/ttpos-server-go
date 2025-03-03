<?php

namespace app\common\model\buffet;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\buffet\BuffetCustomer;

/**
 *
 */
class CustomerType extends BaseModel
{
    use SoftDelete;
    protected $name = 'buffet_customer_type';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'name_text',
    ];

    /**
     * 兼容字段
     */

    /**
     * 获取名称
     */
    public function getNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name']);
    }

    public function getIdAttr($value, $data = [])
    {
        return $data['uuid'];
    }

    /**
     * 列表
     */
    public static function getList()
    {
        return (new self())->select();
    }

    /**
     * 删除
     */
    public function setDelete($id)
    {
        $this->startTrans();
        try {
            $find = self::where('uuid', $id)->find();
            // 删除自助餐顾客类型关联
            foreach (BuffetCustomer::where('customer_type_uuid', $find['uuid'])->select() as $buffet) {
                $buffet->delete();
            }
            $find?->delete();
            $this->commit();
            return true;
        } catch (\Exception $e) {
            dump($e->getMessage());
            die;
            $this->rollback();
            return false;
        }
    }
}

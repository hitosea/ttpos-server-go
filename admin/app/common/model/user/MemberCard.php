<?php

namespace app\common\model\user;

use app\common\library\helper;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\service\order\OrderService;

/**
 * 会员卡关联表
 */
class MemberCard extends BaseModel
{
    use SoftDelete;
    protected $name = 'member_card';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 设置折扣
     */
    public function setDiscountAttr($value)
    {
        return helper::bcdiv($value ?? 0, 100, 4);
    }

    /**
     * 获取折扣
     */
    public function getDiscountAttr($value, $data)
    {
        return floatval(helper::bcmul($value ?? 0, 100, 2));
    }

    // 新增
    public function add($data)
    {
        $this->save($data);
        return $this->getLastInsID();
    }

    /**
     * 指定用户是否存在卡
     */
    public static function checkExistByUserId($memberUuid, $cardTypeUuid = 0)
    {
        $model = (new static)->where('member_uuid', '=', $memberUuid);

        if ($cardTypeUuid) {
            $model = $model->where('card_type_uuid', '=', $cardTypeUuid);
        }
        return $model->findOrEmpty();
    }

    /**
     * 关联会员卡表
     */
    public function card()
    {
        return $this->hasOne('app\\common\\model\\user\\Card', 'uuid', 'card_type_uuid');
    }
}

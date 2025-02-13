<?php

namespace app\common\model\user;

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
}

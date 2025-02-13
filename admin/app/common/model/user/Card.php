<?php

namespace app\common\model\user;

use app\common\model\BaseModel;
use app\common\enum\user\pointsLog\PointsLogSceneEnum;
use app\common\model\user\PointsLog as PointsLogModel;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum;
use app\common\model\user\BalanceLog as BalanceLogModel;

/**
 * 会员卡模型
 */
class Card extends BaseModel
{
    protected $name = 'member_card_type';
    protected $pk = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'expire_time_text',
        //
        'card_id',
        'card_name',
        'is_discount',
        'money',
        'receive_num',
    ];

    /**
     * 兼容字段
     */
    public function getCardIdAttr()
    {
        return $this->uuid ?: 0;
    }
    public function getCardNameAttr()
    {
        return $this->getData('name') ?: '';
    }
    public function getIsDiscountAttr($value, $data)
    {
        return $this->discount > 0 ? 1 : 0;
    }
    public function getMoneyAttr($value, $data)
    {
        return $this->price ?: 0;
    }
    public function getReceiveNumAttr($value, $data)
    {
        return (new CardRecord)->where('member_card_type_uuid', '=', $this->uuid)->count();
    }

    /**
     * 会员卡有效期
     * @param $value
     * @param $data
     * @return string
     */
    public function getExpireTimeTextAttr($value, $data)
    {
        return $data['expire'] ? __('有效期：') . $data['expire'] . __('个月') : __('永久有效');
    }

    /**
     * 优惠券数组转换
     * @param $value
     * @param $data
     * @return string
     */
    public function setOpenCouponsAttr($value)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 优惠券数组转换
     * @param $value
     * @param $data
     * @return string
     */
    public function getOpenCouponsAttr($value)
    {
        return $value ? json_decode($value, 1) : [];
    }

    /**
     * 获取详情
     */
    public static function detail($card_id, $with = [])
    {
        return (new static())->with($with)->where('uuid', $card_id)->find();
    }

    /**
     * 获取可用的会员卡列表
     */
    public static function getCardList()
    {
        $model = new static;
        return $model->order(['sort' => 'desc', 'create_time' => 'asc'])->select();
    }

    /**
     * 卡默认颜色
     */
    public static function getImage()
    {
        $image = [
            ['name' => '001.jpg', 'url' => base_url() . 'image/card/001.jpg'],
            ['name' => '002.jpg', 'url' => base_url() . 'image/card/002.jpg'],
            ['name' => '003.jpg', 'url' => base_url() . 'image/card/003.jpg'],
            ['name' => '004.jpg', 'url' => base_url() . 'image/card/004.jpg'],
            ['name' => '005.jpg', 'url' => base_url() . 'image/card/005.jpg'],
            ['name' => '006.jpg', 'url' => base_url() . 'image/card/006.jpg'],
            ['name' => '007.jpg', 'url' => base_url() . 'image/card/007.jpg'],
            ['name' => '008.jpg', 'url' => base_url() . 'image/card/008.jpg'],
            ['name' => '009.jpg', 'url' => base_url() . 'image/card/009.jpg'],
            ['name' => '010.jpg', 'url' => base_url() . 'image/card/010.jpg'],
            ['name' => '011.jpg', 'url' => base_url() . 'image/card/011.jpg'],
        ];
        return $image;
    }

    // 检测用户是否有余额/积分消费记录
    public function checkUserConsumeRecord($user_id, $card_id = 0)
    {
        if (!(new BalanceLogModel)->where('member_uuid', $user_id)->where('scene', BalanceLogSceneEnum::CONSUME)->findOrEmpty()->isEmpty()) {
            return true;
        }
        return !(new PointsLogModel)->where('member_uuid', $user_id)->where('scene', PointsLogSceneEnum::CONSUME)->findOrEmpty()->isEmpty();
    }
}

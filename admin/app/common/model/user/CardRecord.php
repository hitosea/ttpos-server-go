<?php

namespace app\common\model\user;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\library\helper;
use app\common\service\order\OrderService;

/**
 * 会员卡领取记录模型
 */
class CardRecord extends BaseModel
{
    use SoftDelete;
    protected $name = 'member_card_log';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'pay_time_text',
        'disabled',
        'order_id',
        'pay_price',
        'delete_time',
        'pay_type',
    ];

    /**
     * 兼容字段
     */
    public function getOrderIdAttr()
    {
        return $this->id ?: 0;
    }
    public function getPayPriceAttr()
    {
        return $this->price ?: 0;
    }
    public function getPayTimeTextAttr()
    {
        $createTime = $this->getData('create_time');
        return $createTime ? date('Y-m-d H:i:s', $createTime) : '';
    }
    public function getIsdeleteAttr()
    {
        return $this->delete_time ? 1 : 0;
    }
    public function getPayTypeAttr()
    {
        return 30;
    }

    /**
     * 会员卡是否有效
     * @param $value
     * @param $data
     * @return string
     */
    public function getDisabledAttr($value, $data)
    {
        if (isset($data['expire']) && $data['expire'] != 0 && $data['expire'] < time()) {
            return 1;
        }
        return 0;
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


    /**
     * 关联会员卡表
     */
    public function card()
    {
        return $this->belongsTo('app\\common\\model\\user\\Card', 'member_card_type_uuid', 'uuid');
    }

    /**
     * 关联会员表
     */
    public function user()
    {
        return $this->belongsTo('app\\common\\model\\user\\User', 'member_uuid', 'uuid')->field('*, nickname as nickName');
    }

    /**
     * 获取详情
     */
    public static function detail($order_id)
    {
        return (new static())->with(['card'])->where('uuid', $order_id)->find();
    }

    /**
     * 指定卡下是否存在用户
     */
    public static function checkExistByRecordId($card_id)
    {
        $model = new static;
        return !!$model->alias('member_card_log')
            ->leftJoin('member', 'member.uuid=member_card_log.member_uuid')
            ->where('member.delete_time', '=', 0)
            ->where('member_card_type_uuid', '=', (int)$card_id)
            ->count();
    }

    /**
     * 指定用户是否存在卡
     */
    public static function checkExistByUserId($user_id, $order_id = 0)
    {
        $model = (new static)->where('member_uuid', '=', $user_id);
        if ($order_id) {
            $model = $model->where('uuid', '=', $order_id);
        }
        return $model->findOrEmpty();
    }

    /**
     * 生成订单号
     */
    public function orderNo()
    {
        return OrderService::createOrderNo();
    }

    /**
     * 生成交易号
     * @return string
     */
    public function tradeNo()
    {
        return OrderService::createTradeNo();
    }

    /**
     * 获取会员卡最小折扣
     */
    public function getDiscount($user_id)
    {
        $discount = $this->alias('r')
            ->where('r.delete_time', '=', 0)
            ->where('r.pay_status', '=', 20)
            ->where('r.user_id', '=', $user_id)
            ->where('r.discount', '>', 0)
            ->where(function ($query) {
                $query->where('expire_time', '=', 0)->whereOr('expire_time', '>', time());
            })
            ->order('r.discount asc')
            ->value('r.discount');
        return $discount ? round($discount / 100, 2) : 0;
    }
}

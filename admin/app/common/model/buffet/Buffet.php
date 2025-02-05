<?php

namespace app\common\model\buffet;

use think\facade\Db;
use app\common\library\helper;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model\order\Order as OrderModel;
use app\common\model\order\OrderBuffet as OrderBuffetModel;

/**
 *
 */
class Buffet extends BaseModel
{
    use SoftDelete;
    protected $name = 'buffet';
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
     * 关联自助餐产品
     */
    public function buffetProducts()
    {
        return $this->hasMany(BuffetProduct::class, 'buffet_id', 'id')->with('product');
    }

    // 与BuffetDiscount模型的多对多关联
    public function buffetDiscount()
    {
        return $this->belongsToMany(BuffetDiscount::class, 'buffet_discount_rel', 'buffet_discount_id', 'buffet_id');
    }

    /**
     * 关联自助餐产品
     */
    public function buffetCustomerType()
    {
        return $this->hasMany('app\\common\\model\\buffet\\BuffetCustomer', 'buffet_id', 'id');
    }

    /**
     * 关联自助餐限购产品
     */
    public function buffetLimitProducts()
    {
        return $this->hasMany('app\\common\\model\\buffet\\BuffetProduct', 'buffet_id', 'id')->where('limit_num', '>', 0)->with('product');
    }

    /**
     * 关联自助餐税类
     */
    public function buffetTaxes()
    {
        return $this->hasMany('app\\common\\model\\buffet\\BuffetTax', 'buffet_id', 'id');
    }

    /**
     * 获取自助餐名称
     */
    public function getNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name']);
    }

    /**
     * 获取自助餐是否能删除 0-否 1-是
     */
    public function getCanDelete($data = [])
    {
        $isDeletable = 1;
        $orderBuffetModel = new OrderBuffetModel();
        $orderModel = new OrderModel();
        $buffetOrders = $orderBuffetModel->where('buffet_id', '=', $data['id'] ?? 0)->column('order_id');
        if (!empty($buffetOrders)) {
            $pendingOrderCount = $orderModel->whereIn('order_id', $buffetOrders)
                ->where('order_status', '=', OrderPayStatusEnum::PENDING)
                ->count();

            if ($pendingOrderCount > 0) {
                $isDeletable = 0;
            }
        }
        return $isDeletable;
    }

    public static function getList()
    {
        return (new self())->with([
            'buffetTaxes',
            'buffetCustomerType' => function ($q) {
                $q->order('id asc');
            }
        ])->where('status', '=', 1)
            ->order('sort asc,id desc')
            ->select();
    }

    // 获取自助餐优惠列表
    public static function getBuffetDiscountList($buffet_id)
    {
        return (new self)->with([
            'buffetDiscount' => function ($q) {
                $q->where('status', '=', 1);
            }
        ])->where('id', '=', $buffet_id)->find()->toArray();
    }

    // 获取自助餐商品ID集
    public static function getBuffetProductIds(array $buffet_ids)
    {
        return $buffet_ids ? BuffetProduct::where('buffet_id', 'in', $buffet_ids)->column('product_id') : [];
    }

    /**
     * 获取自助餐消费税
     * @param $rate
     * @param $price
     * @param $is_tax   // 是否含税 1-已含税 2-未含税
     * @return float|int
     */
    public static function getConsumptionTax($rate, $price, $is_tax)
    {
        if (!$rate) {
            return 0;
        }
        if ($is_tax == 1) {
            /**
             * 商品价格含税
             */
            // 商品税前价 = 商品价格 / （1 + 税率）
            $original_price = helper::bcdiv($price, helper::bcadd(1, helper::bcdiv($rate, 100, 7), 7), 3);  //  $price / (1 + $rate/100)
            $original_price = round($original_price, 2);    // 四舍五入保留两位
            // 消费税 = 商品价格 - 商品税前价
            $tax_price = helper::bcsub($price, $original_price);
        } else {
            /**
             * 商品价格不含税
             */
            // 消费税 = 商品价格 * 税率
            $tax_price = helper::bcmul($price,  helper::bcdiv($rate, 100, 7), 3);
            $tax_price = round($tax_price, 2);  // 四舍五入保留两位
        }
        return floatval($tax_price);
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null, $lang = 'zh')
    {
        //
        $filter = [
            [Db::raw("JSON_UNQUOTE(JSON_EXTRACT(name, '$.$lang'))"), '=', $name],
            'shop_supplier_id' => $shop_supplier_id
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['id', '<>', $id];
        }
        return static::where($filter)->value('id') ? true : false;
    }
}

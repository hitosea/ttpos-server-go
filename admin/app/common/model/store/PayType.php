<?php


namespace app\common\model\store;

use help\ImgHelp;
use app\common\library\helper;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model\file\UploadFile;
use app\common\service\payment\PaymentService;

/**
 * 门店支付类型模型
 */
class PayType extends BaseModel
{
    use SoftDelete;

    protected $pk = 'id';
    protected $name = 'payment_method';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    // 系统默认支付方式
    const BALANCE_VALUE = 10; // 余额
    const CASH_VALUE = 40;    // 现金

    /**
     * 渠道提供方
     */
    const SOURCE_SYSTEM = 0;
    const SOURCE_DEFAULT = 1;
    const SOURCE_LIANLIANPAY = 2;
    const SOURCE = [
        0 => '系统默认',
        1 => '自行添加',
        2 => 'LianLianPay',
    ];

    /**
     * lianlian渠道支付方式
     */
    const SOURCE_LIANLIAN_WECHAT_PAY = 90111;
    const SOURCE_LIANLIAN_ALI_PAY = 90222;
    const SOURCE_LIANLIAN_QR_PROMPT_PAY = 90333;
    const SOURCE_LIANLIAN_PAY_METHOD = [
        90111 => 'LIANLIAN_WECHAT_PAY',
        90222 => 'LIANLIAN_ALI_PAY',
        90333 => 'LIANLIAN_QR_PROMPT_PAY',
    ];

    // 结账显示类型
    const SHOW_CHECKOUT = 'checkout';
    const SHOW_RECHARGE = 'recharge';

    // 显示端
    const CASHIER_SHOW_VALUE = 10; // 收银台
    const ASSISTANT_SHOW_VALUE = 20; // 点餐助手端

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['source_text', 'remark', 'fee', 'value', 'img', 'qrcode', 'is_show_checkout', 'is_show_recharge'];

    /**
     * 兼容字段
     */
    public function getRemarkAttr($value, $data)
    {
        return $this->payment_name ?: '';
    }
    public function getFeeAttr($value, $data)
    {
        return $this->fee_percent ?: 0;
    }
    public function getValueAttr($value, $data)
    {
        return $this->code ?: 0;
    }
    public function getImgAttr($value, $data)
    {
        return $this->logo_url ?: '';
    }
    public function getQrcodeAttr($value, $data)
    {
        return $this->qrcode_url ?: '';
    }

    /**
     * 获取手续费
     */
    public function setFeePercentAttr($value)
    {
        return floatval(helper::bcdiv($value, 100, 4));
    }

    /**
     * 设置手续费
     */
    public function getFeePercentAttr($value, $data)
    {
        return floatval(helper::bcmul($value, 100, 2));
    }


    /**
     * 渠道获取
     */
    public function getSourceTextAttr($value, $data)
    {
        return __(self::SOURCE[$data['source']] ?? '-');
    }

    /**
     * 获取结账显示ids
     */
    public function getIsShowCheckoutAttr($value)
    {
        $str = '';
        if ($this->is_show_cashier) {
            $str .= self::CASHIER_SHOW_VALUE . ','; // 收银台
        }
        if ($this->is_show_assistant) {
            $str .= self::ASSISTANT_SHOW_VALUE . ','; // 点餐助手端
        }
        return $str ? explode(',', trim($str, ',')) : [];
    }

    /**
     * 获取充值显示ids
     */
    public function getIsShowRechargeAttr($value)
    {
        $str = '';
        if ($this->is_show_member_recharge) {
            $str .= self::CASHIER_SHOW_VALUE . ','; // 收银台
        }
        return $str ? explode(',', trim($str, ',')) : [];
    }

    /**
     * 设置结账显示ids
     */
    public function setIsShowCheckoutAttr($value)
    {
        return is_array($value) ? implode(',', $value) : $value;
    }

    /**
     * 设置充值显示idsids
     */
    public function setIsShowRechargeAttr($value)
    {
        return is_array($value) ? implode(',', $value) : $value;
    }

    /**
     * 关联logo图片
     */
    public function logoFile()
    {
        return $this->belongsTo(UploadFile::class, 'logo_file_uuid', 'uuid');
    }

    /**
     * 关联二维码图片
     */
    public function qrcodeFile()
    {
        return $this->belongsTo(UploadFile::class, 'qrcode_file_uuid', 'uuid');
    }

    /**
     * 计算手续费
     * @param $price
     * @return float
     */
    public function calFeeMoney($price)
    {
        $fee = $this->fee ?: 0;
        if ($fee > 0) {
            $fee_money = helper::bcmul($price, $this->fee / 100, 3);
            return round($fee_money, 2);
        }
        return 0;
    }

    /**
     * 获取数据
     */
    public static function listAll($shopSupplierId = 0, $appId = 0)
    {
        return self::with([
            'logoFile', 'qrcodeFile'
        ])->orderRaw('CAST(sort AS UNSIGNED)')
            ->field('*, code as value')
            ->order('create_time', 'desc')
            ->select()
            ->toArray();
    }

    /**
     * 获取数据 - (商家后台)
     */
    public static function list($shopSupplierId = 0, $appId = 0)
    {
        $LianLian_enable = (bool)PaymentService::checkSignSalt($shopSupplierId);
        $enumData = self::listAll($shopSupplierId, $appId);
        $licenses = request()->licenses;
        return array_values(array_filter($enumData, function ($item) use ($LianLian_enable, $licenses) {
            // 过滤 Grab 和 LINE MAN 支付方式
            if ($item['value'] == OrderPayTypeEnum::GRAB || $item['value'] == OrderPayTypeEnum::LINE_MAN) {
                return false;
            }
            switch ($item['value']) {
                case OrderPayTypeEnum::BALANCE:
                    return $licenses['is_open_member'] == 1; // 会员功能开启时显示 0关闭 1开启
                case OrderPayTypeEnum::FREE_PAY:
                    return false;
                case OrderPayTypeEnum::FREE_MEAL_FOR_ERP:
                    return false;
                case OrderPayTypeEnum::LIANLIAN_ALI_PAY:
                case OrderPayTypeEnum::LIANLIAN_QR_PROMPT_PAY:
                case OrderPayTypeEnum::LIANLIAN_WECHAT_PAY:
                    return $LianLian_enable && $item['source'] == 2;
            }
            return true;
        }));
    }

    /**
     * 获取启用数据 - 收银
     */
    public static function getEnableList($showType, $showValue, $shopSupplierId = 0, $appId = 0): array
    {
        $LianLian_enable = (bool)PaymentService::checkSignSalt($shopSupplierId);
        return array_values(array_filter(self::list($shopSupplierId, $appId), function ($item) use ($LianLian_enable, $showType, $showValue) {
            switch ($item['value']) {
                case OrderPayTypeEnum::LIANLIAN_ALI_PAY:
                case OrderPayTypeEnum::LIANLIAN_QR_PROMPT_PAY:
                case OrderPayTypeEnum::LIANLIAN_WECHAT_PAY:
                    return $item['status'] == 1 && $LianLian_enable && (
                        ($showType == self::SHOW_CHECKOUT && in_array($showValue, $item['is_show_checkout'])) ||
                        ($showType == self::SHOW_RECHARGE && in_array($showValue, $item['is_show_recharge']))
                    );
                case OrderPayTypeEnum::FREE_PAY:
                    return false;
            }
            // 结账显示类型
            if ($showType == self::SHOW_CHECKOUT) {
                return $item['status'] == 1 && in_array($showValue, $item['is_show_checkout']);
            }
            // 充值显示类型
            if ($showType == self::SHOW_RECHARGE) {
                return $item['status'] == 1 && in_array($showValue, $item['is_show_recharge']);
            }
            // 默认
            return $item['status'] == 1;
        }));
    }

    /**
     * 获取启用数据 - 程序
     */
    public static function getEnableListAll($shopSupplierId = 0, $appId = 0): array
    {
        return array_values(array_filter(self::listAll($shopSupplierId, $appId), function ($item) {
            return $item['status'] == 1;
        }));
    }

    /**
     * 验证数据
     */
    public function validateData($param)
    {
        $remark = $param['remark'] ?? '';
        if (!$remark) {
            $this->error = '请输入名称';
            return false;
        }
        if (!($param['name'] ?? '')) {
            $this->error = '请输入支付方式';
            return false;
        }
        $param['fee'] = $param['fee'] ?? 0;
        if ($param['fee'] < 0 || $param['fee'] > 100) {
            $this->error = '支付手续费范围0-100';
            return false;
        }
        if (!($param['img'] ?? '')) {
            $this->error = '请上传图片';
            return false;
        }
        $sort = $param['sort'] ?? '';
        if ($sort == '' || !is_numeric($sort)) {
            $this->error = '请输入排序';
            return false;
        }

        //
        unset($param['id']);
        $max_value = self::withTrashed()->where('source', self::SOURCE_DEFAULT)->max('code');
        $param['code'] = ($max_value !== null && $max_value >= 20000)  ? $max_value + 100 : 20000; // 删除的值也要计算，防止重复，如果数据库找不到，则默认从20000开始
        $param['logo_file_uuid'] = $param['img']['file_id'] ?? 0;
        $param['qrcode_file_uuid'] = $param['qrcode']['file_id'] ?? 0;
        $param['logo_url'] = ImgHelp::removeImageDomain($param['img']['file_path'] ?? '');
        $param['qrcode_url'] = ImgHelp::removeImageDomain($param['qrcode']['file_path'] ?? '');
        $param['payment_name'] = $remark;
        $param['fee_percent'] = $param['fee'];
        $param['is_show_cashier'] = in_array(self::CASHIER_SHOW_VALUE, $param['is_show_checkout'] ?: []) ? 1 : 0;
        $param['is_show_assistant'] = in_array(self::ASSISTANT_SHOW_VALUE, $param['is_show_checkout'] ?: []) ? 1 : 0;
        $param['is_show_member_recharge'] = in_array(self::CASHIER_SHOW_VALUE, $param['is_show_recharge'] ?: []) ? 1 : 0;
        $param['uuid'] = createUuid();
        return $param;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null)
    {
        $LianLian_enable = (bool)PaymentService::checkSignSalt($shop_supplier_id);
        $filter = [
            'payment_name' => $name,
            'source' => self::SOURCE_DEFAULT, // v.2.6.0 35257 渠道提供方不一样，支付方式跟支付名称可以相同
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['id', '<>', $id];
        }
        if (!$LianLian_enable) {
            $filter[] = ['source', '<>', '2'];
        }
        return static::where($filter)->value('id') ? true : false;
    }

    /**
     * 设置支付方式状态 0-关闭 1-开启
     */
    public function setStatus($value, $status)
    {
        if (!in_array($status, [0, 1])) {
            $this->error = '状态值不合法';
            return false;
        }
        return $this->where('code', $value)->update(['status' => $status]);;
    }
}

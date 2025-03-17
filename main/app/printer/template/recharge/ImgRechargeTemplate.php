<?php

namespace app\common\template\recharge;

use help\DateHelp;
use base\imgs\ImgFont;
use app\common\library\helper;
use app\common\template\BaseTemplate;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model\order\UserRechargeOrder;

/**
 * 图片打印 - 发票模版
 */
class ImgRechargeTemplate  extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create(UserRechargeOrder $order)
    {
        $settingStore = $this->setting[SettingEnum::STORE]['values'];
        $shopName = $settingStore['name'] ?? '';
        
        // 余额日志
        $balanceLog = $order->balanceLog()->where('scene', 10)->find();
         
        /* *
        * 模版1
        */
        $payTime = date('Y/m/d H:i:s', strtotime($order['update_time']));
        if ($this->defaultCalendar == '3') {
            $payTime = DateHelp::changeBuddhistCalendar($payTime);
        }
        /* *
        * 打印模版
        */
        $printer = new ImgFont(568);
        $printer->setTextLineHeight(45);
        $printer->setImagePadding(0);
        // 
        $printer->setAlignment(ImgFont::ALIGN_LEFT);
        $printer->appendText(__("充值单"));
        // 
        $printer->lineFeed();
        $printer->lineFeed(1, 24);
        $printer->setAlignment(ImgFont::ALIGN_CENTER);
        $printer->setFontWeight(2);
        $printer->setFontSize(34);
        $printer->appendText("{$shopName}");
        $printer->setFontSize(20);
        $printer->setFontWeight(1);
        $printer->lineFeed();
        $printer->lineFeed(1, 24);
        // 
        $printer->printInColumns(
            [__("收银员"), 350, ImgFont::ALIGN_LEFT],
            [$order->cashier?->real_name ?: $order->cashier?->user_name ?: '', 0, ImgFont::ALIGN_RIGHT],
        );
        $printer->printInColumns(
            [__("时间"), 350, ImgFont::ALIGN_LEFT],
            [$payTime, 0, ImgFont::ALIGN_RIGHT],
        );
        $printer->printInColumns(
            [__("充值前"), 350, ImgFont::ALIGN_LEFT],
            [$this->getPriceAndUnit($balanceLog?->before_money ?: 0), 0, ImgFont::ALIGN_RIGHT],
        );
        $printer->printInColumns(
            [__("本次充值"), 350, ImgFont::ALIGN_LEFT],
            [$this->getPriceAndUnit($order['recharge_money']), 0, ImgFont::ALIGN_RIGHT],
        );
        $printer->printInColumns(
            [__("赠送金额"), 350, ImgFont::ALIGN_LEFT],
            [$this->getPriceAndUnit($order['gift_money']), 0, ImgFont::ALIGN_RIGHT],
        );
        $printer->printInColumns(
            [__("赠送积分"), 350, ImgFont::ALIGN_LEFT],
            [helper::amountPermillage($order['gift_point']) . '', 0, ImgFont::ALIGN_RIGHT],
        );
        $printer->setTextLineHeight(35);
        $printer->printInColumns(
            [__("充值后"), 350, ImgFont::ALIGN_LEFT],
            [$this->getPriceAndUnit($balanceLog?->after_money ?: 0), 0, ImgFont::ALIGN_RIGHT],
        );
        $printer->setTextLineHeight(35);
        // 退款
        if ($order['refund_money'] && $order['refund_money'] > 0) {
            $printer->appendSplitline(true);
            $printer->printInColumns(
                [__("退款"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($order['refund_money']), 0, ImgFont::ALIGN_RIGHT],
            );
        }
        // 合计應收：
        $printer->appendSplitline(true);
        $printer->lineFeed(1, 14);
        $printer->printInColumns(
            [__("合计应收："), 350, ImgFont::ALIGN_LEFT, 2, 28],
            [$this->getPriceAndUnit($order['actual_price']), 0, ImgFont::ALIGN_RIGHT, 2, 28],
        );
        $printer->appendSplitline(true);
        $printer->setTextLineHeight(45);
        // 支付方式
        $payTypes = $order->payType()->select()->append(['pay_type'])->toArray();
        foreach ($payTypes as $payType) {
            $printer->printInColumns(
                [__("支付方式"), 280, ImgFont::ALIGN_LEFT],
                [$payType['pay_type']['text'] . ($this->lang == 'my' ? ' '  : ''), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("实收金额"), 280, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($payType['price']), 0, ImgFont::ALIGN_RIGHT],
            );
            if ($payType['pay_type']['value'] == OrderPayTypeEnum::CASH) {
                $printer->printInColumns(
                    [__("找零"), 280, ImgFont::ALIGN_LEFT],
                    [helper::amountPermillage($order['change_due'] ?? 0) . '', 0, ImgFont::ALIGN_RIGHT],
                );
            }
        }
        // 
        $printer->lineFeed(1);
        $printer->lineFeed(1, 20);
        // 
        return $printer->save('', !$this->isSunmi);
    }
}

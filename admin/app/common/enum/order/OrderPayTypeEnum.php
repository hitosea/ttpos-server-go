<?php

namespace app\common\enum\order;

use MyCLabs\Enum\Enum;

/**
 * 订单支付方式枚举类
 */
class OrderPayTypeEnum extends Enum
{
    // 免单
    const FREE_PAY = -1;

    // 余额支付
    const BALANCE = 10;

    // 现金支付
    const CASH = 40;

    // 微信支付
    const WECHAT = 20;

    // 支付宝支付
    const ALIPAY = 30;

    // 自有微信
    const OWECHAT = 50;

    // 自有支付宝
    const OALIPAY = 60;

    // 自有POS刷卡
    const POS = 70;

    // QR PromptPay
    const QRPROMPTPAY = 80;

    // QR code
    const QRCODE = 90;

    // SCB easy
    const SCBEASY = 100;

    // Krungthai NEXT
    const KRUNGTHAINEXT = 110;

    // Krungsri Mobile
    const KRUNGSRIMOBILE = 120;

    // Cross-Border QR
    const CROSSBORDERQR = 130;

    // TrueMoneyWallet
    const TRUEMONEYWALLET = 140;

    // LINE Pay
    const LINEPAY = 150;

    // ja  credit card
    const JACREDITCARD = 160;

    // ja  credit card
    const JAICTRAFFICCARD = 170;

    // JA QRCODE
    const JAQRCODE = 180;

    // JA QRCODE
    const JAQCREDITDEBIT = 190;

    const LIANLIAN_WECHAT_PAY = 90111;

    const LIANLIAN_ALI_PAY = 90222;

    const LIANLIAN_QR_PROMPT_PAY = 90333;

    /**
     * Grab 和 LINE MAN 支付方式
     */
    const GRAB = 91100;
    const LINE_MAN = 91200;

    /**
     * Free Meal for ERP（用于ERP同步的免单支付方式）
     */
    const FREE_MEAL_FOR_ERP = 92000;

    const KBANK_ALIPAY = 93000;
    const KBANK_WECHAT = 93100;
    const KBANK_CREDIT_QR = 93200;
    const KBANK_THAI_QR = 93300;
    const KBANK_CREDIT_CARD = 93400;
    
    /**
     * 获取枚举数据
     */
    public static function data($key = 0, $type = 1)
    {
        $arr = [
            self::KBANK_ALIPAY => [
                'name' => "Alipay",
                'value' => self::KBANK_ALIPAY,
                'status' => 1,
                'sort' => 0,
                'img' => '/image/pay/alipay.png',
                'source' => 3,
                'can_add' => true,
            ],
            self::KBANK_WECHAT => [
                'name' => "WeChatPay",
                'value' => self::KBANK_WECHAT,
                'status' => 1,
                'sort' => 1,
                'img' => '/image/pay/wechat_pay.png',
                'source' => 3,
                'can_add' => true,
            ],
            self::KBANK_CREDIT_QR => [
                'name' => "Credit QR",
                'value' => self::KBANK_CREDIT_QR,
                'status' => 1,
                'sort' => 2,
                'img' => '/image/pay/credit_qr.png',
                'source' => 3,
                'can_add' => true,
            ],
            self::KBANK_THAI_QR => [
                'name' => "Thai QR",
                'value' => self::KBANK_THAI_QR,
                'status' => 1,
                'sort' => 3,
                'img' => '/image/pay/thai_qr.png',
                'source' => 3,
                'can_add' => true,
            ],
            self::KBANK_CREDIT_CARD => [
                'name' => "Credit Card",
                'value' => self::KBANK_CREDIT_CARD,
                'status' => 1,
                'sort' => 4,
                'img' => '/image/pay/credit_card.png',
                'source' => 3,
                'can_add' => true,
            ],
            self::JACREDITCARD => [
                'name' => "クレジットカード",
                'value' => self::JACREDITCARD,
                'status' => 1,
                'sort' => 5,
                'img' => '/image/pay/ja_pay.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::JAICTRAFFICCARD => [
                'name' => "IC交通卡",
                'value' => self::JAICTRAFFICCARD,
                'status' => 1,
                'sort' => 6,
                'img' => '/image/pay/ja_pay.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::JAQRCODE => [
                'name' => "QRコード",
                'value' => self::JAQRCODE,
                'status' => 1,
                'sort' => 7,
                'img' => '/image/pay/ja_pay.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::QRPROMPTPAY => [
                'name' => "QR PromptPay",
                'value' => self::QRPROMPTPAY,
                'status' => 1,
                'sort' => 8,
                'img' => '/image/pay/qr_prompt_pay.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::QRCODE => [
                'name' => "QR Code",
                'value' => self::QRCODE,
                'status' => 0,
                'sort' => 9,
                'img' => '/image/pay/qr_code.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::SCBEASY => [
                'name' => "SCB EASY",
                'value' => self::SCBEASY,
                'status' => 1,
                'sort' => 10,
                'img' => '/image/pay/scb_easy.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::KRUNGTHAINEXT => [
                'name' => "Krungthai NEXT",
                'value' => self::KRUNGTHAINEXT,
                'status' => 0,
                'sort' => 11,
                'img' => '/image/pay/krungthai_next.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::KRUNGSRIMOBILE => [
                'name' => "Krungsri Mobile",
                'value' => self::KRUNGSRIMOBILE,
                'status' => 1,
                'sort' => 12,
                'img' => '/image/pay/krungsri_mobile.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::CROSSBORDERQR => [
                'name' => "Cross-Border QR",
                'value' => self::CROSSBORDERQR,
                'status' => 0,
                'sort' => 13,
                'img' => '/image/pay/cross_border_qr.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::TRUEMONEYWALLET => [
                'name' => "TrueMoney",
                'value' => self::TRUEMONEYWALLET,
                'status' => 1,
                'sort' => 14,
                'img' => '/image/pay/true_money.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::LINEPAY => [
                'name' => "LINE Pay",
                'value' => self::LINEPAY,
                'status' => 0,
                'sort' => 15,
                'img' => '/image/pay/line_pay.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::OALIPAY => [
                'name' => 'Alipay',         //线下支付(自有支付宝支付)
                'value' => self::OALIPAY,
                'status' => 1,
                'sort' => 16,
                'img' => '/image/pay/alipay.png',
                'source' => 1,
                'can_add' => true,
            ],
            self::OWECHAT => [
                'name' => 'WeChat Pay',     // 线下支付(自有微信支付)
                'value' => self::OWECHAT,
                'status' => 0,
                'sort' => 17,
                'img' => '/image/pay/wechat_pay.png',
                'source' => 1,
                'can_add' => true,
            ],
            //
            self::JAQCREDITDEBIT => [
                'name' => 'Credit/Debit',
                'value' => self::JAQCREDITDEBIT,
                'status' => 1,
                'sort' => 18,
                'img' => '/image/pay/ja_pay.png',
                'source' => 1,
                'can_add' => true,
            ],
            //
            self::BALANCE => [
                'name' => $type == 1 ? __('余额支付') : __('余额收入'),
                'value' => self::BALANCE,
                'status' => 1,
                'sort' => 19,
                'img' => '/image/pay/ja_pay.png',
                'source' => 0,
                'can_add' => true,
            ],
            self::CASH => [
                'name' => $type == 1 ? __('现金支付') : __('现金收入'),
                'value' => self::CASH,
                'status' => 1,
                'sort' => 20,
                'img' => '/image/pay/ja_pay.png',
                'source' => 0,
                'can_add' => true,
            ],
            self::WECHAT => [
                'name' => $type == 1 ? __('微信支付') : __('微信收入'),
                'value' => self::WECHAT,
                'status' => 0,
                'sort' => 21,
                'img' => '/image/pay/wechat_pay.png',
                'source' => 0,
                'can_add' => true,
            ],
            self::ALIPAY => [
                'name' => $type == 1 ? __('支付宝支付') : __('支付宝收入'),
                'value' => self::ALIPAY,
                'status' => 0,
                'sort' => 22,
                'img' => '/image/pay/alipay.png',
                'source' => 0,
                'can_add' => true,
            ],
            self::POS => [
                'name' => $type == 1 ? __('POS刷卡支付') : __('POS刷卡收入'),
                'value' => self::POS,
                'status' => 0,
                'sort' => 23,
                'img' => '/image/pay/ja_pay.png',
                'source' => 0,
                'can_add' => true,
            ],
            self::FREE_PAY => [
                'name' => __('免单'),
                'value' => self::FREE_PAY,
                'status' => 1,
                'sort' => 24,
                'img' => '',
                'source' => 0,
                'can_add' => true,
            ],

        ];

        return $key != 0 ? $arr[$key] ?? '' : $arr;
    }

    /**
     * 获取支付枚举数据
     */
    public static function pay()
    {
        return  array_filter(self::data(), function ($item) {
            switch ($item['value']) {
                case self::BALANCE:
                case self::WECHAT:
                case self::ALIPAY:
                    return true;
                    break;
            }
            return false;
        });
    }
}

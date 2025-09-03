import { languageStore } from '@/store/model/language';
const languageKey = languageStore().getLanguageKey().language;
// 导出预览数据
export const previewData = {
    one: [
        [
            {
                name: $t('当班编号'),
                label: 2024012536958425,
            },
            {
                name: $t('交班人'),
                label: $t('张三'),
            },
            {
                name: $t('当班时间'),
                label: '2023/12/15 14:00:21' + $t('至') + '2023/12/15 14:00:21',
            },
        ],
        [
            {
                name: $t('总销售额'),
                label: '￥150,125.00',
            },
            {
                name: $t('原商品金额'),
                label: '￥500.00',
            },
            {
                name: $t('支付手续费'),
                label: '￥300.00',
            },
            {
                name: $t('服务费'),
                label: '￥500.00',
            },
            {
                name: $t('税费'),
                label: '￥50.00',
            },
            {
                name: $t('商品数量'),
                label: '128',
            },
            {
                name: $t('优惠折扣'),
                label: '￥839.00',
            },
            {
                name: $t('会员折扣'),
                label: '￥597.00',
            },
            {
                name: $t('退款金额'),
                label: '￥50',
            },
            {
                name: $t('赠菜金额'),
                label: '￥50',
            },
            {
                name: $t('免单金额'),
                label: '￥50',
            },
        ],
        [
            {
                name: $t('实收金额'),
                label: '￥150,125.00',
                big: true,
            },
            {
                name: $t('VAT（10%）'),
                label: '',
                font700: true,
            },
            {
                name: $t('合计'),
                label: '￥68,181.00',
            },
            {
                name: '',
                label: '(' + $t('其中VAT') + '￥1,524.00)',
            },
            {
                name: $t('VAT（8%）'),
                label: '',
                font700: true,
            },
            {
                name: $t('合计'),
                label: '￥68,295.45',
            },
            {
                name: '',
                label: '(' + $t('其中VAT') + '￥1,524.00)',
            },
        ],
        [
            {
                name: $t('合计'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('所有订单数'),
                label: '52',
            },
            {
                name: $t('取消订单数'),
                label: '52',
            },
            {
                name: $t('桌数'),
                label: '52',
            },
            {
                name: $t('人数'),
                label: '52',
            },
            {
                name: $t('最小/大订单金额'),
                label: '￥28.15 / ￥5246.12',
            },
            {
                name: $t('取消订单金额'),
                label: '￥28.15',
            },
            {
                name: $t('平均订单金额'),
                label: '￥689.12',
            },
            {
                name: $t('桌台方式'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('订单数（桌数）'),
                label: '52',
            },
            {
                name: $t('人数'),
                label: '52',
            },
            {
                name: $t('最小/大订单金额'),
                label: '￥28.15 / ￥5246.12',
            },
            {
                name: $t('平均订单金额'),
                label: '￥689.12',
            },
            {
                name: $t('人均'),
                label: '￥689.12',
            },
            {
                name: $t('点餐方式'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('订单数'),
                label: '52',
            },
            {
                name: $t('最小/大订单金额'),
                label: '￥28.15 / ￥5246.12',
            },
            {
                name: $t('平均订单金额'),
                label: '￥689.12',
            },
        ],
        [
            {
                name: $t('支付方式'),
                num: $t('订单数'),
                label: $t('金额'),
                font500: true,
            },
            {
                name: $t('现金'),
                num: '68',
                label: '￥520.00',
            },
            {
                name: 'Krungsri Mobile',
                num: '50',
                label: '￥520.00',
            },
            {
                name: 'Cross-Border QR',
                num: '55',
                label: '￥520.00',
            },
            {
                name: 'TrueMoney',
                num: '35',
                label: '￥520.00',
            },
            {
                name: 'LINE Pay',
                num: '47',
                label: '￥520.00',
            },
            {
                name: 'Alipay',
                num: '32',
                label: '￥520.00',
            },
            {
                name: 'WeChat Pay',
                num: '28',
                label: '￥520.00',
            },
            {
                name: $t('总金额'),
                label: '￥150,1.00',
            },
        ],
        [
            {
                name: $t('高峰时间'),
                num: $t('订单数'),
                label: $t('金额'),
                font500: true,
            },
            {
                name: '5/31 17:00-18:00',
                num: '68',
                label: '￥520.00',
            },
            {
                name: '5/31 17:00-18:00',
                num: '50',
                label: '￥520.00',
            },
        ],
        [
            {
                name: $t('分类'),
                num: $t('数量'),
                label: $t('小计'),
                font500: true,
            },
            {
                name: $t('肉类'),
                num: '68',
                label: '￥50.00',
            },
            {
                name: $t('蔬菜类'),
                num: '50',
                label: '￥50.00',
            },
            {
                name: $t('小计'),
                num: '50',
                label: '￥50.00',
            },
        ],
        [
            {
                name: $t('本班实收金额'),
                label: '￥150,125.00',
            },
            {
                name: $t('上一班遗留备用金'),
                label: '￥250.00',
            },
            {
                name: $t('中途存入现金'),
                label: '￥50.00',
            },
            {
                name: $t('中途取出现金'),
                label: '￥50.00',
            },
            {
                name: $t('本班取出现金'),
                label: '￥387.00',
            },
            {
                name: $t('本班遗留备用金'),
                label: '￥672.00',
            },
        ],
    ],
    two: [
        [
            {
                name: $t('桌位: A01 (4人)'),
                label: '',
                hide: 2,
                font24: true,
                right: false,
            },
            {
                name: $t('这是桌台备注，非桌台/桌台没有备注的则不显示，需要换行显示'),
                label: '',
                big: true,
                typeShow: '3,4,5',
                right: false,
            },
            {
                name: $t('收银员'),
                label: $t('张三'),
            },
            {
                name: $t('时间'),
                label: '2023/12/15 14:00:21',
                hide: 2,
            },
            {
                name: $t('订单号'),
                label: 2024012536958425,
            },
        ],
        [
            {
                name: $t('商品'),
                num: $t('单价 | 数量'),
                label: $t('小计'),
                bold: true,
                typeShow: '1,2',
            },
        ],
        [
            {
                name: $t('（赠）') + $t('商品名称商品名称品名称商品名称商品名称品名称') + `(${$t('少冰、')}${$t('加珍珠')})`,
                num: '24*12',
                label: '￥350',
                typeShow: '1,2,3,5',
            },
            {
                name: $t('（赠）') + $t('商品名称商品名称品名称商品名称商品名称品名称'),
                num: '24*12',
                label: '￥350',
                typeShow: '4',
            },
            {
                name: $t('（打包）') + $t('（赠）') + $t('套餐名称套餐名称'),
                num: '55*2',
                label: '￥50',
            },

            {
                name: '-' + $t('套餐商品1'),
                num: '',
                label: '',
            },
            {
                name: $t('少冰'),
                num: '',
                label: '',
                typeShow: '1,2,3,5',
            },
            {
                name: '-' + $t('套餐商品2'),
                num: '',
                label: '',
            },
        ],
        [
            {
                name: '',
                label: $t('商品数量') + $t('：') + '14',
                typeShow: '1,2',
            },
            {
                name: '',
                label: $t('商品金额') + $t('：') + '￥410',
                typeShow: '1,2',
            },
            {
                name: $t('商品数量') + $t('：') + '14',
                label: $t('商品金额') + $t('：') + '￥410',
                typeShow: '3,4,5',
            },
            {
                name: '',
                label: $t('服务费') + $t('：') + '￥50',
            },
            {
                name: '',
                label: $t('VAT（10%）') + $t('：') + '￥50',
                typeShow: '2',
            },
            {
                name: '',
                label: $t('VAT（8%）') + $t('：') + '￥50',
                typeShow: '2',
            },
            {
                name: '',
                label: $t('优惠折扣') + $t('：') + '￥50',
                typeShow: '1,2',
                left: false,
            },
            {
                name: '',
                label: $t('优惠折扣') + $t('：') + '￥50（4.28% OFF）',
                typeShow: '3,4,5',
                left: false,
            },
            {
                name: '',
                label: $t('会员优惠') + $t('：') + '￥3',
                typeShow: '1,2,3,4',
                left: false,
            },
            {
                name: '',
                label: $t('会员折扣') + $t('：') + $t('3.8折'),
                typeShow: '3,4,5',
                left: false,
            },
            {
                name: '',
                label: $t('会员卡折扣') + $t('：') + $t('3.8折'),
                typeShow: '3,4,5',
                left: false,
            },
            {
                name: '',
                label: $t('免单') + $t('：') + '￥84.89',
                typeShow: '2',
                left: false,
            },
            {
                name: '',
                label: $t('退款金额') + $t('：') + '￥3',
                typeShow: '1,3,4,5',
                left: false,
            },
            {
                name: '',
                label: $t('支付手续费') + $t('：') + '￥3',
                typeShow: '1,3,4,5',
                left: false,
            },
            {
                name: '',
                label: $t('手动抹零') + $t('：') + '￥1.15',
                typeShow: '1,2,3,4,5',
                left: false,
            },
            {
                typeShow: '1',
                name: $t('合计应收'),
                label: '￥410',
                big: true,
            },
            {
                typeShow: '2',
                name: $t('合计应收'),
                label: '￥0',
                big: true,
            },
        ],
        [
            {
                allHide: '3,4,5',
                name: $t('合计应收'),
                label: '￥410',
                big: true,
            },
        ],
        [
            {
                name: '',
                label: $t('合计(其中VAT)'),
                allHide: '1,3,4,5',
            },
            {
                name: $t('VAT（10%）'),
                label: '100.00 (9.09)',
                typeShow: '1,3,4,5',
            },
            {
                name: $t('VAT（8%）'),
                label: '98.90 (7.36)',
                typeShow: '1,3,4,5',
            },
            {
                name: $t('VAT（10%）'),
                label: '￥100.00 (￥9.09)',
                typeShow: '2',
            },
            {
                name: $t('VAT（8%）'),
                label: '￥98.90 (￥7.36)',
                typeShow: '2',
            },
        ],
        [
            {
                name: $t('支付方式'),
                label: $t('现金'),
                typeShow: '1',
            },
            {
                name: $t('实收金额'),
                label: '￥31',
                typeShow: '1',
            },
            {
                name: $t('找零'),
                label: '0.02',
                typeShow: '1',
            },
            {
                name: $t('支付方式'),
                label: 'WeChat',
                typeShow: '1',
            },
            {
                name: $t('实收金额'),
                label: '￥40',
                typeShow: '1',
            },
            {
                name: $t('支付方式'),
                label: $t('免单'),
                typeShow: '2',
            },
            {
                name: $t('实收金额'),
                label: '￥0',
                typeShow: '2',
            },
            {
                name: $t('支付方式'),
                label: $t('微信支付'),
                typeShow: '3,4,5',
            },
            {
                name: $t('实收金额'),
                label: '￥410',
                typeShow: '3,4,5',
            },
            {
                name: $t('找零'),
                label: '0.1',
                typeShow: '3,4,5',
            },
        ],
        [
            {
                name: $t('会员剩余余额'),
                label: '￥100',
            },
            {
                name: $t('本次积分'),
                label: '410',
                typeShow: '1,3,4,5',
            },
            {
                name: $t('本次积分'),
                label: '0',
                typeShow: '2',
            },
        ],
    ],
    three: [
        [
            {
                name: $t('桌位: A01 (4人)'),
                label: '',
                hide: 2,
                font24: true,
                right: false,
            },
            {
                name: $t('这是桌台备注，非桌台/桌台没有备注的则不显示，需要换行显示'),
                label: '',
                big: true,
                typeShow: '3,4',
                right: false,
            },
            {
                name: $t('订单号'),
                label: 2024012536958425,
                typeShow: '1,2',
            },
            {
                name: $t('收银员'),
                label: $t('张三'),
                typeShow: '1,2',
            },
            {
                name: $t('收银员：') + $t('张三'),
                label: '',
                typeShow: '3,4',
            },
            {
                name: $t('订单号：') + 2024012536958425,
                label: '',
                typeShow: '3,4',
                right: false,
            },
        ],
        [
            {
                name: $t('商品'),
                num: $t('单价 | 数量'),
                label: $t('小计'),
                bold: true,
                typeShow: '1,2',
            },
        ],
        [
            {
                name: $t('（赠）') + $t('商品名称商品名称品名称商品名称商品名称品名称') + `(${$t('少冰、')}${$t('加珍珠')})`,
                num: '24*12',
                label: '￥50',
                typeShow: '1,2,3',
            },
            {
                name: $t('（赠）') + $t('商品名称商品名称品名称商品名称商品名称品名称'),
                num: '24*12',
                label: '￥50',
                typeShow: '4',
            },
            {
                name: $t('（打包）') + $t('（赠）') + $t('套餐名称套餐名称'),
                num: '55*2',
                label: '￥50',
            },

            {
                name: '-' + $t('套餐商品1'),
                num: '',
                label: '',
            },
            {
                name: $t('少冰'),
                num: '',
                label: '',
                typeShow: '1,2,3',
            },
            {
                name: '-' + $t('套餐商品2'),
                num: '',
                label: '',
            },
        ],
        [
            {
                name: '',
                label: $t('商品数量') + $t('：') + '14',
                typeShow: '1,2',
            },
            {
                name: '',
                label: $t('商品金额') + $t('：') + '￥410',
                typeShow: '1,2',
            },
            {
                name: $t('商品数量') + $t('：') + '14',
                label: $t('商品金额') + $t('：') + '￥410',
                typeShow: '3,4',
            },
            {
                name: '',
                label: $t('服务费') + $t('：') + '￥50',
            },
            {
                name: '',
                label: $t('VAT（10%）') + $t('：') + '￥50',
                typeShow: '2,3,4',
            },
            {
                name: '',
                label: $t('VAT（8%）') + $t('：') + '￥50',
                typeShow: '2,3,4',
            },
            {
                name: '',
                label: $t('优惠折扣') + $t('：') + '￥50',
                typeShow: '1,2',
                left: false,
            },
            {
                name: '',
                label: $t('优惠折扣') + $t('：') + '￥6.99（9.9% OFF）',
                typeShow: '3,4',
                left: false,
            },
            {
                name: '',
                label: $t('会员优惠') + $t('：') + '￥3',
                typeShow: '1,2,3,4',
                left: false,
            },
            {
                name: '',
                label: $t('会员折扣') + $t('：') + $t('3.8折'),
                typeShow: '3,4',
                left: false,
            },
            {
                name: '',
                label: $t('会员卡折扣') + $t('：') + $t('3.8折'),
                typeShow: '3,4',
                left: false,
            },
            {
                name: '',
                label: $t('手动抹零') + $t('：') + '￥1.15',
                typeShow: '1,2,3,4',
                left: false,
            },
            {
                typeShow: '1,2',
                name: $t('合计应收'),
                label: '￥410',
                big: true,
            },
        ],
        [
            {
                allHide: '3,4',
                name: $t('合计应收'),
                label: '￥410',
                big: true,
            },
        ],
        [
            {
                name: '',
                label: $t('合计(其中VAT)'),
                allHide: '1',
            },
            {
                name: $t('VAT（10%）'),
                label: '100.00 (9.09)',
                typeShow: '1',
            },
            {
                name: $t('VAT（8%）'),
                label: '98.90 (7.36)',
                typeShow: '1',
            },
            {
                name: $t('VAT（10%）'),
                label: '￥100.00 (￥9.09)',
                typeShow: '2,3,4',
            },
            {
                name: $t('VAT（8%）'),
                label: '￥98.90 (￥7.36)',
                typeShow: '2,3,4',
            },
        ],
    ],
    five: [
        [
            {
                name: $t('时间'),
                label: '2023/12/15 14:00:21' + $t('至') + '\n' + '2023/12/15 14:00:21',
            },
            {
                name: $t('总销售额'),
                label: '￥150,125.00',
            },
            {
                name: $t('原商品金额'),
                label: '￥500.00',
            },
            {
                name: $t('服务费'),
                label: '￥500.00',
            },
            {
                name: $t('支付手续费'),
                label: '￥300.00',
            },
            {
                name: $t('税费'),
                label: '￥500.00',
            },

            {
                name: $t('商品数量'),
                label: '128',
            },
            {
                name: $t('优惠折扣'),
                label: '￥839.00',
            },
            {
                name: $t('会员折扣'),
                label: '￥597.00',
            },
            {
                name: $t('退款金额'),
                label: '￥50',
            },
            {
                name: $t('赠菜金额'),
                label: '￥50.00',
            },
            {
                name: $t('免单金额'),
                label: '￥500',
            },
            {
                name: $t('实收金额'),
                label: '￥150,125.00',
                big: true,
            },
        ],
        [
            {
                name: $t('VAT（10%）'),
                label: '',
                font700: true,
            },
            {
                name: $t('合计'),
                label: '￥68,181.00',
            },
            {
                name: '',
                label: '(' + $t('其中VAT') + '￥1,524.00)',
            },
            {
                name: $t('VAT（8%）'),
                label: '',
                font700: true,
            },
            {
                name: $t('合计'),
                label: '￥68,295.45',
            },
            {
                name: '',
                label: '(' + $t('其中VAT') + '￥1,524.00)',
            },
        ],
        [
            {
                name: $t('会员数据'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('充值金额'),
                label: '￥500',
            },
            {
                name: $t('赠送金额'),
                label: '￥100',
            },
            {
                name: $t('赠送积分'),
                label: '100',
            },
        ],
        [
            {
                name: $t('未结账数据'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('订单数'),
                label: '52',
            },
            {
                name: $t('金额'),
                label: '￥34',
            },
        ],
        [
            {
                name: $t('合计'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('所有订单数'),
                label: '52',
            },
            {
                name: $t('取消订单数'),
                label: '52',
            },
            {
                name: $t('桌数'),
                label: '52',
            },
            {
                name: $t('人数'),
                label: '52',
            },
            {
                name: $t('最小/大订单金额'),
                label: '￥28.15 / ￥5246.12',
            },
            {
                name: $t('取消订单金额'),
                label: '￥28.15',
            },
            {
                name: $t('平均订单金额'),
                label: '￥689.12',
            },
            {
                name: $t('桌台方式'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('订单数（桌数）'),
                label: '52',
            },
            {
                name: $t('人数'),
                label: '52',
            },
            {
                name: $t('最小/大订单金额'),
                label: '￥28.15 / ￥5246.12',
            },
            {
                name: $t('平均订单金额'),
                label: '￥689.12',
            },
            {
                name: $t('人均'),
                label: '￥689.12',
            },
            {
                name: $t('点餐方式'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('订单数'),
                label: '52',
            },
            {
                name: $t('最小/大订单金额'),
                label: '￥28.15 / ￥5246.12',
            },
            {
                name: $t('平均订单金额'),
                label: '￥689.12',
            },
        ],
        [
            {
                name: $t('支付方式'),
                num: $t('订单数'),
                label: $t('金额'),
                font500: true,
            },
            {
                name: $t('现金'),
                num: '68',
                label: '￥520.00',
            },
            {
                name: 'Krungsri Mobile',
                num: '50',
                label: '￥520.00',
            },
            {
                name: 'Cross-Border QR',
                num: '55',
                label: '￥520.00',
            },
            {
                name: 'TrueMoney',
                num: '35',
                label: '￥520.00',
            },
            {
                name: 'LINE Pay',
                num: '47',
                label: '￥520.00',
            },
            {
                name: 'Alipay',
                num: '32',
                label: '￥520.00',
            },
            {
                name: 'WeChat Pay',
                num: '28',
                label: '￥520.00',
            },
            {
                name: $t('总金额'),
                label: '￥520.00',
            },
        ],
        [
            {
                name: $t('高峰时间'),
                num: $t('订单数'),
                label: $t('金额'),
                font500: true,
            },
            {
                name: '5/31 17:00-18:00',
                num: '68',
                label: '￥520.00',
            },
            {
                name: '5/31 17:00-18:00',
                num: '50',
                label: '￥520.00',
            },
        ],
    ],
    six: [
        [
            {
                name: $t('订单号'),
                label: 2024012536958425,
                font16Small: true,
            },
            {
                name: $t('时间'),
                label: '2023/12/15 14:00:21',
                font16Small: true,
            },
        ],
        [
            {
                name: $t('商品'),
                label: $t('数量'),
                font17Normal: true,
            },
        ],
        [
            {
                name: $t('商品名称商品名称品名称商品名称商品名称品名称'),
                label: 'X2',
                typeShow: '1',
                flexWidth: true,
                font16Normal: true,
            },
            {
                name: `${$t('少冰')}` + '\n' + `${$t('加珍珠')}`,
                typeShow: '1',
                flexWidth: true,
                showSkuAttr: true,
            },
            {
                name: `<span class="font24-400">${$t('这是备注这是备注这是备注')}</span>`,
                typeShow: '1',
                flexWidth: true,
                font18Small: true,
            },
            {
                name: $t('（打包）') + $t('商品名称商品名称'),
                label: 'X2',
                typeShow: '1',
                flexWidth: true,
                font16Normal: true,
            },
            {
                name:  $t('（打包）') + $t('套餐名称套餐名称'),
                label: 'X2',
                typeShow: '1',
                flexWidth: true,
                font16Normal: true,
            },

            {
                name: '-' + $t('套餐商品1'),
                typeShow: '1',
                flexWidth: true,
                showSkuAttr: true,
            },
            {
                name: $t('少冰'),
                typeShow: '1',
                flexWidth: true,
                showSkuAttr: true,
            },
            {
                name: '-' + $t('套餐商品2'),
                typeShow: '1',
                flexWidth: true,
                showSkuAttr: true,
            },
            {
                name: $t('商品名称商品名称品名称商品名称商品名称品名称'),
                label: 'X2',
                font24: true,
                typeShow: '2',
                flexWidth: true,
            },
            {
                name: `${$t('少冰')}` + '\n' + `${$t('加珍珠')}`,
                typeShow: '2',
                flexWidth: true,
                font18Small: true,
                showSkuAttr: true,
            },
            {
                name: `<span class="font24-400">${$t('这是备注这是备注这是备注')}</span>`,
                typeShow: '2',
                flexWidth: true,
                font18Small: true,
            },
            {
                name: $t('（打包）') + $t('商品名称商品名称'),
                label: 'X2',
                font24: true,
                typeShow: '2',
                flexWidth: true,
            },
            {
                name:  $t('（打包）') + $t('套餐名称套餐名称'),
                label: 'X2',
                font24: true,
                typeShow: '2',
                flexWidth: true,
            },

            {
                name: '-' + $t('套餐商品1'),
                typeShow: '2',
                flexWidth: true,
                font18Small: true,
                showSkuAttr: true,
            },
            {
                name: $t('少冰'),
                typeShow: '2',
                flexWidth: true,
                font18Small: true,
                showSkuAttr: true,
            },
            {
                name: '-' + $t('套餐商品2'),
                typeShow: '2',
                flexWidth: true,
                font18Small: true,
                showSkuAttr: true,
            },

            {
                name: $t('商品名称商品名称品名称商品名称商品名称品名称'),
                label: 'X2',
                font24: true,
                typeShow: '3',
                flexWidth: true,
            },
            {
                name: `<span class="grey">${$t('少冰')}</span>` + '\n' + `<span class="grey">${$t('加珍珠')}</span>`,
                typeShow: '3',
                flexWidth: true,
                font24: true,
                right: false,
                showSkuAttr: true,
                lineHeight: '2-5',
            },
            {
                name: `<span class="grey">${$t('这是备注这是备注这是备注')}</span>`,
                typeShow: '3',
                flexWidth: true,
                font24: true,
                right: false,
                lineHeight: '2-5',
            },
            {
                name: $t('（打包）') + $t('商品名称商品名称'),
                label: 'X2',
                font24: true,
                typeShow: '3',
                flexWidth: true,
            },
            {
                name: $t('（打包）') + $t('套餐名称套餐名称'),
                label: 'X2',
                font24: true,
                typeShow: '3',
                flexWidth: true,
            },
            {
                name:`<span class="grey"> -${$t('套餐商品1')}</span>`,
                typeShow: '3',
                flexWidth: true,
                font24: true,
                right: false,
                showSkuAttr: true,
                lineHeight: '2-5',
            },
            {
                name: `<span class="grey"> ${$t('少冰')}</span>`,
                typeShow: '3',
                flexWidth: true,
                font24: true,
                right: false,
                showSkuAttr: true,
                lineHeight: '2-5',
            },
            {
                name: `<span class="grey"> -${$t('套餐商品2')}</span>`,
                typeShow: '3',
                flexWidth: true,
                font24: true,
                right: false,
                showSkuAttr: true,
                lineHeight: '2-5',
            },
        ],
    ],
    seven: [
        [
            {
                name: $t('合计金额'),
                label: '¥4502',
                big: true,
            },
            {
                name: $t('其中服务费'),
                label: '￥20.00)',
            },
            {
                name: $t('其中VAT'),
                label: '￥200.00)',
            },
        ],
        [
            {
                name: $t('仅作为餐饮费收取以上金额'),
                label: '',
            },
            {
                name: '',
                label: $t('合计 (其中VAT)'),
            },
            {
                name: $t('VAT（10%）'),
                label: '4,000.00 (363.65)',
            },
            {
                name: $t('VAT（8%）'),
                label: '501.00 (45.55)',
            },
            {
                name: $t('不包含退款金額￥3.00'),
                label: '',
            },
        ],
        [
            {
                name: $t('现金'),
                label: '¥4502',
                big: true,
            },
        ],
        [
            {
                name: $t('发票信息'),
                label: '',
            },
            {
                name: $t('公司名称：') + $t('公司名称公司名称公司名称'),
                label: '',
            },
            {
                name: $t('公司地址：') + $t('公司地址公司地址公司地址公司地址'),
                label: '',
            },
            {
                name: $t('税号：') + '252452524144',
                label: '',
            },
            {
                name: $t('联系电话：') + '02-15-1441414',
                label: '',
            },
        ],
        [
            {
                name: $t('收银员：张三'),
                label: '',
            },
            {
                name: $t('订单号：') + 'NO.252452524144',
                label: '',
            },
            {
                name: $t('打印次数：1'),
                label: '',
            },
            {
                name: $t('公司名称：') + $t('公司名称公司名称公司名称'),
                label: '',
            },
            {
                name: $t('商家地址：') + $t('商家地址商家地址商家地址商家地址商家地址商家地址商家地址商家地址'),
                label: '',
            },
            {
                name: $t('税号：') + '252452524144',
                label: '',
            },
            {
                name: $t('电话：') + '02-15-1441414',
                label: '',
            },
            // languageKey == "ja"的时候才才插入这个对象 ,
            languageKey.value == 'ja'
                ? {
                    name: ' ',
                    label: '担当者',
                    img: true,
                    big: true,
                    left: false,
                }
                : '',
        ],
        [
            {
                name: $t('*保管注意事項'),
                label: '',
            },
            {
                name: $t('如需保管时请将印刷页面朝内摺叠'),
                label: '',
            },
        ],
    ],
    eight: [
        [
            {
                name: $t('桌位: A01 (4人)'),
                label: '',
                font24: true,
                right: false,
            },
            {
                name: $t('订单号'),
                label: 2024012536958425,
            },
            {
                name: $t('收银员'),
                label: $t('张三'),
            },
            {
                name: $t('时间'),
                label: '2023/12/15 14:00:21',
            },
        ],

        [
            {
                name: $t('（赠）') + $t('商品名称商品名称品名称商品名称商品名称品名称'),
                num: '24*12',
                label: '￥350',
                addLabel: `(${$t('少冰、')}${$t('加珍珠')})`,
            },

            {
                name: $t('（打包）') + $t('（赠）') + $t('套餐名称套餐名称'),
                num: '55*2',
                label: '￥50',
            },

            {
                name: '-' + $t('套餐商品1'),
                num: '',
                label: '',
            },
            {
                name: $t('少冰'),
                num: '',
                label: '',
            },
            {
                name: '-' + $t('套餐商品2'),
                num: '',
                label: '',
            },
        ],
        [
            {
                name: $t('商品数量') + $t('：') + '14',
                label: $t('商品金额') + $t('：') + '￥410',
            },
            {
                name: '',
                label: $t('服务费') + $t('：') + '￥50',
            },

            {
                name: '',
                label: $t('优惠折扣') + $t('：') + '￥6.99（9.9% OFF）',
            },
            {
                name: '',
                label: $t('会员优惠') + $t('：') + '￥3',
            },
            {
                name: '',
                label: $t('会员折扣') + $t('：') + $t('3.8折'),
            },
            {
                name: '',
                label: $t('会员卡折扣') + $t('：') + $t('3.8折'),
            },
            {
                name: '',
                label: $t('免单') + $t('：') + '￥73.98',
            },
            {
                name: '',
                label: $t('手动抹零') + $t('：') + '￥1.15',
            },
        ],
        [
            {
                name: $t('合计应收'),
                label: '￥410',
                big: true,
            },
        ],
        [
            {
                name: '',
                label: $t('合计(其中VAT)'),
            },
            {
                name: $t('VAT（10%）'),
                label: '100.00 (9.09)',
            },
            {
                name: $t('VAT（8%）'),
                label: '98.90 (7.36)',
            },
        ],
        [
            {
                name: $t('支付方式'),
                label: $t('免单'),
            },
            {
                name: $t('实收金额'),
                label: '￥0',
            },
        ],
        [
            {
                name: $t('会员剩余余额'),
                label: '￥100',
            },
            {
                name: $t('本次积分'),
                label: '410',
            },
        ],
        [
            {
                name: $t('发票信息'),
                label: '',
            },
            {
                name: $t('公司名称：') + $t('公司名称公司名称公司名称'),
                label: '',
            },
            {
                name: $t('公司地址：') + $t('公司地址公司地址公司地址公司地址'),
                label: '',
            },
            {
                name: $t('税号：') + '252452524144',
                label: '',
            },
            {
                name: $t('联系电话：') + '02-15-1441414',
                label: '',
            },
        ],
    ],
    nine: [
        [
            {
                name: $t('当班编号'),
                label: 2024012536958425,
            },
            {
                name: $t('交班人'),
                label: $t('张三'),
            },
            {
                name: $t('总销售额'),
                label: '￥150,125.00',
            },
            {
                name: $t('实收金额'),
                label: '￥150,125.00',
            },
        ],
        [
            {
                name: $t('支付方式'),
                num: $t('订单数'),
                label: $t('金额'),
                font500: true,
            },
            {
                name: $t('现金'),
                num: '68',
                label: '￥520.00',
            },
            {
                name: 'Krungsri Mobile',
                num: '50',
                label: '￥520.00',
            },
            {
                name: 'Cross-Border QR',
                num: '55',
                label: '￥520.00',
            },
            {
                name: 'TrueMoney',
                num: '35',
                label: '￥520.00',
            },
            {
                name: 'LINE Pay',
                num: '47',
                label: '￥520.00',
            },
            {
                name: 'Alipay',
                num: '32',
                label: '￥520.00',
            },
            {
                name: 'WeChat Pay',
                num: '28',
                label: '￥520.00',
            },
            {
                name: $t('总金额'),
                label: '￥150,1.00',
            },
        ],
        [
            {
                name: $t('原商品金额'),
                label: '￥500.00',
            },
            {
                name: $t('支付手续费'),
                label: '￥300.00',
            },
            {
                name: $t('服务费'),
                label: '￥500.00',
            },
            {
                name: $t('税费'),
                label: '￥50.00',
            },
        ],
        [
            {
                name: $t('优惠折扣'),
                label: '￥839.00',
            },
            {
                name: $t('会员折扣'),
                label: '￥597.00',
            },
            {
                name: $t('赠菜金额'),
                label: '￥50.00',
            },
            {
                name: $t('免单金额'),
                label: '￥50.00',
            },
        ],
        [
            {
                name: $t('退款金额'),
                label: '￥50',
            },
        ],
        [
            {
                name: $t('会员数据'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('充值金额'),
                label: '￥500',
            },
            {
                name: $t('赠送金额'),
                label: '￥100',
            },
            {
                name: $t('赠送积分'),
                label: '100',
            },
        ],
        [
            {
                name: $t('所有订单数'),
                label: '52',
            },
            {
                name: $t('取消订单数'),
                label: '52',
            },
            {
                name: $t('人数'),
                label: '52',
            },
            {
                name: $t('取消订单金额'),
                label: '￥28.15',
            },
            {
                name: $t('平均订单金额'),
                label: '￥689.12',
            },
        ],
        [
            {
                name: $t('高峰时间'),
                num: $t('订单数'),
                label: $t('金额'),
                font500: true,
            },
            {
                name: '5/31 17:00-18:00',
                num: '68',
                label: '￥520.00',
            },
            {
                name: '5/31 17:00-18:00',
                num: '50',
                label: '￥520.00',
            },
        ],
        [
            {
                name: $t('分类'),
                num: $t('数量'),
                label: $t('小计'),
                font500: true,
            },
            {
                name: $t('肉类'),
                num: '68',
                label: '￥50.00',
            },
            {
                name: $t('蔬菜类'),
                num: '50',
                label: '￥50.00',
            },
            {
                name: $t('小计'),
                num: '50',
                label: '￥50.00',
            },
        ],
        [
            {
                name: $t('上一班遗留备用金'),
                label: '￥250.00',
            },
            {
                name: $t('中途存入现金'),
                label: '￥50.00',
            },
            {
                name: $t('中途取出现金'),
                label: '￥50.00',
            },
            {
                name: $t('本班取出现金'),
                label: '￥387.00',
            },
            {
                name: $t('本班遗留备用金'),
                label: '￥672.00',
            },
        ],
    ],
    ten: [
        [
            {
                name: $t('收银员'),
                label: $t('张三'),
            },
            {
                name: $t('时间'),
                label: '2023-12-15 14:45:21',
            },

            {
                name: $t('充值前'),
                label: '￥50',
            },
            {
                name: $t('本次充值'),
                label: '￥500',
            },
            {
                name: $t('赠送金额'),
                label: '￥50',
            },
            {
                name: $t('赠送积分'),
                label: '500',
            },
            {
                name: $t('充值后'),
                label: '￥550',
            },
        ],
        [
            {
                name: $t('退款'),
                label: '￥1500',
            },
        ],
        [
            {
                name: $t('合计应收'),
                label: '￥500',
                big: true,
            },
        ],
        [
            {
                name: $t('支付方式'),
                label: $t('现金'),
            },
            {
                name: $t('实收金额'),
                label: '￥100',
            },
            {
                name: $t('找零'),
                label: '5',
            },
            {
                name: $t('支付方式'),
                label: 'Wechat',
            },
            {
                name: $t('实收金额'),
                label: '￥400',
            },
        ],
    ],
    eleven: [
        [
            {
                name: $t('订单号'),
                label: 2024012536958425,
                font16Small: true,
            },
            {
                name: $t('时间'),
                label: '2023/12/15 14:00:21',
                font16Small: true,
            },
        ],
        [
            {
                name: $t('商品'),
                label: $t('数量'),
                font17Normal: true,
            },
        ],
        [
            {
                name: `(${$t('退')})` + $t('商品名称商品名称品名称商品名称商品名称品名称'),
                label: 'X5',
                font22: true,
                flexWidth: true,
            },
            {
                name: `${$t('少冰')}` + '\n' + `${$t('加珍珠')}` + '\n' + $t('这是备注这是备注这是备注'),
                font18Small: true,
                flexWidth: true,
            },
            {
                name: `(${$t('退')})` + $t('商品名称商品名称'),
                label: 'X2',
                font22: true,
                flexWidth: true,
            },
            {
                name: `(${$t('退')})` + $t('套餐名称套餐名称'),
                label: 'X2',
                font22: true,
                flexWidth: true,
            },

            {
                name: '-' + $t('套餐商品1'),
                font18Small: true,
                flexWidth: true,
            },
            {
                name: $t('少冰'),
                font18Small: true,
                flexWidth: true,
            },
            {
                name: '-' + $t('套餐商品2'),
                font18Small: true,
                flexWidth: true,
            },
        ],
        [
            {
                name: $t('退菜原因：等待时间长，口味不好'),
                label: '',
                right: false,
                font18Small: true,
                flexWidth: 3,
            },
        ],
    ],
    twelve: [
        [
            {
                name: $t('当班编号'),
                label: 2024012536958425,
            },
            {
                name: $t('交班人'),
                label: $t('张三'),
            },
            {
                name: $t('总销售额'),
                label: '￥150,125.00',
            },
            {
                name: $t('实收金额'),
                label: '￥150,125.00',
            },
        ],
        [
            {
                name: $t('支付方式'),
                num: $t('订单数'),
                label: $t('金额'),
                font500: true,
            },
            {
                name: $t('现金'),
                num: '68',
                label: '￥520.00',
            },
            {
                name: 'Krungsri Mobile',
                num: '50',
                label: '￥520.00',
            },
            {
                name: 'Cross-Border QR',
                num: '55',
                label: '￥520.00',
            },
            {
                name: 'TrueMoney',
                num: '35',
                label: '￥520.00',
            },
            {
                name: 'LINE Pay',
                num: '47',
                label: '￥520.00',
            },
            {
                name: 'Alipay',
                num: '32',
                label: '￥520.00',
            },
            {
                name: 'WeChat Pay',
                num: '28',
                label: '￥520.00',
            },
            {
                name: $t('总金额'),
                label: '￥150,1.00',
            },
        ],
        [
            {
                name: $t('原商品金额'),
                label: '￥500.00',
            },
            {
                name: $t('支付手续费'),
                label: '￥300.00',
            },
            {
                name: $t('服务费'),
                label: '￥500.00',
            },
            {
                name: $t('税费'),
                label: '￥50.00',
            },
        ],
        [
            {
                name: $t('优惠折扣'),
                label: '￥839.00',
            },
            {
                name: $t('会员折扣'),
                label: '￥597.00',
            },
            {
                name: $t('赠菜金额'),
                label: '￥50.00',
            },
            {
                name: $t('免单金额'),
                label: '￥50.00',
            },
        ],
        [
            {
                name: $t('退款金额'),
                label: '￥50',
            },
        ],
        [
            {
                name: $t('退菜次数'),
                label: '5',
            },
            {
                name: $t('退款次数'),
                label: '5',
            },
            {
                name: $t('反结账次数'),
                label: '5',
            },
            {
                name: $t('赠菜次数'),
                label: '5',
            },
            {
                name: $t('免单次数'),
                label: '5',
            },
            {
                name: $t('转菜次数'),
                label: '5',
            },
            {
                name: $t('单品改价次数'),
                label: '5',
            },
            {
                name: $t('整单改价次数'),
                label: '5',
            },
            {
                name: $t('整单折扣次数'),
                label: '5',
            },
            {
                name: $t('整单抹零次数'),
                label: '5',
            },
        ],
        [
            {
                name: $t('会员数据'),
                label: '',
                font700: true,
                textCenter: true,
                right: false,
            },
            {
                name: $t('充值金额'),
                label: '￥500',
            },
            {
                name: $t('赠送金额'),
                label: '￥100',
            },
            {
                name: $t('赠送积分'),
                label: '100',
            },
        ],
        [
            {
                name: $t('所有订单数'),
                label: '52',
            },
            {
                name: $t('取消订单数'),
                label: '52',
            },
            {
                name: $t('人数'),
                label: '52',
            },
            {
                name: $t('取消订单金额'),
                label: '￥28.15',
            },
            {
                name: $t('平均订单金额'),
                label: '￥689.12',
            },
        ],
        [
            {
                name: $t('高峰时间'),
                num: $t('订单数'),
                label: $t('金额'),
                font500: true,
            },
            {
                name: '5/31 17:00-18:00',
                num: '68',
                label: '￥520.00',
            },
            {
                name: '5/31 17:00-18:00',
                num: '50',
                label: '￥520.00',
            },
        ],
        [
            {
                name: $t('分类'),
                num: $t('数量'),
                label: $t('小计'),
                font500: true,
            },
            {
                name: $t('肉类'),
                num: '68',
                label: '￥50.00',
            },
            {
                name: $t('蔬菜类'),
                num: '50',
                label: '￥50.00',
            },
            {
                name: $t('小计'),
                num: '50',
                label: '￥50.00',
            },
        ],
        [
            {
                name: $t('上一班遗留备用金'),
                label: '￥250.00',
            },
            {
                name: $t('中途存入现金'),
                label: '￥50.00',
            },
            {
                name: $t('中途取出现金'),
                label: '￥50.00',
            },
            {
                name: $t('本班取出现金'),
                label: '￥387.00',
            },
            {
                name: $t('本班遗留备用金'),
                label: '￥672.00',
            },
        ],
    ],
    thirteen: [
        [
            {
                name: $t('订单号'),
                label: 2024012536958425,
                font16Small: true,
            },
            {
                name: $t('时间'),
                label: '2023/12/15 14:00:21',
                font16Small: true,
            },
        ],
        [
            {
                name: $t('商品'),
                label: $t('数量'),
                font17Normal: true,
            },
        ],
        [
            {
                name: `!!!(${$t('退')})` + $t('商品名称商品名称品名称商品名称商品名称品名称'),
                label: '-5',
                font22: true,
                flexWidth: true,
            },
            {
                name: `${$t('少冰')}` + '\n' + `${$t('加珍珠')}` + '\n' + $t('这是备注这是备注这是备注'),
                font18Small: true,
                flexWidth: true,
            },
            {
                name: `!!!(${$t('退')})` + $t('商品名称商品名称'),
                label: '-2',
                font22: true,
                flexWidth: true,
            },
            {
                name: `!!!(${$t('退')})` + $t('套餐名称套餐名称'),
                label: '-2',
                font22: true,
                flexWidth: true,
            },

            {
                name: '-' + $t('套餐商品1'),
                label: '-1',
                font18Small: true,
                flexWidth: true,
            },
            {
                name: $t('少冰'),
                font18Small: true,
                flexWidth: true,
            },
            {
                name: '-' + $t('套餐商品2'),
                label: '-1',
                font18Small: true,
                flexWidth: true,
            },
        ],
        [
            {
                name: $t('退菜原因：等待时间长，口味不好'),
                label: '',
                right: false,
                font18Small: true,
                flexWidth: 3,
            },
        ],
    ],
    fourteen: [
        [
            {
                name: $t('订单号'),
                label: 2024012536958425,
                font16Small: true,
            },
            {
                name: $t('时间'),
                label: '2023/12/15 14:00:21',
                font16Small: true,
            },
        ],
        [
            {
                name: $t('商品'),
                label: $t('数量'),
                font17Normal: true,
            },
        ],
        [
            {
                name: $t('商品名称商品名称品名称商品名称商品名称品名称'),
                label: 'X5',
                font22: true,
                flexWidth: true,
            },
            {
                name: `${$t('少冰')}` + '\n' + `${$t('加珍珠')}` + '\n' + $t('这是备注这是备注这是备注'),
                font18Small: true,
                flexWidth: true,
            },
            {
                name: $t('（打包）') + $t('套餐') + '-' + $t('商品名称商品名称'),
                label: 'X2',
                font22: true,
                flexWidth: true,
            },
        ],
    ],
    fifteen: [
        [
            {
                name: $t('订单号'),
                label: '2024012536958425',
            },
            {
                name: $t('时间'),
                label: '2023/12/15 14:00:21',
            },
        ],
        [
            {
                name: $t('商品'),
                num: $t('单价 | 数量'),
                label: $t('小计'),
                bold: true,
            },
        ],
        [
            {
                name: $t('（赠）') + $t('商品名称商品名称品名称商品名称商品名称品名称') + `(${$t('少冰、')}${$t('加珍珠')})`,
                num: '24*12',
                label: '￥350',
            },

            {
                name: $t('商品名称商品名称'),
                num: '55*2',
                label: '￥50',
            },
        ],
        [
            {
                name: '',
                label: $t('商品数量') + $t('：') + '14',
            },
            {
                name: '',
                label: $t('商品金额') + $t('：') + '￥410',
            },
        ],
        [
            {
                name: $t('合计应收'),
                label: '￥410',
                big: true,
            },
        ],
        [
            {
                name: '',
                label: $t('合计 (其中VAT)'),
            },
            {
                name: $t('VAT（8%）'),
                label: '501.00 (45.55)',
            },
        ],
        [
            {
                name: $t('会员折扣'),
                label: '￥50',
            },
            {
                name: $t('支付方式'),
                label: $t('微信支付'),
            },
            {
                name: $t('实收金额'),
                label: '￥410',
            },
        ],
        [
            {
                name: $t('客户备注：'),
                label: '',
                big: true,
                right: false,
            },
            {
                name: $t('不要香菜和辣椒！到了就敲门，外卖放在门口就行！！！'),
                label: '',
                font24: true,
                right: false,
            },
        ],
        [
            {
                name: $t('联系人'),
                label: 'MO',
            },
            {
                name: $t('手机号码'),
                label: '138****8000',
            },
            {
                name: $t('收货地址'),
                label: $t('这是收货地址这是收货地址这是收货地址这是收货地址这是收货地址'),
                flexWidthRight: true,
            },
        ],
    ],
};

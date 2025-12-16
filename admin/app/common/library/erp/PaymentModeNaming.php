<?php

namespace app\common\library\erp;

use app\common\model\store\PayType;

/**
 * ERP 支付方式命名规则工具类
 * 
 * 规则：{渠道}-{支付方式}-{序号}-{商家缩写}
 * - 渠道：LianLianPay 或空（自行添加和系统默认）
 * - 支付方式：TTPOS 的 payment_name 字段
 * - 序号：系统默认=0000，自行添加=0001起，LianLianPay=0000起
 * - 商家缩写：CompanyAbbr
 */
class PaymentModeNaming
{
    /**
     * 根据 source 确定 channel
     * 
     * @param int $source 来源：0-系统默认，1-自行添加，2-LianLianPay
     * @return string channel 名称
     */
    public static function getChannelBySource($source)
    {
        if ($source == PayType::SOURCE_LIANLIANPAY) {
            return 'LianLianPay';
        }
        return '';
    }

    /**
     * 构建前缀（用于查询同名支付方式）
     * 
     * @param string $channel 渠道
     * @param string $payType 支付方式名称
     * @return string 前缀
     */
    public static function buildPrefix($channel, $payType)
    {
        $parts = [];
        if (!empty(trim($channel))) {
            $parts[] = $channel;
        }
        if (!empty(trim($payType))) {
            $parts[] = $payType;
        }
        if (!empty($parts)) {
            return implode('-', $parts) . '-';
        }
        return '';
    }

    /**
     * 获取下一个序号
     * 
     * 规则：
     * - 系统默认（source=0）：固定使用 0000
     * - 自行添加（source=1）：从 0001 起递增
     * - LianLianPay（source=2）：从 0000 起递增
     * 
     * @param \think\db\Query $db 数据库连接
     * @param string $prefix 前缀
     * @param int $source 来源
     * @param string $companyAbbr 商家缩写
     * @return int 下一个序号
     */
    public static function getNextSequenceNumber($db, $prefix, $source, $companyAbbr)
    {
        // 系统默认：固定使用 0000
        if ($source == PayType::SOURCE_SYSTEM) {
            return 0;
        }

        // 查询已有同名支付方式（通过 erpnext_payment 字段匹配）
        $pattern = '^' . preg_quote($prefix, '/') . '(\d{4})-' . preg_quote($companyAbbr, '/') . '$';
        $existingMethods = $db->name('payment_method')
            ->where('delete_time', 0)
            ->where('erpnext_payment', 'like', $prefix . '%')
            ->column('erpnext_payment');

        $maxSeq = -1;
        foreach ($existingMethods as $erpnextPayment) {
            if (preg_match('/' . $pattern . '/', $erpnextPayment, $matches)) {
                $seq = intval($matches[1]);
                if ($seq > $maxSeq) {
                    $maxSeq = $seq;
                }
            }
        }

        // 确定下一个序号
        if ($maxSeq < 0) {
            // 没有找到同名支付方式
            if ($source == PayType::SOURCE_LIANLIANPAY) {
                return 0; // LianLianPay 从 0000 起
            } else {
                return 1; // 自行添加从 0001 起
            }
        } else {
            // 找到同名支付方式，递增
            $nextSeq = $maxSeq + 1;
            // 自行添加不能使用 0000（系统默认专用）
            if ($source == PayType::SOURCE_DEFAULT && $nextSeq == 0) {
                $nextSeq = 1;
            }
            return $nextSeq;
        }
    }

    /**
     * 生成 Mode of Payment ID
     * 
     * @param \think\db\Query $db 数据库连接
     * @param array $paymentMethod 支付方式数据（包含 source 和 payment_name）
     * @param string $companyAbbr 商家缩写
     * @return string Mode of Payment ID
     */
    public static function generateModeOfPaymentID($db, $paymentMethod, $companyAbbr)
    {
        // 1. 根据 source 确定 channel
        $channel = self::getChannelBySource($paymentMethod['source'] ?? PayType::SOURCE_DEFAULT);

        // 2. 构建前缀（用于查询同名支付方式）
        $prefix = self::buildPrefix($channel, $paymentMethod['payment_name'] ?? '');

        // 3. 查询已有同名支付方式，确定下一个序号
        $nextSeq = self::getNextSequenceNumber($db, $prefix, $paymentMethod['source'] ?? PayType::SOURCE_DEFAULT, $companyAbbr);

        // 4. 生成完整的 Mode of Payment ID
        $modeOfPaymentID = sprintf('%s%04d-%s', $prefix, $nextSeq, $companyAbbr);

        return $modeOfPaymentID;
    }
}

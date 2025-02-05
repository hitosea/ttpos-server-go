<?php

namespace app\common\enum\sync;

use MyCLabs\Enum\Enum;

/**
 * 同步表的枚举类
 */
class SyncEnum extends Enum
{

    /**
     * 基础表
     */
    const TABLES = [
        //
        'user',
        //
        'shop_user',
        'shop_user_shift_log',
        //
        'erp_supplier',
        //
        'upload_group',
        'upload_file',
        //
        'tax_category',
        //
        'spec',
        'feed',
        'category',
        'product_print_label',
        'product',
        'product_tax',
        'product_attribute',
        'product_feed',
        'product_feed_material',
        'product_image',
        'product_sku',
        'product_sku_material',
        'product_sold_out',
        'product_unit',
        //
        'customer_type',
        //
        'buffet',
        'buffet_tax',
        'buffet_customer',
        'buffet_discount',
        'buffet_discount_rel',
        'buffet_product',
        //
        'setting',
        //
        'delay',
        //
        'erp_purchase_order',
        'erp_purchase_detail',
        'erp_monthly_statistics',
        'erp_inventory_record',
        'erp_damaged_product_record',
        'erp_purchase_operation_log',
        'erp_monthly_product_statistics',
        // 
        'pay_type'
    ];

    /**
     * ERP相关表
     */
    const TABLES_ERP = [
        'erp_purchase_order',
        'erp_purchase_detail',
        'erp_monthly_statistics',
        'erp_inventory_record',
        'erp_damaged_product_record',
        'erp_purchase_operation_log',
    ];

    /**
     * 订单相关表
     */
    const TABLES_ORDER = [
        'order',
        'order_product',
        'order_product_return',
        'order_address',
        'order_buffet',
        'order_buffet_customer',
        'order_buffet_discount',
        'order_delay',
        'order_extract',
        'order_finance',
        'order_pay_type',
        'order_peak_time',
        'order_settled',
        'order_refund',
    ];

    /**
     * 获取数据
     */
    public static function getAllTables()
    {
        return array_merge(self::TABLES, self::TABLES_ORDER);
    }

    /**
     * 获取订单关联数据表
     */
    public static function getOrderTables()
    {
        return array_merge(self::TABLES_ORDER, ['shop_user_shift_log']);
    }
}

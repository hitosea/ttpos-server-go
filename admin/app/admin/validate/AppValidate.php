<?php

namespace app\admin\validate;

use app\admin\model\app\App;
use app\common\model\supplier\Supplier;
use app\common\validate\BaseValidate;


class AppValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule = [
        'app_id' => 'require|checkIdExist',
        'name|商家名称' => 'require|max:100',
        'level|商家等级' => 'require|integer',
        'logo' => 'require',
        'parent_id|上级商家ID' => 'requireIf:level,2|integer',
        'auth_start_time|授权开始时间' => 'require|dateFormat:Y-m-d H:i:s',
        'auth_day|授权时间' => 'require|integer',
        'link_phone|联系电话' => 'require|max:20',
        'user_name|超管邮箱' => 'require|max:64|email',
        'password|超管密码' =>  'require|checkPassword',
        'confirm_password|确认超管密码' => 'requireWith:password|confirm:password',
        'cash_limit|收银机上限' => 'require|integer|between:0,999',
        'kitchen_limit|厨显上限' => 'require|integer|between:0,999',
        'tablet_limit|平板上限' => 'require|integer|between:0,999',
        'sale_stock|进销存' => 'require|in:0,1',
        'reserve|预订' => 'require|in:0,1',
        'store_type|经营模式' => 'require|in:10,20',
        'category_set|商品分类' => 'require|in:10,20',
        'address|地址' => 'max:500',
        'coordinates|经纬度' => 'isCoordinates',
        'minutes|过期时间' => 'require|between:1,999',
        'mac_addr|工控机MAC地址' => 'require',
        'serial_number|工控机序列号' => 'require',
        // v1.0.6
        'is_open_member|是否开启会员' => 'require|in:0,1',
        'is_open_tablet|是否开启平板' => 'require|in:0,1',
        'is_open_scan|是否开启扫码H5' => 'require|in:0,1',
        'is_open_assistant|是否开启点餐助手' => 'require|in:0,1',
        'is_open_kitchen_kds|是否开启后厨KDS' => 'require|in:0,1',
        'is_open_buffet|是否开启自助餐' => 'require|in:0,1',
        'table_limit|桌台上限' => 'require|integer',
        'printer_limit|打印机上限' => 'require|integer',
        'languages|支持语言' => 'require',
        'timezone|时区' => 'require',
        // v1.0.9
        'is_accept_scan_order|是否开启扫码点餐接单' => 'require|in:0,1',
        // v1.1.0
        'is_open_local_print|是否开启本地打印服务' => 'require|in:0,1',
        // v2.0.0
        'is_open_tax|是否开启税务对接' => 'require|in:0,1,2,3,4,5,6,7,8,9,10',
        // v2.10.0
        'enable_table_map|是否启用桌台地图能力' => 'in:0,1',
        'enable_data_management|是否启用数据管理能力' => 'in:0,1',
        // v2.11.0
        'enable_kiosk|是否启用自助点餐机' => 'in:0,1',
        // Grab外卖控制
        'enable_grab_delivery|是否启用Grab外卖' => 'in:0,1',
        // LINE MAN外卖控制
        'enable_lineman_delivery|是否启用LINEMAN外卖' => 'in:0,1',
        // 扫码点餐到店自取
        'is_open_member_instant|是否开启扫码点餐到店自取' => 'in:0,1',
    ];

    protected $message = [
        'app_id.require' => '参数错误',
        'app_id.checkIdExist' => '商家不存在',
        'parent_id.requireIf' => '请选择上级商家',
        'name.max' => '商家名称不能超过100个字符',
        'link_phone.require' => '联系电话不能为空',
        'link_phone.max' => '联系电话不能超过20个字符',
        'user_name.max' => '邮箱长度不能超过64个字符',
        'user_name.email' => '请确认邮箱格式',
        'coordinates.isCoordinates' => '经纬度格式错误',
    ];

    protected $scene = [
        'add' => [
            'name',
            'level',
            'logo',
            'link_phone',
            'user_name',
            'parent_id',
            'password',
            'auth_day',
            'auth_start_time',
            'cash_limit',
            'kitchen_limit',
            'tablet_limit',
            'sale_stock',
            'reserve',
            'category_set',
            'store_type',
            'address',
            // v1.0.6
            // 'mac_addr',
            // 'serial_number',
            'is_open_member',
            'is_open_scan',
            'is_open_assistant',
            'is_open_kitchen_kds',
            'is_open_buffet',
            'table_limit',
            'printer_limit',
            'languages',
            'timezone',
            // v1.0.9
            'is_accept_scan_order',
            // v1.1.0
            'is_open_local_print',
            // v2.0.0
            'is_open_tax',
            // v2.4.0
            'coordinates',
            // v2.10.0
            'enable_table_map',
            'enable_data_management',
            // v2.11.0
            'enable_kiosk',
            // Grab外卖控制
            'enable_grab_delivery',
            // LINE MAN外卖控制
            'enable_lineman_delivery',
            // 扫码点餐到店自取
            'is_open_member_instant',
        ],
        'edit' => [
            'app_id',
            'name',
            'level',
            'link_phone',
            'user_name',
            'parent_id',
            'auth_day',
            'auth_start_time',
            'cash_limit',
            'kitchen_limit',
            'tablet_limit',
            'sale_stock',
            'reserve',
            'category_set',
            'store_type',
            'address',
            // v1.0.6
            // 'mac_addr',
            // 'serial_number',
            'is_open_member',
            'is_open_scan',
            'is_open_assistant',
            'is_open_kitchen_kds',
            'is_open_buffet',
            'table_limit',
            'printer_limit',
            'languages',
            'timezone',
            // v1.0.9
            'is_accept_scan_order',
            // v1.1.0
            'is_open_local_print',
            // v2.0.0
            'is_open_tax',
            // v2.4.0
            'coordinates',
            // v2.10.0
            'enable_table_map',
            'enable_data_management',
            // v2.11.0
            'enable_kiosk',
            // Grab外卖控制
            'enable_grab_delivery',
            // LINE MAN外卖控制
            'enable_lineman_delivery',
            // 扫码点餐到店自取
            'is_open_member_instant',
        ],
        'id' => [
            'app_id',
        ],
        'test' => [
            'minutes'
        ],
    ];

    /**
     * 验证id是否存在
     */
    protected function checkIdExist($value, $rule, $data = [])
    {
        if (!App::where('uuid', $value)->find()) {
            return false;
        } else {
            return true;
        }
    }

    /**
     * 验证商家名称是否存在
     */
    protected function checkNameExist($value, $rule, $data = [])
    {
        $id = $data['app_id'] ?? 0;
        $app = App::where('name', $value)
            ->when($id, function ($q) use ($id) {
                $q->where('uuid', '<>', $id);
            })
            ->find();
        if ($app) {
            return false;
        } else {
            return true;
        }
    }

    protected function isCoordinates($value, $rule, $data = [])
    {
        if (empty($value)) {
            return true;
        }
        if (!isset($value[0]) || !isset($value[1])) {
            return false;
        }
        // 判断经纬度，是否是经度取值范围为[-180,180],纬度取值范围为[-90,90]
        $lat = floatval(explode(',', $value)[0]);
        $lng = floatval(explode(',', $value)[1]);
        if ($lat < -90 || $lat > 90 || $lng < -180 || $lng > 180) {
            return false;
        }
        return true;
    }
}

<?php

namespace app\shop\controller\setting;

use help\ImgHelp;
use app\common\model\order\Order;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\shop\BindRecord;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderStatusEnum;
use app\shop\model\settings\Setting as SettingModel;
use app\common\model\settings\Printer as PrinterModel;

/**
 * 各端设置
 * @Apidoc\Group("terminal_setting")
 * @Apidoc\Sort(10)
 */
class Terminal extends Controller
{
    /**
     * @Apidoc\Title("收银机设置(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/cashier")
     * @Apidoc\Param("carousel", type="array", require=true, desc="上传后的轮播内容url（图片 + 视频）")
     * @Apidoc\Param("is_auto_send", type="int", require=true, default=0, desc="收银结账自动送厨房 0-关闭 1-开启")
     * @Apidoc\Param("order_method", type="object", require=true, desc="v1.0.3用餐方式 收银-is_cashier_order（0-关闭 1-开启） 桌台-is_table_order（0-关闭 1-开启）", children={
     *     @Apidoc\Param("is_cashier_order", type="int", require=true, desc="收银用餐"),
     *     @Apidoc\Param("is_table_order", type="int", require=true, desc="桌台用餐"),
     * })
     * @Apidoc\Param("server", type="array", require=true, desc="收银机服务器连接，【ip、port】")
     * @Apidoc\Param("is_remain_color", type="int", require=true, desc="是否开启剩余时长颜色 0-关闭 1-开启")
     * @Apidoc\Param("remain_color", type="array", require=true, desc="时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)")
     * @Apidoc\Param("is_auto_lock_screen", type="int", require=true, default=0, desc="v1.0.3是否开启自动锁屏 0-关闭 1-开启")
     * @Apidoc\Param("auto_lock_screen", type="int", require=true, default=0, desc="自动锁屏")
     * @Apidoc\Param("is_open_cashier_password", type="int", require=true, default=0, desc="是否开启钱箱密码 0-关闭 1-开启（v1.0.6）")
     * @Apidoc\Param("language", type="array", require=true, desc="常用语言")
     * @Apidoc\Param("default_language", type="string", require=true, default="", desc="默认语言")
     * @Apidoc\Returned()
     */
    public function cashier()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::CASHIER);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $order_method = $data['order_method'] ?? [];
        if (!isset($order_method['is_cashier_order']) || !isset($order_method['is_table_order'])) {
            return $this->renderError('用餐方式不能为空');
        }
        // 检查是否至少有一种用餐方式开启
        if (empty($order_method['is_cashier_order']) && empty($order_method['is_table_order'])) {
            return $this->renderError('请至少开启一种用餐方式');
        }
        // 判断用餐方式是否有未完成订单，有则不能关闭
        $orderModel = new Order();
        if ($order_method['is_cashier_order'] == 0) {
            $count = $orderModel->where("status", 0)->where("desk_uuid", 0)->count();
            if ($count > 0) {
                $orderModel->delStayOrder();
            }
        }
        if ($order_method['is_table_order'] == 0) {
            $count = $orderModel->where("status", 0)->where("desk_uuid", '>', 0)->count();
            if ($count > 0) {
                $orderModel->delStayTableOrder();
            }
        }
        //
        $is_auto_lock_screen = $data['is_auto_lock_screen'] ?? 0;
        if ($is_auto_lock_screen && empty($data['auto_lock_screen'])) {
            return $this->renderError('自动锁屏不能为空');
        }
        $is_remain_color = $data['is_remain_color'] ?? 0;
        if ($is_remain_color && empty($data['remain_color'])) {
            return $this->renderError('时长颜色不能为空');
        }
        if (empty($data['language'])) {
            return $this->renderError('常用语言不能为空');
        }
        if (empty($data['default_language'])) {
            return $this->renderError('默认语言不能为空');
        }

        // 对轮播图片列表进行处理
        $carousel = [];
        $carousels = $data['carousel'] ?? [];
        // 轮播内容数量限制：最多15个
        if (count($carousels) > 15) {
            return $this->renderError('轮播内容最多15个');
        }
        foreach ($carousels as $item) {
            $carousel[] = [
                'file_path' => ImgHelp::removeImageDomain($item['file_path']) ?? '',
                'real_name' => $item['real_name'] ?? '',
                'sort' => $item['sort'] ?? 0,
                'type' => $item['type'] ?? '', // image video
            ];
        }
        $carousels = $carousel;

        // 接收新字段参数
        $no_order_carousel_interval = $data['no_order_carousel_interval'] ?? 10;
        $order_display_mode = $data['order_display_mode'] ?? 'carousel';
        $order_carousel_interval = $data['order_carousel_interval'] ?? 10;

        // 参数验证
        if ($no_order_carousel_interval < 10 || $no_order_carousel_interval > 120) {
            return $this->renderError('未点餐时轮播间隔必须在10-120秒之间');
        }
        if (!in_array($order_display_mode, ['carousel', 'order', 'order_carousel'])) {
            return $this->renderError('点餐时展示模式无效');
        }
        if ($order_carousel_interval < 10 || $order_carousel_interval > 120) {
            return $this->renderError('点餐时轮播间隔必须在10-120秒之间');
        }

        $arr = [
            'carousel' => $carousels, // 轮播内容url
            'order_method' => $order_method, // 用餐方式 收银-is_cashier_order（0-关闭 1-开启） 桌台-is_table_order（0-关闭 1-开启）
            'is_remain_color' => $is_remain_color, // 是否开启剩余时长颜色 0-关闭 1-开启
            'is_auto_send' => $data['is_auto_send'] ?? 0, // 收银结账自动送厨房
            'is_auto_lock_screen' => $is_auto_lock_screen, // 自动锁屏 0-关闭 1-开启
            'auto_lock_screen' => $data['auto_lock_screen'] ?? 300, // 自动锁屏 5分钟
            'is_open_cashier_password' => $data['is_open_cashier_password'] ?? 0, // 是否开启钱箱密码 0-关闭 1-开启
            'language' => $data['language'] ?? [], // 常用语言
            'default_language' => $data['default_language'] ?? 'en', // 默认语言
            'main_cashier_uuid' => $data['main_cashier_uuid'] ?? '', // 主收银机uuid
            'no_order_carousel_interval' => $no_order_carousel_interval, // 未点餐时轮播间隔(秒)
            'order_display_mode' => $order_display_mode, // 点餐时展示模式 carousel/order/order_carousel
            'order_carousel_interval' => $order_carousel_interval, // 点餐时轮播间隔(秒)
        ];
        // 当开启剩余时长颜色时，才会更新时长颜色
        if ($is_remain_color) {
            $arr['remain_color'] = $data['remain_color'] ?? []; // 时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
        }
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::CASHIER, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("收银端-设置高级密码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/editCashierAdvancedPassword")
     * @Apidoc\Param("new_advanced_password", type="string", require=true, default="", desc="新密码")
     * @Apidoc\Param("confirm_advanced_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Returned()
     */
    public function editCashierAdvancedPassword()
    {
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $setting = SettingModel::getItem(SettingEnum::CASHIER);
        if (empty($data['new_advanced_password']) || empty($data['confirm_advanced_password'])) {
            return $this->renderError('请输入新密码');
        }
        if ($data['new_advanced_password'] != $data['confirm_advanced_password']) {
            return $this->renderError('两次密码不一致');
        }
        $setting['advanced_password'] = $data['new_advanced_password'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::CASHIER, $setting, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("收银机-设置钱箱密码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/editCashierPassword")
     * @Apidoc\Param("new_cashier_password", type="string", require=true, default="", desc="新密码")
     * @Apidoc\Param("confirm_cashier_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Returned()
     */
    public function editCashierPassword()
    {
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $setting = SettingModel::getItem(SettingEnum::CASHIER);
        if (empty($data['new_cashier_password']) || empty($data['confirm_cashier_password'])) {
            return $this->renderError('请输入新密码');
        }
        if ($data['new_cashier_password'] != $data['confirm_cashier_password']) {
            return $this->renderError('两次密码不一致');
        }
        $setting['cashier_password'] = $data['new_cashier_password'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::CASHIER, $setting, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("收银机-设置锁屏密码（v1.0.7）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/editCashierLockPassword")
     * @Apidoc\Param("new_lock_password", type="string", require=true, default="", desc="新密码")
     * @Apidoc\Param("confirm_lock_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Returned()
     */
    public function editCashierLockPassword()
    {
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $setting = SettingModel::getItem(SettingEnum::CASHIER);
        if (empty($data['new_lock_password']) || empty($data['confirm_lock_password'])) {
            return $this->renderError('请输入新密码');
        }
        if ($data['new_lock_password'] != $data['confirm_lock_password']) {
            return $this->renderError('两次密码不一致');
        }
        $setting['lock_password'] = $data['new_lock_password'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::CASHIER, $setting, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("平板端设置(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/tablet")
     * @Apidoc\Param("carousel", type="array", require=true, desc="上传后的轮播内容url（图片 + 视频）")
     * @Apidoc\Param("is_call_service", type="int", require=true, default=0, desc="呼叫服务员 0-关闭 1-开启")
     * @Apidoc\Param("is_customer_order", type="int", require=true, default=0, desc="顾客可开桌 0-关闭 1-开启")
     * @Apidoc\Param("is_voice_remind", type="int", require=true, default=0, desc="声音提醒 0-关闭 1-开启（v1.0.5）")
     * @Apidoc\Param("is_show_sold_out", type="int", require=true, default=0, desc="显示售罄商品 0-关闭 1-开启")
     * @Apidoc\Param("is_buffet_order_limit", type="string", require=true, desc="v1.0.4是否开启自助餐下单限制 0-关闭 1-开启")
     * @Apidoc\Param("buffet_order_limit", type="object", require=true, desc="v1.0.4自助餐下单限制", children={
     *     @Apidoc\Param("is_limit_time", type="int", require=true, desc="是否开启时间限制 0-关闭 1-开启"),
     *     @Apidoc\Param("limit_time", type="int", require=true, desc="时间限制（分钟）"),
     *     @Apidoc\Param("is_limit_num", type="int", require=true, desc="是否开启数量限制 0-关闭 1-开启"),
     *     @Apidoc\Param("limit_num", type="int", require=true, desc="数量限制"),
     * })
     * @Apidoc\Param("is_order_limit", type="string", require=true, desc="v1.0.4是否开启非自助餐下单限制 0-关闭 1-开启")
     * @Apidoc\Param("order_limit", type="object", require=true, desc="v1.0.4非自助餐下单限制", children={
     *     @Apidoc\Param("is_limit_time", type="int", require=true, desc="是否开启时间限制 0-关闭 1-开启"),
     *     @Apidoc\Param("limit_time", type="int", require=true, desc="时间限制（分钟）"),
     *     @Apidoc\Param("is_limit_num", type="int", require=true, desc="是否开启数量限制 0-关闭 1-开启"),
     *     @Apidoc\Param("limit_num", type="int", require=true, desc="数量限制"),
     * })
     * @Apidoc\Param("server", type="array", require=true, desc="平板服务器连接，【ip、port】")
     * @Apidoc\Param("language", type="array", require=true, desc="常用语言")
     * @Apidoc\Param("default_language", type="string", require=true, default="", desc="默认语言")
     * @Apidoc\Returned()
     */
    public function tablet()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::TABLET);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        //
        if (empty($data['carousel'])) {
            return $this->renderError('轮播内容不能为空');
        }
        if (empty($data['language'])) {
            return $this->renderError('常用语言不能为空');
        }
        // 自助餐下单限制
        $is_buffet_order_limit = $data['is_buffet_order_limit'] ?? 0;
        if ($is_buffet_order_limit) {
            // 开启后出现时间限制跟数量限制（多选，必选一项）
            $limit_time = $data['buffet_order_limit']['limit_time'] ?? null;
            $limit_num = $data['buffet_order_limit']['limit_num'] ?? null;
            if (!isset($limit_time) && !isset($limit_num)) {
                return $this->renderError('自助餐下单限制不能为空');
            }
            if (isset($limit_time)) {
                $limit_time = floatval($limit_time);
                if ($limit_time < 1 || $limit_time > 999) {
                    return $this->renderError('自助餐时间限制必须在1到999的范围内');
                }
            }
            if (isset($limit_num)) {
                $limit_num = floatval($limit_num);
                if ($limit_num < 1 || $limit_num > 999) {
                    return $this->renderError('自助餐数量限制必须在1到999的范围内');
                }
            }
        }
        // 非自助餐下单限制
        $is_order_limit = $data['is_order_limit'] ?? 0;
        if ($is_order_limit) {
            // 开启后出现时间限制跟数量限制（多选，必选一项）
            $limit_time = $data['order_limit']['limit_time'] ?? null;
            $limit_num = $data['order_limit']['limit_num'] ?? null;
            if (!isset($limit_time) && !isset($limit_num)) {
                return $this->renderError('非自助餐下单限制不能为空');
            }
            if (isset($limit_time)) {
                $limit_time = floatval($limit_time);
                if ($limit_time < 1 || $limit_time > 999) {
                    return $this->renderError('非自助餐时间限制必须在1到999的范围内');
                }
            }
            if (isset($limit_num)) {
                $limit_num = floatval($limit_num);
                if ($limit_num < 1 || $limit_num > 999) {
                    return $this->renderError('非自助餐数量限制必须在1到999的范围内');
                }
            }
        }
        if (empty($data['default_language'])) {
            return $this->renderError('默认语言不能为空');
        }

        // 对轮播图片列表进行处理
        $carousel = [];
        foreach ($data['carousel'] as $item) {
            $carousel[] = [
                'file_path' => ImgHelp::removeImageDomain($item['file_path']) ?? '',
                'real_name' => $item['real_name'] ?? '',
                'sort' => $item['sort'] ?? 0,
                'type' => $item['type'] ?? '', // image video
            ];
        }
        $data['carousel'] = $carousel;

        $arr = [
            'carousel' => $data['carousel'] ?? [], // 轮播内容url
            'is_call_service' => $data['is_call_service'] ?? 0, // 是否开启呼叫服务员
            'is_customer_order' => $data['is_customer_order'] ?? 0, // 是否开启顾客可开桌
            'is_voice_remind' => $data['is_voice_remind'] ?? 0, // 是否开启声音提醒
            'is_show_sold_out' => $data['is_show_sold_out'] ?? 0, // 是否显示售罄商品
            'is_buffet_order_limit' => $is_buffet_order_limit, // 是否开启自助餐下单限制 0-关闭 1-开启
            'buffet_order_limit' => $is_buffet_order_limit ? $data['buffet_order_limit'] : json_decode("{}"), // 自助餐下单限制
            'is_order_limit' => $is_order_limit, // 是否开启非自助餐下单限制 0-关闭 1-开启
            'order_limit' =>  $is_order_limit ? $data['order_limit'] : json_decode("{}"), // 非自助餐下单限制
            'language' => $data['language'] ?? [], // 常用语言
            'default_language' => $data['default_language'] ?? 'en', // 默认语言
        ];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::TABLET, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("平板端-设置高级密码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/editAdvancedPassword")
     * @Apidoc\Param("new_advanced_password", type="string", require=true, default="", desc="新密码")
     * @Apidoc\Param("confirm_advanced_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Returned()
     */
    public function editAdvancedPassword()
    {
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $setting = SettingModel::getItem(SettingEnum::TABLET);
        if (empty($data['new_advanced_password']) || empty($data['confirm_advanced_password'])) {
            return $this->renderError('请输入新密码');
        }
        if ($data['new_advanced_password'] != $data['confirm_advanced_password']) {
            return $this->renderError('两次密码不一致');
        }
        $setting['advanced_password'] = $data['new_advanced_password'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::TABLET, $setting, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("扫码H5设置(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/h5")
     * @Apidoc\Param("is_call_service", type="int", require=true, default=0, desc="呼叫服务员 0-关闭 1-开启")
     * @Apidoc\Param("is_customer_order", type="int", require=true, default=0, desc="顾客可开桌 0-关闭 1-开启")
     * @Apidoc\Param("is_voice_remind", type="int", require=true, default=0, desc="声音提醒 0-关闭 1-开启（v1.0.5）")
     * @Apidoc\Param("is_show_sold_out", type="int", require=true, default=0, desc="显示售罄商品 0-关闭 1-开启")
     * @Apidoc\Param("is_buffet_order_limit", type="string", require=true, desc="v1.0.4是否开启自助餐下单限制 0-关闭 1-开启")
     * @Apidoc\Param("buffet_order_limit", type="object", require=true, desc="v1.0.4自助餐下单限制", children={
     *     @Apidoc\Param("is_limit_time", type="int", require=true, desc="是否开启时间限制 0-关闭 1-开启"),
     *     @Apidoc\Param("limit_time", type="int", require=true, desc="时间限制（分钟）"),
     *     @Apidoc\Param("is_limit_num", type="int", require=true, desc="是否开启数量限制 0-关闭 1-开启"),
     *     @Apidoc\Param("limit_num", type="int", require=true, desc="数量限制"),
     * })
     * @Apidoc\Param("is_order_limit", type="string", require=true, desc="v1.0.4是否开启非自助餐下单限制 0-关闭 1-开启")
     * @Apidoc\Param("order_limit", type="object", require=true, desc="v1.0.4非自助餐下单限制", children={
     *     @Apidoc\Param("is_limit_time", type="int", require=true, desc="是否开启时间限制 0-关闭 1-开启"),
     *     @Apidoc\Param("limit_time", type="int", require=true, desc="时间限制（分钟）"),
     *     @Apidoc\Param("is_limit_num", type="int", require=true, desc="是否开启数量限制 0-关闭 1-开启"),
     *     @Apidoc\Param("limit_num", type="int", require=true, desc="数量限制"),
     * })
     * @Apidoc\Param("language", type="array", require=true, desc="常用语言")
     * @Apidoc\Param("default_language", type="string", require=true, default="", desc="默认语言")
     * @Apidoc\Returned()
     */
    public function h5()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::H5);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        //
        if (empty($data['language'])) {
            return $this->renderError('常用语言不能为空');
        }
        // 自助餐下单限制
        $is_buffet_order_limit = $data['is_buffet_order_limit'] ?? 0;
        if ($is_buffet_order_limit) {
            // 开启后出现时间限制跟数量限制（多选，必选一项）
            $limit_time = $data['buffet_order_limit']['limit_time'] ?? null;
            $limit_num = $data['buffet_order_limit']['limit_num'] ?? null;
            if (!isset($limit_time) && !isset($limit_num)) {
                return $this->renderError('自助餐下单限制不能为空');
            }
            if (isset($limit_time)) {
                $limit_time = floatval($limit_time);
                if ($limit_time < 1 || $limit_time > 999) {
                    return $this->renderError('自助餐时间限制必须在1到999的范围内');
                }
            }
            if (isset($limit_num)) {
                $limit_num = floatval($limit_num);
                if ($limit_num < 1 || $limit_num > 999) {
                    return $this->renderError('自助餐数量限制必须在1到999的范围内');
                }
            }
        }
        // 非自助餐下单限制
        $is_order_limit = $data['is_order_limit'] ?? 0;
        if ($is_order_limit) {
            // 开启后出现时间限制跟数量限制（多选，必选一项）
            $limit_time = $data['order_limit']['limit_time'] ?? null;
            $limit_num = $data['order_limit']['limit_num'] ?? null;
            if (!isset($limit_time) && !isset($limit_num)) {
                return $this->renderError('非自助餐下单限制不能为空');
            }
            if (isset($limit_time)) {
                $limit_time = floatval($limit_time);
                if ($limit_time < 1 || $limit_time > 999) {
                    return $this->renderError('非自助餐时间限制必须在1到999的范围内');
                }
            }
            if (isset($limit_num)) {
                $limit_num = floatval($limit_num);
                if ($limit_num < 1 || $limit_num > 999) {
                    return $this->renderError('非自助餐数量限制必须在1到999的范围内');
                }
            }
        }
        if (empty($data['default_language'])) {
            return $this->renderError('默认语言不能为空');
        }

        $arr = [
            'is_call_service' => $data['is_call_service'] ?? 0, // 是否开启呼叫服务员
            'is_customer_order' => $data['is_customer_order'] ?? 0, // 是否开启顾客可开桌
            'is_voice_remind' => $data['is_voice_remind'] ?? 0, // 是否开启声音提醒
            'is_show_sold_out' => $data['is_show_sold_out'] ?? 0, // 是否显示售罄商品
            'is_buffet_order_limit' => $is_buffet_order_limit, // 是否开启自助餐下单限制 0-关闭 1-开启
            'buffet_order_limit' => $is_buffet_order_limit ? $data['buffet_order_limit'] : json_decode("{}"), // 自助餐下单限制
            'is_order_limit' => $is_order_limit, // 是否开启非自助餐下单限制 0-关闭 1-开启
            'order_limit' =>  $is_order_limit ? $data['order_limit'] : json_decode("{}"), // 非自助餐下单限制
            'language' => $data['language'] ?? [], // 常用语言
            'default_language' => $data['default_language'] ?? 'en', // 默认语言
        ];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::H5, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("厨显设置(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/kitchen")
     * @Apidoc\Param("is_open", type="int", require=true, default=0, desc="是否开启厨显功能 0关闭 1开启（v1.0.4）")
     * @Apidoc\Param("is_come_dish", type="int", require=true, default=0, desc="来菜提醒 0-关闭 1-开启（v1.0.5）")
     * @Apidoc\Param("is_smart_kitchen", type="int", require=true, default=0, desc="智能后厨 0-关闭 1-开启（v2.6.0）")
     * @Apidoc\Param("is_call_service", type="int", require=true, default=0, desc="顾客呼叫提醒 0-关闭 1-开启（v1.0.5）")
     * @Apidoc\Param("server", type="array", require=true, desc="厨显服务器连接")
     * @Apidoc\Param("is_wait_color", type="int", require=true, default=0, desc="是否开启等待颜色 0-关闭 1-开启")
     * @Apidoc\Param("wait_color", type="array", require=true, desc="等待颜色")
     * @Apidoc\Param("wait_time_color_ranges", type="array", require=true, desc="等待时长颜色区间配置（新格式）", children={
     *     @Apidoc\Param("minute", type="string", require=true, desc="时间阈值（分钟，字符串类型）"),
     *     @Apidoc\Param("color", type="string", require=true, desc="颜色值（RGB格式，如 #100A05）"),
     * })
     * @Apidoc\Param("language", type="array", require=true, desc="常用语言")
     * @Apidoc\Param("default_language", type="string", require=true, default="", desc="默认语言")
     * @Apidoc\Param("related_printers", type="array", require=true, desc="关联打印机uuid", children={
     *     @Apidoc\Param("uuid", type="string", require=true, desc="设备uuid"),
     *     @Apidoc\Param("related_printer_uuid", type="string", require=true, desc="关联打印机uuid"),
     * })
     * @Apidoc\Returned()
     */
    public function kitchen()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::KITCHEN);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        //

        if (empty($data['language'])) {
            return $this->renderError('常用语言不能为空');
        }
        if (empty($data['default_language'])) {
            return $this->renderError('默认语言不能为空');
        }

        $arr = [
            'is_open' => $data['is_open'] ?? 0, // 是否开启厨显功能 0关闭 1开启
            'is_come_dish' => $data['is_come_dish'] ?? 0, // 来菜提醒 0关闭 1开启
            'is_call_service' => $data['is_call_service'] ?? 0, // 顾客呼叫提醒 0关闭 1开启
            'is_wait_color' => $data['is_wait_color'] ?? 0, // 是否开启等待颜色
            'wait_color' => $data['wait_color'] ??  [], // 等待颜色（旧格式，保持兼容）
            'wait_time_color_ranges' => $data['wait_time_color_ranges'] ?? [], // 等待时长颜色区间配置（新格式）
            'language' => $data['language'] ?? [], // 常用语言
            'default_language' => $data['default_language'] ?? 'en', // 默认语言
            'bind_list' => $data['bind_list'] ?? [],
            'is_smart_kitchen' => $data['is_smart_kitchen'] ?? 0, // 智能后厨 0-关闭 1-开启
        ];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::KITCHEN, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("厨显端-设置高级密码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/editKitchenAdvancedPassword")
     * @Apidoc\Param("new_advanced_password", type="string", require=true, default="", desc="新密码")
     * @Apidoc\Param("confirm_advanced_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Returned()
     */
    public function editKitchenAdvancedPassword()
    {
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $setting = SettingModel::getItem(SettingEnum::KITCHEN);
        if (empty($data['new_advanced_password']) || empty($data['confirm_advanced_password'])) {
            return $this->renderError('请输入新密码');
        }
        if ($data['new_advanced_password'] != $data['confirm_advanced_password']) {
            return $this->renderError('两次密码不一致');
        }
        $setting['advanced_password'] = $data['new_advanced_password'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::KITCHEN, $setting, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("点餐助手设置(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/assistant")
     * @Apidoc\Param("server", type="array", require=true, desc="收银机服务器连接", children={
     *     @Apidoc\Param("ip", type="string", require=true, desc="ip"),
     *     @Apidoc\Param("port", type="int", require=true, desc="端口"),
     * })
     * @Apidoc\Param("is_remain_color", type="int", require=true, desc="是否开启剩余时长颜色 0-关闭 1-开启")
     * @Apidoc\Param("default_mode", type="int", require=true, desc="默认模式 0-服务员模式 1-顾客模式")
     * @Apidoc\Param("support_function", type="array", require=true, desc="支持功能，多选，add_member-添加会员, people-人数, adjust_buffet-调整自助餐, turn_table-转台, clear_table-清台, merge_table-并台, discount-优惠折扣, price-价格, return_dish-退菜, remark-备注")
     * @Apidoc\Param("remain_color", type="array", require=true, desc="时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)")
     * @Apidoc\Param("is_auto_lock_screen", type="int", require=true, default=0, desc="是否开启自动锁屏 0-关闭 1-开启")
     * @Apidoc\Param("auto_lock_screen", type="int", require=true, default=0, desc="自动锁屏")
     * @Apidoc\Param("language", type="array", require=true, desc="常用语言")
     * @Apidoc\Param("default_language", type="string", require=true, default="", desc="默认语言")
     * @Apidoc\Returned()
     */
    public function assistant()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::ASSISTANT);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $is_auto_lock_screen = $data['is_auto_lock_screen'] ?? '0';
        if ($is_auto_lock_screen && empty($data['auto_lock_screen'])) {
            return $this->renderError('自动锁屏不能为空');
        }
        $is_remain_color = $data['is_remain_color'] ?? '0';
        if ($is_remain_color && empty($data['remain_color'])) {
            return $this->renderError('时长颜色不能为空');
        }
        //
        $is_check_order = $data['is_check_order'] ?? '0';
        if ($is_check_order && empty($data['is_check_order'])) {
            return $this->renderError('下单校验高级密码不能为空');
        }
        if (empty($data['language'])) {
            return $this->renderError('常用语言不能为空');
        }
        if (empty($data['default_language'])) {
            return $this->renderError('默认语言不能为空');
        }

        $arr = [
            'is_remain_color' => $is_remain_color, // 是否开启剩余时长颜色 0-关闭 1-开启
            'is_auto_send' => $data['is_auto_send'] ?? '0', // 收银结账自动送厨房
            'default_mode' => $data['default_mode'] ?? '0', // 默认模式 0-服务员模式 1-顾客模式
            'support_function' => $data['support_function'] ?? [], // 支持功能
            'is_auto_lock_screen' => $is_auto_lock_screen, // 自动锁屏 0-关闭 1-开启
            'auto_lock_screen' => $data['auto_lock_screen'] ?? '300', // 自动锁屏 5分钟
            'language' => $data['language'] ?? [], // 常用语言
            'default_language' => $data['default_language'] ?? 'en', // 默认语言
            'is_check_order' => "$is_check_order", // 下单校验高级密码不能为空 '0'-关闭 '1'-开启
        ];
        // 当开启剩余时长颜色时，才会更新时长颜色
        if ($is_remain_color) {
            $arr['remain_color'] = $data['remain_color'] ?? []; // 时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
        }
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::ASSISTANT, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("点餐助手-设置高级密码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/editAssistantAdvancedPassword")
     * @Apidoc\Param("new_advanced_password", type="string", require=true, default="", desc="新密码")
     * @Apidoc\Param("confirm_advanced_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Returned()
     */
    public function editAssistantAdvancedPassword()
    {
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $setting = SettingModel::getItem(SettingEnum::ASSISTANT);
        if (empty($data['new_advanced_password']) || empty($data['confirm_advanced_password'])) {
            return $this->renderError('请输入新密码');
        }
        if ($data['new_advanced_password'] != $data['confirm_advanced_password']) {
            return $this->renderError('两次密码不一致');
        }
        $setting['advanced_password'] = $data['new_advanced_password'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::ASSISTANT, $setting, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("点餐助手-设置锁屏密码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Terminal/editAssistantLockPassword")
     * @Apidoc\Param("new_lock_password", type="string", require=true, default="", desc="新密码")
     * @Apidoc\Param("confirm_lock_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Returned()
     */
    public function editAssistantLockPassword()
    {
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $setting = SettingModel::getItem(SettingEnum::ASSISTANT);
        if (empty($data['new_lock_password']) || empty($data['confirm_lock_password'])) {
            return $this->renderError('请输入新密码');
        }
        if ($data['new_lock_password'] != $data['confirm_lock_password']) {
            return $this->renderError('两次密码不一致');
        }
        $setting['lock_password'] = $data['new_lock_password'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::ASSISTANT, $setting, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * 获取配置值
     */
    public function fetchData($key)
    {
        if ($key == '') {
            return $this->renderError('缺少参数');
        }
        $shopSupplierId = $this->store['user']['shop_supplier_id'];
        $ret = SettingModel::getSupplierItem($key, $shopSupplierId);
        // 收银机密码
        if ($key == SettingEnum::CASHIER) {
            if (!empty($ret['advanced_password'])) {
                $ret['advanced_password'] = true;
            } else {
                $ret['advanced_password'] = false;
            }
            if (!empty($ret['cashier_password'])) {
                $ret['cashier_password'] = true;
            } else {
                $ret['cashier_password'] = false;
            }
            if (!empty($ret['lock_password'])) {
                $ret['lock_password'] = true;
            } else {
                $ret['lock_password'] = false;
            }
            // 收银/桌台未完成订单数
            $orderModel = new Order();
            $cashierCount = $orderModel->where("status", '=', OrderStatusEnum::NORMAL)->where("desk_uuid", '=', 0)->count();
            $tableCount = $orderModel->where("status", '=', OrderStatusEnum::NORMAL)->where("desk_uuid", '>', 0)->count();
            $ret['cashier_count'] = $cashierCount;
            $ret['table_count'] = $tableCount;
            // 绑定设备列表
            $ret['bind_list'] = (new BindRecord)->getBindList(BindRecord::SOURCE_CASHIER, $shopSupplierId) ?: [];
        }
        // 平板端高级设置密码
        if ($key == SettingEnum::TABLET) {
            if (!empty($ret['advanced_password'])) {
                $ret['advanced_password'] = true;
            } else {
                $ret['advanced_password'] = false;
            }
            // 绑定设备列表
            $ret['bind_list'] = (new BindRecord)->getBindList(BindRecord::SOURCE_TABLET, $shopSupplierId) ?: [];
        }
        // 厨显端高级设置密码
        if ($key == SettingEnum::KITCHEN) {
            if (!empty($ret['advanced_password'])) {
                $ret['advanced_password'] = true;
            } else {
                $ret['advanced_password'] = false;
            }

            // 判断如果wait_color不为空，则将wait_color转换为wait_time_color_ranges
            if (!empty($ret['wait_color']) && empty($ret['wait_time_color_ranges'])) {
                $waitColor = $ret['wait_color'];
                $waitTimeColorRanges = [];
                
                // 第一区间固定黑色
                $waitTimeColorRanges[] = [
                    'minute' => '0',
                    'color' => '#100A05'
                ];
                
                // 颜色映射
                $colorMap = [
                    'red' => '#E50028',
                    'yellow' => '#FFBE00'
                ];
                
                // 旧格式：["red", "yellow"] 或 ["yellow", "red"]
                // 第一个元素对应第二区间（10分钟），第二个元素对应第三区间（20分钟）
                for ($i = 0; $i < 2 && $i < count($waitColor); $i++) {
                    $minute = ($i == 0) ? '10' : '20';
                    $color = '#FFBE00'; // 默认黄色
                    if (isset($colorMap[$waitColor[$i]])) {
                        $color = $colorMap[$waitColor[$i]];
                    }
                    $waitTimeColorRanges[] = [
                        'minute' => $minute,
                        'color' => $color
                    ];
                }
                
                // 如果旧格式数据不足，使用默认值
                if (count($waitTimeColorRanges) < 3) {
                    if (count($waitTimeColorRanges) == 1) {
                        $waitTimeColorRanges[] = [
                            'minute' => '10',
                            'color' => '#FFBE00'
                        ];
                    }
                    if (count($waitTimeColorRanges) == 2) {
                        $waitTimeColorRanges[] = [
                            'minute' => '20',
                            'color' => '#E50028'
                        ];
                    }
                }
                
                $ret['wait_time_color_ranges'] = $waitTimeColorRanges;
            }

            // 绑定设备列表
            $ret['bind_list'] = (new BindRecord)->getBindList(BindRecord::SOURCE_KITCHEN, $shopSupplierId) ?: [];
            // 打印机列表
            $ret['printer_list'] = PrinterModel::getAll(false);
        }
        // 点餐助手密码
        if ($key == SettingEnum::ASSISTANT) {
            if (!empty($ret['advanced_password'])) {
                $ret['advanced_password'] = true;
            } else {
                $ret['advanced_password'] = false;
            }
            if (!empty($ret['lock_password'])) {
                $ret['lock_password'] = true;
            } else {
                $ret['lock_password'] = false;
            }
            // 绑定设备列表
            $ret['bind_list'] = (new BindRecord)->getBindList(BindRecord::SOURCE_ASSISTANT, $shopSupplierId) ?: [];
        }
        //
        $licenses = request()->licenses;
        $ret['licenses'] = [
            'cash_limit' => $licenses['c_l'] ?? 0,
            'tablet_limit' => $licenses['t_l'] ?? 0,
            'kitchen_limit' => $licenses['k_l'] ?? 0,
            'assistant_limit' => $licenses['a_l'] ?? 0,
        ];
        //
        $vars['values'] = $ret;
        return $this->renderSuccess('', compact('vars'));
    }
}

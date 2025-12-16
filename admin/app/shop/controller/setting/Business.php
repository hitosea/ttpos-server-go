<?php

namespace app\shop\controller\setting;

use help\StringHelp;
use help\ValidateHelp;
use Endroid\QrCode\QrCode;
use app\common\model\store\FreeTag;
use app\common\model\store\OrderRemark;
use app\common\model\store\OrderItemRemark;
use app\shop\controller\Controller;
use app\shop\model\store\TableArea;
use hg\apidoc\annotation as Apidoc;
use Endroid\QrCode\Writer\PngWriter;
use app\common\model\order\OrderScheme;
use app\common\model\store\ReturnReason;
use app\common\enum\settings\SettingEnum;
use app\common\service\qrcode\AuthService;
use app\common\model\store\MultiLanguageName;
use app\shop\model\settings\Setting as SettingModel;
use app\common\model\shop\User as UserModel;

/**
 * 门店业务设置（v1.0.7）
 * @Apidoc\Group("supplier")
 * @Apidoc\Sort(4)
 */
class Business extends Controller
{
    /**
     * @Apidoc\Title("门店业务设置(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/index")
     * @Apidoc\Param("zeroing_method", type="string", require=true, desc="优惠折扣自动抹零方式 0-实款实收 1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元")
     * @Apidoc\Param("checkout_zeroing_method", type="string", require=true, desc="结账自动抹零方式 0-实款实收 1-抹分 2-抹角 5-抹元（v1.1.0）")
     * @Apidoc\Param("gift_method", type="string", require=true, desc="赠菜方式 10-计入总销售额、优惠折扣 20-不计入总销售额、优惠折扣")
     * @Apidoc\Param("free_method", type="string", require=true, desc="免单方式 10-计入总销售额、优惠折扣 20-不计入总销售额、优惠折扣")
     * @Apidoc\Param("no_clear_table", type="string", require=true, desc="结账后不清台 0-清台 1-不清台")
     * @Apidoc\Param("is_need_password", type="string", require=true, desc="取消订单/退菜 0-无需密码 1-需要密码（v1.0.8）")
     * @Apidoc\Param("dish_card_style", type="string", require=true, desc="菜品卡片样式 0-无图模式 1-图片模式（v1.0.8）")
     * @Apidoc\Param("discount_method", type="string", require=true, desc="折扣计算方式 10-按百分比 20-直接减免（v1.1.0）")
     * @Apidoc\Param("is_invoice", type="string", require=true, desc="开票信息 0-不需要填写 1-需要填写（v1.0.9补充）")
     * @Apidoc\Returned()
     */
    public function index()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::BUSINESS);
        }
        $model = new SettingModel;
        $data = $this->request->param();

        $setting = SettingModel::getItem(SettingEnum::BUSINESS);
        $old_dish_card_style = $setting['dish_card_style'] ?? 0;
        $update_style_time = $old_dish_card_style != $data['dish_card_style'];
        $opening_hours = $data['opening_hours'] ?? '';
        if ($opening_hours != '') {
            $opening_hours_arr = explode('-', $opening_hours);
            if (count($opening_hours_arr) != 2) {
                return $this->renderError('营业时间格式错误');
            }
            // 验证时间格式：如：08:00-22:00，时间格式为：HH:MM
            $start_time = $opening_hours_arr[0];
            $end_time = $opening_hours_arr[1];
            // 验证开始时间
            if (!preg_match('/^([01][0-9]|2[0-3]):([0-5][0-9])$/', $start_time)) {
                return $this->renderError('开始时间格式不正确');
            }
            
            // 验证结束时间
            if (!preg_match('/^([01][0-9]|2[0-3]):([0-5][0-9])$/', $end_time)) {
                return $this->renderError('结束时间格式不正确');
            }

            $data['opening_hours'] = $opening_hours;
        }

        $deliveryPriceRation = intval($data['delivery_price_ratio'] ?? 0);
        if ($deliveryPriceRation == 0 || !($deliveryPriceRation > 0 && $deliveryPriceRation <= 300)) {
            return $this->renderError('外送商品价格错误');
        }
        if ($setting['delivery_price_ratio'] != $deliveryPriceRation && $this->store['supplier']['delivery_status'] != 1) {
            return $this->renderError('当前没有权限使用此功能');
        }

        // 验证敏感操作设置参数
        $discountNeedPassword = $data['discount_need_password'] ?? '';
        if ($discountNeedPassword !== '' && !in_array($discountNeedPassword, ['0', '1'])) {
            return $this->renderError('折扣权限验证开关参数错误');
        }
        $refundNeedPassword = $data['refund_need_password'] ?? '';
        if ($refundNeedPassword !== '' && !in_array($refundNeedPassword, ['0', '1'])) {
            return $this->renderError('退款权限验证开关参数错误');
        }

        // 验证授权员工ID有效性
        $discountAuthorizedStaffIds = $data['discount_authorized_staff_ids'] ?? [];
        if (!is_array($discountAuthorizedStaffIds)) {
            $discountAuthorizedStaffIds = [];
        }
        $validDiscountStaffIds = [];
        foreach ($discountAuthorizedStaffIds as $staffId) {
            $staffIdInt = intval($staffId);
            if ($staffIdInt > 0 && UserModel::checkUserExist($staffIdInt)) {
                $validDiscountStaffIds[] = $staffIdInt;
            }
        }

        $refundAuthorizedStaffIds = $data['refund_authorized_staff_ids'] ?? [];
        if (!is_array($refundAuthorizedStaffIds)) {
            $refundAuthorizedStaffIds = [];
        }
        $validRefundStaffIds = [];
        foreach ($refundAuthorizedStaffIds as $staffId) {
            $staffIdInt = intval($staffId);
            if ($staffIdInt > 0 && UserModel::checkUserExist($staffIdInt)) {
                $validRefundStaffIds[] = $staffIdInt;
            }
        }

        $arr = [
            'zeroing_method' => $data['zeroing_method'] ?? 0,
            'checkout_zeroing_method' => $data['checkout_zeroing_method'] ?? 0,
            'gift_method' => $data['gift_method'] ?? 10,
            'free_method' => $data['free_method'] ?? 10,
            'no_clear_table' => $data['no_clear_table'] ?? 0,
            'is_need_password' => $data['is_need_password'] ?? 1,
            'dish_card_style' => $data['dish_card_style'] ?? 0,
            'discount_method' => $data['discount_method'] ?? 10,
            'is_invoice' => $data['is_invoice'] ?? 0,
            'opening_hours' => $opening_hours,
            'delivery_price_ratio' => $deliveryPriceRation,
            'start_serial_no' => $data['start_serial_no'] ?? '0001',
        ];

        // 添加敏感操作设置字段
        if ($discountNeedPassword !== '') {
            $arr['discount_need_password'] = $discountNeedPassword;
        }
        if (!empty($validDiscountStaffIds)) {
            $arr['discount_authorized_staff_ids'] = $validDiscountStaffIds;
        } elseif (isset($data['discount_authorized_staff_ids'])) {
            $arr['discount_authorized_staff_ids'] = [];
        }
        if ($refundNeedPassword !== '') {
            $arr['refund_need_password'] = $refundNeedPassword;
        }
        if (!empty($validRefundStaffIds)) {
            $arr['refund_authorized_staff_ids'] = $validRefundStaffIds;
        } elseif (isset($data['refund_authorized_staff_ids'])) {
            $arr['refund_authorized_staff_ids'] = [];
        }
        if ($update_style_time) {
            $arr['dish_card_style_time'] = time() . '';
        }
        // 过滤掉不需要的列表字段
        array_diff_key($arr, array_flip([
            'zeroing_method_list',
            'checkout_zeroing_method_list',
            'gift_method_list',
            'free_method_list'
        ]));
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::BUSINESS, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * 电子菜单二维码生成（v1.1.0）
     * @Apidoc\Title("电子菜单二维码生成（v1.1.0）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/qrcode")
     * @Apidoc\Param("action", type="string", require=true, default="", desc="动作：get-获取/update-更新")
     * @Apidoc\Returned()
     */
    public function qrcode()
    {
        $shop_supplier_id = $this->store['user']['shop_supplier_id'] ?: 0;

        $data = $this->postData();
        $action = $data['action'] ?? '';
        if ($action == '') {
            return $this->renderError('参数错误');
        }
        $setting = SettingModel::getItem(SettingEnum::BUSINESS);
        $qr_code = $setting['qr_code'] ?? 0;

        if ($action == 'update') {
            $qr_code = StringHelp::generatePassword(6, 1);
            $setting['qr_code'] = $qr_code;
            (new SettingModel)->edit(SettingEnum::BUSINESS, $setting, $shop_supplier_id);
        }

        $arr = ['a' => request()->appId, 'q' => $qr_code];
        $auth = new AuthService();
        $token = $auth->generateToken($arr);
        $qrCode = new QrCode(env('EMENU_BASE_URL') . "/home?token={$token}");

        return $this->renderSuccess('', (new PngWriter())->write($qrCode)->getDataUri());
    }

    /**
     * 会员端二维码
     * @Apidoc\Title("会员端二维码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/companyQrcode")
     * @Apidoc\Returned()
     */
    public function companyQrcode()
    {
        if ($this->store['supplier']['is_open_member'] != 1) {
            return $this->renderError('当前没有权限使用此功能');
        }
        $shop_supplier_id = $this->store['user']['shop_supplier_id'] ?: 0;
        $qrCode = new QrCode(env('MEMBER_BASE_URL') . "/launch/{$shop_supplier_id}");
        return $this->renderSuccess('', (new PngWriter())->write($qrCode)->getDataUri());
    }

    /**
     * @Apidoc\Title("免单/赠菜原因(get-获取/post-提交)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/freeTag")
     * @Apidoc\Param("free_tag", type="array", require=true, desc="标签", children={
     *     @Apidoc\Returned("id", type="int", desc="ID"),
     *     @Apidoc\Returned("free_tag", type="string", desc="标签"),
     *     @Apidoc\Returned("action", type="string", desc="操作结果 delete-删除 edit-编辑 add-新增"),
     * })
     * @Apidoc\Returned()
     */
    public function freeTag()
    {
        $shop_supplier_id = $this->store['user']['shop_supplier_id'] ?: 0;
        $app_id = $this->store['app']['app_id'] ?: 0;
        $model = new FreeTag;

        if ($this->request->isGet()) {
            return $this->renderSuccess('操作成功', $model->getList($app_id, $shop_supplier_id));
        }

        $data = $this->request->param();
        $defaultLang = getDefaultLanguage();
        if (isset($data['free_tag'])) {
            if (empty($data['free_tag'])) {
                return $this->renderError('请输入免单/赠菜原因');
            }
            foreach ($data['free_tag'] as $item) {
                $reason = $item['free_tag'] ?? '';
                if (ValidateHelp::hasEmptyValue($reason)) {
                    return $this->renderError('原因不能为空');
                }
                $id = $item['id'] ?? 0;
                $languageUuid = $item['multi_language_name_uuid'] ?? 0;
                $languageData = json_decode($reason, true);
                $name = $languageData[$defaultLang] ?? '';

                if (isset($item['action'])) {
                    if ($item['action'] == 'add' || $item['action'] == 'edit') {
                        $existing = $model->where('id', $id)->find();
                        if ($existing) {
                            $id = $existing['id'];
                            $languageUuid = $existing['multi_language_name_uuid'];
                            $item['action'] = 'edit';
                        } else {
                            $item['app_id'] = $app_id;
                            $item['shop_supplier_id'] = $shop_supplier_id;
                        }
                    }
                    //
                    if ($item['action'] == 'add') {
                        unset($item['id']);
                        unset($item['action']);
                        //
                        $languageUuid = (new MultiLanguageName)->saveNames($languageData);
                        $item['uuid'] = createUuid();
                        $item['name'] = $name;
                        $item['multi_language_name_uuid'] = $languageUuid;
                        (new FreeTag)->save($item);
                    } elseif ($item['action'] == 'delete') {
                        if ($id) {
                            (new MultiLanguageName)->where('uuid', $model->where('id', $id)->value('multi_language_name_uuid'))->delete();
                            $model->where('id', $id)->delete();
                        }
                    } elseif ($item['action'] == 'edit') {
                        if ($id) {
                            unset($item['action']);
                            $item['name'] = $name;
                            (new MultiLanguageName)->saveNames($languageData, $languageUuid);
                            $model->update($item, ['id' => $id]);
                        }
                    }
                }
            }
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("退菜原因(get-获取/post-提交) v1.0.9")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/returnReason")
     * @Apidoc\Param("reason", type="array", require=true, desc="标签", children={
     *     @Apidoc\Returned("id", type="int", desc="ID"),
     *     @Apidoc\Returned("reason", type="string", desc="标签"),
     *     @Apidoc\Returned("action", type="string", desc="操作结果 delete-删除 edit-编辑 add-新增"),
     * })
     * @Apidoc\Returned()
     */
    public function returnReason()
    {
        $shop_supplier_id = $this->store['user']['shop_supplier_id'] ?: 0;
        $app_id = $this->store['app']['app_id'] ?: 0;
        $model = new ReturnReason;

        if ($this->request->isGet()) {
            return $this->renderSuccess('操作成功', $model->getList($app_id, $shop_supplier_id));
        }

        $data = $this->request->param();
        $defaultLang = getDefaultLanguage();
        if (isset($data['reason'])) {
            if (empty($data['reason'])) {
                return $this->renderError('请输入退菜原因');
            }
            foreach ($data['reason'] as $item) {
                $reason = $item['reason'] ?? '';
                if (ValidateHelp::hasEmptyValue($reason)) {
                    return $this->renderError('原因不能为空');
                }
                $id = $item['id'] ?? 0;
                $languageUuid = $item['multi_language_name_uuid'] ?? 0;
                $languageData = json_decode($reason, true);
                $name = $languageData[$defaultLang] ?? '';

                if (isset($item['action'])) {
                    if ($item['action'] == 'add' || $item['action'] == 'edit') {
                        $existing = $model->where('id', $id)->find();
                        if ($existing) {
                            $id = $existing['id'];
                            $languageUuid = $existing['multi_language_name_uuid'];
                            $item['action'] = 'edit';
                        } else {
                            $item['app_id'] = $app_id;
                            $item['shop_supplier_id'] = $shop_supplier_id;
                        }
                    }
                    //
                    if ($item['action'] == 'add') {
                        unset($item['id']);
                        unset($item['action']);
                        //
                        $languageUuid = (new MultiLanguageName)->saveNames($languageData);
                        $item['uuid'] = createUuid();
                        $item['name'] = $name;
                        $item['multi_language_name_uuid'] = $languageUuid;
                        (new ReturnReason)->save($item);
                    } elseif ($item['action'] == 'delete') {
                        if ($id) {
                            (new MultiLanguageName)->where('uuid', $model->where('id', $id)->value('multi_language_name_uuid'))->delete();
                            $model->where('id', $id)->delete();
                        }
                    } elseif ($item['action'] == 'edit') {
                        if ($id) {
                            unset($item['action']);
                            $item['name'] = $name;
                            (new MultiLanguageName)->saveNames($languageData, $languageUuid);
                            $model->update($item, ['id' => $id]);
                        }
                    }
                }
            }
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("整单备注(get-获取/post-提交) v2.8.0")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/orderRemark")
     * @Apidoc\Param("remark", type="array", require=true, desc="整单备注", children={
     *     @Apidoc\Returned("id", type="int", desc="ID"),
     *     @Apidoc\Returned("remark", type="string", desc="整单备注名称"),
     *     @Apidoc\Returned("action", type="string", desc="操作结果 delete-删除 edit-编辑 add-新增"),
     * })
     * @Apidoc\Returned()
     */
    public function orderRemark()
    {
        $shop_supplier_id = $this->store['user']['shop_supplier_id'] ?: 0;
        $app_id = $this->store['app']['app_id'] ?: 0;
        $model = new OrderRemark;

        if ($this->request->isGet()) {
            return $this->renderSuccess('操作成功', $model->getList($app_id, $shop_supplier_id));
        }

        $data = $this->request->param();
        $defaultLang = getDefaultLanguage();
        if (isset($data['remark'])) {
            if (empty($data['remark'])) {
                return $this->renderError('请输入整单备注');
            }
            $count = 0;
            foreach ($data['remark'] as $item) {
                if (isset($item['action']) && $item['action'] != 'delete') {
                    $count++;
                }
            }
            if ($count > 100) {
                return $this->renderError('整单备注数量不能超过100个');
            }
            foreach ($data['remark'] as $item) {
                $remark = $item['remark'] ?? '';
                if (ValidateHelp::hasEmptyValue($remark)) {
                    return $this->renderError('整单备注不能为空');
                }
                $id = $item['id'] ?? 0;
                $languageUuid = $item['multi_language_name_uuid'] ?? 0;
                $languageData = json_decode($remark, true);
                $name = $languageData[$defaultLang] ?? '';

                if (isset($item['action'])) {
                    if ($item['action'] == 'add' || $item['action'] == 'edit') {
                        $existing = $model->where('id', $id)->find();
                        if ($existing) {
                            $id = $existing['id'];
                            $languageUuid = $existing['multi_language_name_uuid'];
                            $item['action'] = 'edit';
                        } else {
                            $item['app_id'] = $app_id;
                            $item['shop_supplier_id'] = $shop_supplier_id;
                        }
                    }
                    //
                    if ($item['action'] == 'add') {
                        unset($item['id']);
                        unset($item['action']);
                        //
                        $languageUuid = (new MultiLanguageName)->saveNames($languageData);
                        $item['uuid'] = createUuid();
                        $item['name'] = $name;
                        $item['multi_language_name_uuid'] = $languageUuid;
                        (new OrderRemark())->save($item);
                    } elseif ($item['action'] == 'delete') {
                        if ($id) {
                            (new MultiLanguageName)->where('uuid', $model->where('id', $id)->value('multi_language_name_uuid'))->delete();
                            OrderRemark::destroy($id);
                        }
                    } elseif ($item['action'] == 'edit') {
                        if ($id) {
                            unset($item['action']);
                            $item['name'] = $name;
                            (new MultiLanguageName)->saveNames($languageData, $languageUuid);
                            $model->update($item, ['id' => $id]);
                        }
                    }
                }
            }
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("单品备注管理")
     * @Apidoc\Method ("GET|POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/orderItemRemark")
     * @Apidoc\Param("order_item_remark", type="array", require=false, desc="单品备注", children={
     *     @Apidoc\Returned("id", type="int", desc="ID"),
     *     @Apidoc\Returned("remark", type="string", desc="单品备注名称"),
     *     @Apidoc\Returned("action", type="string", desc="操作结果 delete-删除 edit-编辑 add-新增"),
     * })
     * @Apidoc\Returned()
     */
    public function orderItemRemark()
    {
        $model = new OrderItemRemark;

        if ($this->request->isGet()) {
            return $this->renderSuccess('操作成功', $model->getList());
        }

        $data = $this->request->param();
        $defaultLang = getDefaultLanguage();
        if (isset($data['order_item_remark'])) {
            if (empty($data['order_item_remark'])) {
                return $this->renderError('请输入单品备注');
            }
            $count = 0;
            foreach ($data['order_item_remark'] as $item) {
                if (isset($item['action']) && $item['action'] != 'delete') {
                    $count++;
                }
            }
            if ($count > 100) {
                return $this->renderError('单品备注数量不能超过100个');
            }
            foreach ($data['order_item_remark'] as $item) {
                $remark = $item['remark'] ?? '';
                if (ValidateHelp::hasEmptyValue($remark)) {
                    return $this->renderError('单品备注不能为空');
                }
                $id = $item['id'] ?? 0;
                $languageUuid = $item['multi_language_name_uuid'] ?? 0;
                $languageData = json_decode($remark, true);
                $name = $languageData[$defaultLang] ?? '';

                if (isset($item['action'])) {
                    if ($item['action'] == 'add' || $item['action'] == 'edit') {
                        $existing = $model->where('id', $id)->find();
                        if ($existing) {
                            $id = $existing['id'];
                            $languageUuid = $existing['multi_language_name_uuid'];
                            $item['action'] = 'edit';
                        }
                    }
                    //
                    if ($item['action'] == 'add') {
                        unset($item['id']);
                        unset($item['action']);
                        //
                        $languageUuid = (new MultiLanguageName)->saveNames($languageData);
                        $item['uuid'] = createUuid();
                        $item['name'] = $name;
                        $item['multi_language_name_uuid'] = $languageUuid;
                        (new OrderItemRemark())->save($item);
                    } elseif ($item['action'] == 'delete') {
                        if ($id) {
                            (new MultiLanguageName)->where('uuid', $model->where('id', $id)->value('multi_language_name_uuid'))->delete();
                            OrderItemRemark::destroy($id);
                        }
                    } elseif ($item['action'] == 'edit') {
                        if ($id) {
                            unset($item['action']);
                            $item['name'] = $name;
                            (new MultiLanguageName)->saveNames($languageData, $languageUuid);
                            $model->update($item, ['id' => $id]);
                        }
                    }
                }
            }
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("订单方案列表（v1.0.9）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/orderSchemeList")
     * @Apidoc\Param("name", type="string", require=true, desc="方案名称")
     * @Apidoc\Param("status", type="int", require=true, desc="状态，1-开启 0-关闭")
     * @Apidoc\Returned()
     */
    public function orderSchemeList()
    {
        $shop_supplier_id = $this->store['user']['shop_supplier_id'] ?: 0;
        $app_id = $this->store['app']['app_id'] ?: 0;
        //
        $model = new OrderScheme;
        $data = $this->request->param();
        $list = $model->getList($data, $app_id, $shop_supplier_id);
        // 桌台区域列表
        $params['list_rows'] = 1000;
        $table_area_list = (new TableArea)->getList($params, $shop_supplier_id);
        return $this->renderSuccess('', compact('list', 'table_area_list'));
    }

    /**
     * @Apidoc\Title("订单方案新增（v1.0.9）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/orderSchemeAdd")
     * @Apidoc\Param("name", type="string", require=true, desc="方案名称")
     * @Apidoc\Param("use_channel", type="array", require=true, desc="使用渠道 10-点餐方式 20-桌台方式")
     * @Apidoc\Param("table_area_ids", type="array", require=true, desc="桌台区域ids")
     * @Apidoc\Param("must_type", type="int", require=true, desc="必点类型 1-每人必点1份 2-每笔订单必点1份")
     * @Apidoc\Param("must_rule", type="int", require=true, desc="必点规则 1-固定商品 2-可选商品")
     * @Apidoc\Param("product_ids", type="array", require=true, desc="必点商品ids")
     * @Apidoc\Param("status", type="int", require=true, desc="状态，1-开启 2-关闭")
     * @Apidoc\Param("auto_cart", type="int", require=true, desc="自动加入购物车 1-是 0-否")
     * @Apidoc\Param("auto_change", type="int", require=true, desc="顾客可修改必点数量 1-是 0-否")
     * @Apidoc\Param("auto_check", type="int", require=true, desc="下单时检查必点商品 1-是 0-否")
     * @Apidoc\Param("auto_checkout", type="int", require=true, desc="结账时检查必点商品 1-是 0-否")
     * @Apidoc\Returned()
     */
    public function orderSchemeAdd()
    {
        //
        $model = new OrderScheme;
        $data = $this->request->param();
        if ($model->add($data)) {
            return $this->renderSuccess('');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("订单方案编辑（v1.0.9）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/orderSchemeEdit")
     * @Apidoc\Param("id", type="int", require=true, desc="方案id")
     * @Apidoc\Param("name", type="string", require=true, desc="方案名称")
     * @Apidoc\Param("use_channel", type="array", require=true, desc="使用渠道 10-点餐方式 20-桌台方式")
     * @Apidoc\Param("table_area_ids", type="array", require=true, desc="桌台区域ids")
     * @Apidoc\Param("must_type", type="int", require=true, desc="必点类型 1-每人必点1份 2-每笔订单必点1份")
     * @Apidoc\Param("must_rule", type="int", require=true, desc="必点规则 1-固定商品 2-可选商品")
     * @Apidoc\Param("product_ids", type="array", require=true, desc="必点商品ids")
     * @Apidoc\Param("status", type="int", require=true, desc="状态，1-开启 2-关闭")
     * @Apidoc\Param("auto_cart", type="int", require=true, desc="自动加入购物车 1-是 0-否")
     * @Apidoc\Param("auto_change", type="int", require=true, desc="顾客可修改必点数量 1-是 0-否")
     * @Apidoc\Param("auto_check", type="int", require=true, desc="下单时检查必点商品 1-是 0-否")
     * @Apidoc\Param("auto_checkout", type="int", require=true, desc="结账时检查必点商品 1-是 0-否")
     * @Apidoc\Returned()
     */
    public function orderSchemeEdit()
    {
        $data = $this->request->param();
        /** @var OrderScheme $model */
        $model = (new OrderScheme)->detail($data['id']);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        $data = $this->request->param();
        if ($model->edit($data)) {
            return $this->renderSuccess('');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("订单方案删除（v1.0.9）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/orderSchemeDel")
     * @Apidoc\Param("id", type="int", require=true, desc="方案id")
     * @Apidoc\Returned()
     */
    public function orderSchemeDel()
    {
        $data = $this->request->param();
        /** @var OrderScheme $model */
        $model = (new OrderScheme)->detail($data['id']);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($model->del()) {
            return $this->renderSuccess('');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("订单方案状态设置（v1.0.9）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Business/orderSchemeStatus")
     * @Apidoc\Param("id", type="int", require=true, desc="方案id")
     * @Apidoc\Param("status", type="int", require=true, desc="状态，1-开启 2-关闭")
     * @Apidoc\Returned()
     */
    public function orderSchemeStatus()
    {
        $data = $this->request->param();
        if (!in_array($data['status'], [1, 0])) {
            return $this->renderError('状态错误');
        }
        /** @var OrderScheme $model */
        $model = (new OrderScheme)->detail($data['id']);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($model->setStatus($data)) {
            return $this->renderSuccess('');
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
        $shop_supplier_id = $this->store['user']['shop_supplier_id'] ?: 0;
        $app_id = $this->store['app']['app_id'] ?: 0;
        $ret = SettingModel::getSupplierItem($key, $shop_supplier_id, $app_id);
        // 门店营业时间
        if ($key == SettingEnum::BUSINESS && $ret['opening_hours'] == '') {
            $ret['opening_hours'] = '00:00-23:59';
        }
        //
        $free_tag_count = FreeTag::count();
        $order_remark_count = OrderRemark::count();
        $order_item_remark_count = OrderItemRemark::count();
        $return_reason_count = ReturnReason::count();
        $company_link = request()->licenses['is_open_member'] == 1 ? env('MEMBER_BASE_URL') . "/launch/{$shop_supplier_id}" : '';
        $vars['values'] = $ret;
        return $this->renderSuccess('', compact('vars', 'free_tag_count', 'return_reason_count', 'order_remark_count', 'order_item_remark_count', 'company_link'));
    }
}

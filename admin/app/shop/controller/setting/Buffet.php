<?php

namespace app\shop\controller\setting;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\enum\settings\SettingEnum;
use app\shop\model\delay\Delay as DelayModel;
use app\shop\model\settings\Setting as SettingModel;
use app\common\model\buffet\CustomerType as CustomerTypeModel;

/**
 * 自助餐设置
 * @Apidoc\Group("product")
 * @Apidoc\Sort(4)
 */
class Buffet extends Controller
{
    /**
     * @Apidoc\Title("自助餐设置(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Buffet/index")
     * @Apidoc\Param("is_open", type="string", require=true, desc="是否开启自助餐")
     * @Apidoc\Param("tablet_end_time", type="string", require=true, default="5", desc="平板结束时间提醒（分）")
     * @Apidoc\Param("is_remain_continue", type="int", require=true, desc="剩余xx分不可继续点餐开关 0-关闭 1-开启（v1.0.4弃用）")
     * @Apidoc\Param("remain_continue_time", type="string", require=true, desc="剩余xx分不可继续点餐（v1.0.4弃用）")
     * @Apidoc\Param("remain_continue_notice_time", type="string", require=true, desc="剩余xx分提醒不可继续点餐（v1.0.4弃用）")
     * @Apidoc\Param("is_show_non_buffet_product", type="string", require=true, desc="是否显示非自助餐商品 0-关闭 1-开启")
     * @Apidoc\Param("is_buy_continue", type="string", require=true, desc="非自助餐商品到时是否能继续选购")
     * @Apidoc\Param("is_buffet_discount", type="string", require=true, desc="是否开启自助餐折扣 0-否 1-是（v1.0.2弃用）")
     * @Apidoc\Param("add_buffet_discount", type="array", require=true, desc="自助餐设置（v1.0.2弃用）", children={
     *     @Apidoc\Returned("id", type="int", desc="自助餐折扣id"),
     *     @Apidoc\Returned("name", type="string", desc="名称"),
     *     @Apidoc\Returned("discount_type", type="string", desc="折扣类型 1-比例 2-优惠金额"),
     *     @Apidoc\Returned("discount_ratio", type="string", desc="折扣百分比"),
     *     @Apidoc\Returned("discount_price", type="string", desc="优惠金额"),
     *     @Apidoc\Returned("buffet_ids", type="string", desc="自助餐id，多个以英文逗号分隔"),
     *     @Apidoc\Returned("action", type="string", desc="操作结果 delete-删除 edit-编辑 add-新增"),
     * })
     * @Apidoc\Param("customer_type", type="array", require=true, desc="顾客类型", children={
     *     @Apidoc\Param("id", type="int", require=true, desc="顾客类型id"),
     *     @Apidoc\Param("name", type="string", require=true, desc="名称"),
     *     @Apidoc\Param("action", type="string", require=true, desc="操作结果 delete-删除 edit-编辑 add-新增"),
     * })
     * @Apidoc\Param("is_add_clock", type="string", require=true, desc="是否开启加钟 0-否 1-是")
     * @Apidoc\Param("add_clock", type="array", require=true, desc="名称 - 加钟时间（分）- 价格", children={
     *     @Apidoc\Returned("id", type="int", desc="加钟id"),
     *     @Apidoc\Returned("name", type="string", desc="名称"),
     *     @Apidoc\Returned("delay_time", type="string", desc="时间（分）"),
     *     @Apidoc\Returned("price", type="string", desc="价格"),
     *     @Apidoc\Returned("action", type="string", desc="操作结果 delete-删除 edit-编辑 add-新增"),
     * })
     * @Apidoc\Returned()
     */
    public function index()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::BUFFET);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        // 是否开启自助餐 0-否 1-是
        $is_open = $data['is_open'] ?? '0';
        //
        if (empty($data['tablet_end_time'])) {
            return $this->renderError('请输入1-999的整数');
        }
        // is_remain_continue 开启时
        if ($data['is_remain_continue'] == 1) {
            // 限制1-999的整数范围
            if (!is_numeric($data['remain_continue_time']) || $data['remain_continue_time'] < 1 || $data['remain_continue_time'] > 999) {
                return $this->renderError('请输入1-999的整数');
            }
            if (!is_numeric($data['remain_continue_notice_time']) || $data['remain_continue_notice_time'] < 1 || $data['remain_continue_notice_time'] > 999) {
                return $this->renderError('请输入1-999的整数');
            }
        }
        // 顾客类型
        if ((!isset($data['customer_type']) || empty($data['customer_type']))) {
            return $this->renderError('顾客类型不能为空');
        }
        if ($data['customer_type']) {
            $customerCommonModel = new CustomerTypeModel;
            // 有多少个顾客类型
            $oldCount = $customerCommonModel->count();
            // 遍历除去删除的顾客类型 不判断数量
            $deleteCount = 0;
            $addCount = 0;
            $editCount = 0;
            foreach ($data['customer_type'] as $customer) {
                $action = $customer['action'] ?? null;
                if (empty($customer['name'])) {
                    return $this->renderError('顾客类型名称不能为空');
                }
                if (mb_strlen($customer['name']) > 50) {
                    return $this->renderError('顾客类型名称不能超过50字符');
                }
                if ($action == 'delete') {
                    $deleteCount++;
                    continue;
                }
                if ($action == 'add') {
                    $addCount++;
                    continue;
                }
                if ($action == 'edit') {
                    $editCount++;
                    continue;
                }
            }
            // 新的顾客类型必须在1到5之间
            if (($oldCount + $addCount - $deleteCount) > 5 || $editCount > 5) {
                return $this->renderError('顾客类型必须在1到5个之间');
            }
            // 开始事务
            foreach ($data['customer_type'] as $customer) {
                $customerModel = new CustomerTypeModel;
                $action = $customer['action'] ?? null;
                // 如果是新增
                if ($action == 'add') {
                    unset($customer['action']);
                    $customerModel->save($customer);
                    continue;
                }
                // 如果是删除 - 保留数据，只是修改状态
                if ($action == 'delete') {
                    if (isset($customer['id'])) {
                        $res = $customerModel->setDelete($customer['id']);
                        if (!$res) {
                            return $this->renderError('删除顾客类型失败');
                        }
                    }
                    continue;
                }
                // 如果是编辑
                if ($action == 'edit') {
                    if (isset($customer['id'])) {
                        $customerModel->update([
                            'name' => $customer['name']
                        ], ['uuid' => $customer['id']]);
                    }
                    continue;
                }
            }
        }
        // 加钟
        if (!empty($data['is_add_clock']) && isset($data['add_clock'])) {
            if (empty($data['add_clock'])) {
                return $this->renderError('加钟设置不能为空');
            }
            foreach ($data['add_clock'] as $clock) {
                $delayModel = new DelayModel;
                if (empty($clock['name']) || empty($clock['delay_time']) || $clock['price'] < 0) {
                    return $this->renderError('加钟设置参数错误');
                }
                // 如果是新增
                if (isset($clock['action']) && $clock['action'] == 'add') {
                    unset($clock['action']);
                    $clock['status'] = 1;
                    $delayModel->save($clock);
                    continue;
                }
                // 如果是删除 - 保留数据，只是修改状态
                if (isset($clock['action']) && $clock['action'] == 'delete') {
                    if (isset($clock['id'])) {
                        $delayModel->where('id', $clock['id'])->find()?->delete();
                    }
                    continue;
                }
                // 如果是编辑
                if (isset($clock['action']) && $clock['action'] == 'edit') {
                    if (isset($clock['id'])) {
                        unset($clock['action']);
                        $delayModel->update($clock, ['id' => $clock['id']]);
                    }
                    continue;
                }
            }
        }
        $arr = [
            'is_open' => $is_open, // 是否开启自助餐 0-否 1-是
            'tablet_end_time' => $data['tablet_end_time'], // 平板结束时间提醒（分）
            'is_remain_continue' => $data['is_remain_continue'] ?? 0, // 剩余xx分不可继续点餐开关 0-关闭 1-开启
            'remain_continue_time' => $data['remain_continue_time'], // 剩余xx分不可继续点餐
            'remain_continue_notice_time' => $data['remain_continue_notice_time'], // 剩余xx分提醒不可继续点餐
            'is_buy_continue' => $data['is_buy_continue'] ?? 0, // 非自助餐商品到时是否能继续选购
            'is_show_non_buffet_product' => $data['is_show_non_buffet_product'] ?? 1, // 是否显示非自助餐商品 0-关闭 1-开启
            'is_buffet_discount' => $data['is_buffet_discount'] ?? 0, // 是否开启自助餐折扣 0-否 1-是
            'is_add_clock' => $data['is_add_clock'] ?? 0, // 是否开启加钟 0-否 1-是
            'add_clock' => [], // 名称 - 加钟时间（分）- 价格（放表，不放设置里）
        ];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::BUFFET, $arr, $shop_supplier_id)) {
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
        $ret = SettingModel::getSupplierItem($key, $this->store['user']['shop_supplier_id'], $this->store['app']['app_id']);
        // 获取加钟列表
        $delay = DelayModel::getList();
        $ret['add_clock'] = $delay;
        // 获取顾客类型列表
        $customerType = CustomerTypeModel::getList();
        $ret['add_buffet_discount'] = [];
        $ret['customer_type'] = $customerType;
        $vars['values'] = $ret;
        return $this->renderSuccess('', compact('vars'));
    }
}

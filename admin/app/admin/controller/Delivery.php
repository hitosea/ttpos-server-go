<?php

namespace app\admin\controller;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\validate\DeliveryValidate;
use app\common\enum\settings\SettingEnum;
use app\admin\model\settings\Setting as SettingModel;
use app\common\model\supplier\Supplier;
use app\admin\model\app\App as AppModel;

/**
 * 外送管理
 * @Apidoc\Group("base")
 * @Apidoc\Sort(6)
 */
class Delivery extends Controller
{

    /**
     * @Apidoc\Title("外送设置")
     * @Apidoc\Desc("get请求是获取，返回一个对象数组，包含多个渠道；post请求是提交修改，只能提交一个渠道")
     * @Apidoc\Method("GET,POST")
     * @Apidoc\Url("/api/admin/delivery/index")
     * @Apidoc\Param("channel", type="string", require=true, desc="配送渠道名称，例如：SKootar")
     * @Apidoc\Param("basic_fee", type="decimal", require=true, desc="基础服务费")
     * @Apidoc\Param("base_delivery_fee", type="decimal", require=true, desc="基础配送费")
     * @Apidoc\Param("rider_acceptance_timeout", type="int", require=true, desc="骑手接单超时时间（单位：分钟）")
     * @Apidoc\Param("distance_range", type="array", require=true, desc="距离区间配置")
     * @Apidoc\Param("distance_range[].start", type="int", require=true, desc="区间起始公里数")
     * @Apidoc\Param("distance_range[].end", type="int", require=true, desc="区间结束公里数")
     * @Apidoc\Param("distance_range[].price_per_km", type="decimal", require=true, desc="每公里价格")
     * @Apidoc\Param("distance_range[].is_unlimited", type="bool", require=true, desc="是否为无限区间")
     */
    public function index(DeliveryValidate $validate)
    {
        if ($this->request->isGet()) {
            return $this->fetchData();
        }
        $param = $validate->goCheck('edit');
        $deliverySetting = SettingModel::where(['key' => SettingEnum::DELIVERY_CONFIG])->find();

        $channels = [];
        if (!is_null($deliverySetting)) {
            $channels = $deliverySetting->values;
        }
        $exists = false;
        foreach ($channels as $k => $item) {
            if ($item["channel"] == $param["channel"]) {
                $channels[$k] = $param;
                $exists = true;
            }
        }
        if (!$exists) {
            $channels[] = $param;
        }
        if (is_null($deliverySetting)) {
            SettingModel::insert([
                "key" => SettingEnum::DELIVERY_CONFIG,
                "describe" => "外送设置",
                "values" => json_encode($channels)
            ]);
        } else {
            $deliverySetting->save([
                'describe' => SettingEnum::data()[SettingEnum::DELIVERY_CONFIG]['describe'],
                'values' =>  $channels,
            ]);
            // 获取 company_setting 表，如果设置了自动同步，需要更新，以及对应的商家数据库
            $companySettings = Supplier::select();
            foreach ($companySettings as $companySetting) {
                if (!$companySetting || !$companySetting->delivery_config) {
                    continue;
                }
                $deliveryConfig = json_decode($companySetting->delivery_config, true);
                foreach ($deliveryConfig as $k => $channel) {
                    if ($channel["channel"] == $param["channel"] && $channel["config_type"] == "auto_sync") {
                        $deliveryConfig[$k] = $param + ["config_type" => "auto_sync"];
                    }
                }
                $companySetting->save(['delivery_config' => json_encode($deliveryConfig)]);
            }
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * 获取外送设置
     */
    public function fetchData()
    {
        $deliverySetting = SettingModel::where(['key' => SettingEnum::DELIVERY_CONFIG])->find();
        if (!$deliverySetting) {
            $channels = $this->appendDefault([], SettingEnum::DELIVERY_CHANNELS);
        } else {
            $channels = $deliverySetting["values"];
            $existsChannels = [];
            foreach ($channels as $k => $channel) {
                if (!in_array($channel["channel"], SettingEnum::DELIVERY_CHANNELS)) {
                    unset($channels[$k]);
                    continue;
                }
                $existsChannels[] = $channel["channel"];
            }
            $channels = array_values($channels);
            $emptyChannels = array_diff(SettingEnum::DELIVERY_CHANNELS, $existsChannels);

            if (!empty($emptyChannels)) {
                $channels = $this->appendDefault($channels, $emptyChannels);
            }
        }

        return $this->renderSuccess('', $channels);
    }

    /**
     * 补充渠道默认数值
     */
    private function appendDefault($channels, $emptyChannels)
    {
        foreach ($emptyChannels as $channel) {
            $channels[] = [
                'channel' => $channel,
                'basic_fee' => 0,
                'base_delivery_fee' => 0,
                'rider_acceptance_timeout' => 0,
                'distance_range' => [
                    [
                        'start' => 0,
                        'end' => 0,
                        'price_per_km' => 0,
                        'is_unlimited' => false,
                    ]
                ],
            ];
        }
        return $channels;
    }

    /**
     * @Apidoc\Title("外送渠道列表")
     * @Apidoc\Desc("")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/delivery/channels")
     * @Apidoc\Param("configured", type="int", default="0", desc="是否仅获取设置的渠道：0-否；1-是")
     * @Apidoc\Returned(type="array", desc="外送渠道选项")
     * @Apidoc\Returned("name", type="string", desc="渠道名称")
     * @Apidoc\Returned("value", type="string", desc="渠道值")
     */
    public function channels()
    {
        $param = $this->getData();
        $getConfigured = ($param['configured'] ?? 0) == 1;  // 是否获取已配置的

        $res = [];
        if (!$getConfigured) {  // 获取所有外送渠道列表
            foreach (SettingEnum::DELIVERY_CHANNELS as $channel) {
                $res[] = [
                    "name" => $channel,
                    "value" => $channel
                ];
            }
        } else { // 获取已设置的外送渠道列表
            $deliverySetting = SettingModel::where(['key' => SettingEnum::DELIVERY_CONFIG])->find();
            if (!$deliverySetting) {
                return $this->renderSuccess('', []);
            }
            $channels = $deliverySetting["values"];
            foreach ($channels as $channel) {
                if (!in_array($channel["channel"], SettingEnum::DELIVERY_CHANNELS)) {
                    continue;
                }
                $res[] = [
                    "name" => $channel["channel"],
                    "value" => $channel["channel"],
                ];
            }
        }
        return $this->renderSuccess('', $res);
    }

    /**
     * @Apidoc\Title("外送商家列表")
     * @Apidoc\Desc("get请求是获取，post请求是提交修改")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/delivery/companyList")
     * @Apidoc\Param("keyword", type="string", default="", desc="商家名称")
     * @Apidoc\Param("channel", type="string", default="", desc="外送渠道：SKootar") 
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("uuid", type="biginteger", desc="商家ID")
     * @Apidoc\Returned("name", type="string", desc="商家名称")
     * @Apidoc\Returned("delivery_status", type="int", desc="外送状态，0为关闭，1为开启")
     * @Apidoc\Returned("delivery_config", type="array", desc="外送渠道配置列表")
     * @Apidoc\Returned("delivery_config[].channel", type="string", desc="配送渠道名称")
     * @Apidoc\Returned("delivery_config[].basic_fee", type="decimal", desc="基础服务费")
     * @Apidoc\Returned("delivery_config[].base_delivery_fee", type="decimal", desc="基础配送费")
     * @Apidoc\Returned("delivery_config[].rider_acceptance_timeout", type="int", desc="骑手接单超时时间（分钟）")
     * @Apidoc\Returned("delivery_config[].distance_range", type="array", desc="距离区间配置")
     * @Apidoc\Returned("delivery_config[].distance_range[].start", type="int", desc="区间起始公里数")
     * @Apidoc\Returned("delivery_config[].distance_range[].end", type="int", desc="区间结束公里数")
     * @Apidoc\Returned("delivery_config[].distance_range[].price_per_km", type="decimal", desc="每公里价格")
     * @Apidoc\Returned("delivery_config[].distance_range[].is_unlimited", type="bool", desc="是否为无限区间")
     * @Apidoc\Returned("delivery_config[].config_type", type="string", desc="配置类型：auto_sync-自动同步；manual手动设置")
     * @Apidoc\Returned("channel_names", type="array", desc="渠道名称，分号分隔")
     * @Apidoc\Returned("channel_config_types", type="array", desc="配置类型：auto_sync-自动同步；manual手动设置，分号分隔")
     */
    public function companyList()
    {
        $param = array_merge($this->getData(), ["configured" => true]);
        $list = (new AppModel)->getDeliveryList($param)?->toArray();
        if (!empty($list["data"])) {
            foreach ($list["data"] as $k => $item) {
                $channels = json_decode($item['delivery_config'], true);
                $channelNames = [];
                $channelConfigTypes = [];
                foreach ($channels as $kk => $channel) {
                    if (!in_array($channel["channel"], SettingEnum::DELIVERY_CHANNELS)) {
                        continue;
                    }
                    $channelNames[] = $channel["channel"];
                    $channelConfigTypes[] = $channel["config_type"];
                }
                unset($list["data"][$k]["delivery_config"]);
                $list["data"][$k]["delivery_config"] = $channels;
                $list["data"][$k]["channel_names"] = $channelNames;
                $list["data"][$k]["channel_config_types"] = $channelConfigTypes;
            }
        }
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("可选商家列表")
     * @Apidoc\Desc("")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/delivery/companySelect")
     * @Apidoc\Param("keyword", type="string", default="", desc="商家名称")
     * @Apidoc\Param("channel", type="string", default="", desc="外送渠道：SKootar") 
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned(type="array", desc="商家列表")
     * @Apidoc\Returned("uuid", type="biginteger", desc="商家ID")
     * @Apidoc\Returned("name", type="string", desc="商家名称")
     * @Apidoc\Returned("channel_names", type="array", desc="商家已配置的外送渠道名称列表，如果长度不为空，表示已配置")
     */
    public function companySelect()
    {
        $param = array_merge($this->getData(), ["configured" => false]);
        $list = (new AppModel)->getDeliveryList($param)?->toArray();
        if (!empty($list["data"])) {
            foreach ($list["data"] as $k => $item) {
                $channelNames = [];
                if ($item['delivery_config'] != "") {
                    $channels = json_decode($item['delivery_config'], true);
                    foreach ($channels as $kk => $channel) {
                        if (in_array($channel["channel"], SettingEnum::DELIVERY_CHANNELS)) {
                            $channelNames[] = $channel["channel"];
                        }
                    }
                }
                unset($list["data"][$k]["delivery_config"]);
                unset($list["data"][$k]["delivery_status"]);
                $list["data"][$k]["channel_names"] = $channelNames;
            }
        }
        return $this->renderSuccess('', compact('list'));
    }


    /**
     * @Apidoc\Title("启用禁用外送状态")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/delivery/updateStatus")
     * @Apidoc\Param("uuid", type="int", require=true, default="", desc="主键")
     */
    public function updateStatus(DeliveryValidate $validate)
    {
        $param = $validate->goCheck('uuid');
        $model = Supplier::where('uuid', $param['uuid'])->find();
        if (!$model->allowField(['delivery_status'])->save([
            'delivery_status' => $model->delivery_status == 1 ? 0 : 1,
        ])) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("添加\修改商家外送配置")
     * @Apidoc\Desc("")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/delivery/company")
     * @Apidoc\Param("uuid", type="biginteger", require=true, desc="商家唯一标识")
     * @Apidoc\Param("channels", type="array", require=true, desc="外送渠道配置列表")
     * @Apidoc\Param("channels[].config_type", type="string", require=true, desc="配置类型，如 auto_sync-自动同步；manual-手动设置")
     * @Apidoc\Param("channels[].channel", type="string", require=true, desc="配送渠道名称")
     * @Apidoc\Param("channels[].basic_fee", type="decimal", require=true, desc="基础服务费")
     * @Apidoc\Param("channels[].base_delivery_fee", type="decimal", require=true, desc="基础配送费")
     * @Apidoc\Param("channels[].rider_acceptance_timeout", type="int", require=true, desc="骑手接单超时时间（分钟）")
     * @Apidoc\Param("channels[].distance_range", type="array", require=true, desc="距离区间配置")
     * @Apidoc\Param("channels[].distance_range[].start", type="int", require=true, desc="区间起始公里数")
     * @Apidoc\Param("channels[].distance_range[].end", type="int", require=true, desc="区间结束公里数")
     * @Apidoc\Param("channels[].distance_range[].price_per_km", type="decimal", require=true, desc="每公里价格")
     */
    public function company(DeliveryValidate $validate)
    {
        $param = $validate->goCheck('add_company');
        $deliverySetting = SettingModel::where(['key' => SettingEnum::DELIVERY_CONFIG])->find();
        if (!$deliverySetting) {
            return $this->renderSuccess('外送渠道未设置');
        }
        $configureChannels = [];
        foreach ($deliverySetting["values"] as $channel) {
            $configureChannels[] = $channel["channel"];
        }
        foreach ($param["channels"] as $channel) {
            if (!in_array($channel["channel"], $configureChannels)) {
                return $this->renderSuccess('外送渠道未设置');
            }
        }
        $companySetting = Supplier::where("uuid", $param['uuid'])->find();

        if (!$companySetting->allowField(['delivery_config'])->save([
            'delivery_config' => json_encode($param["channels"]),
        ])) {
            return $this->renderError('保存失败');
        }
        return $this->renderSuccess('保存成功');
    }


    /**
     * @Apidoc\Title("外送台账")
     * @Apidoc\Desc("")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/delivery/ledger")
     * @Apidoc\Param("uuid", type="biginteger", require=true, desc="商家ID")
     * @Apidoc\Param("channel", type="string", require=false, desc="外送渠道")
     * @Apidoc\Param("month", type="string", require=false, desc="统计月份")
     * @Apidoc\Returned("list", type="object", desc="分页订单列表")
     * @Apidoc\Returned("list.total", type="int", desc="总记录数")
     * @Apidoc\Returned("list.per_page", type="int", desc="每页数量")
     * @Apidoc\Returned("list.current_page", type="int", desc="当前页码")
     * @Apidoc\Returned("list.last_page", type="int", desc="最后一页页码")
     * @Apidoc\Returned("list.data", type="array", desc="订单数据列表")
     * @Apidoc\Returned("list.data[].uuid", type="biginteger", desc="商家ID")
     * @Apidoc\Returned("list.data[].name", type="string", desc="商家名称")
     * @Apidoc\Returned("list.data[].channel", type="string", desc="外送渠道")
     * @Apidoc\Returned("list.data[].order_no", type="string", desc="订单号")
     * @Apidoc\Returned("list.data[].delivery_fee", type="decimal", desc="订单配送费")
     * @Apidoc\Returned("list.data[].distance", type="decimal", desc="距离（公里）")
     * @Apidoc\Returned("list.data[].basic_fee", type="decimal", desc="基础服务费")
     * @Apidoc\Returned("list.data[].base_delivery_fee", type="decimal", desc="起步配送费")
     * @Apidoc\Returned("list.data[].price_per_km", type="decimal", desc="距离单价")
     * @Apidoc\Returned("aggregate", type="object", desc="汇总信息")
     * @Apidoc\Returned("aggregate.order_count", type="int", desc="总订单数")
     * @Apidoc\Returned("aggregate.delivery_fee", type="string", desc="总配送费")
     * @Apidoc\Returned("aggregate.channels", type="array", desc="各渠道配送费")
     * @Apidoc\Returned("aggregate.channels[].channel", type="string", desc="渠道名称")
     * @Apidoc\Returned("aggregate.channels[].fee", type="string", desc="渠道配送费")
     * @Apidoc\Returned("aggregate.status", type="string", desc="结清状态：settle-已结清；unsettle-未结清")
     */
    public function ledger()
    {
        $param = $this->getData();
        if (!isset($param['uuid'])) {
            return $this->renderError('请选择商家');
        }

        // ToDo 获取数据
        return $this->renderSuccess('', []);
    }
}

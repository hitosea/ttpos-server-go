<?php

namespace app\admin\controller;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\validate\DeliveryValidate;
use app\common\enum\settings\SettingEnum;
use app\admin\model\settings\Setting as SettingModel;
use app\common\model\supplier\Supplier;
use app\admin\model\app\App as AppModel;
use app\common\model\order\Order;
use app\common\model\user\DeliveryLedgerSettle;
use app\common\model\user\MemberSaleOrder;
use PhpOffice\PhpSpreadsheet\IOFactory;
use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;
use app\common\enum\order\MemberSaleOrderStatusEnum;
use PhpOffice\PhpSpreadsheet\Cell\DataType;


/**
 * 外送管理
 * @Apidoc\Group("base")
 * @Apidoc\Sort(6)
 */
class Delivery extends Controller
{

    /**
     * @Apidoc\Title("外送设置")
     * @Apidoc\Desc("get请求是获取，返回数组；post请求是提交修改，传递{channels:[渠道配置列表]}")
     * @Apidoc\Method("GET,POST")
     * @Apidoc\Url("/api/admin/delivery/index")
     * @Apidoc\Param("channels", type="array", require=true, desc="外送渠道列表", children={
     *      @Apidoc\Param("channel", type="string", require=true, desc="配送渠道名称"),
     *      @Apidoc\Param("basic_fee", type="decimal", require=true, desc="基础服务费"),
     *      @Apidoc\Param("base_delivery_fee", type="decimal", require=true, desc="起步配送费"),
     *      @Apidoc\Param("rider_acceptance_timeout", type="int", require=true, desc="骑手未接单取消时间：分钟，1-60"),
     *      @Apidoc\Param("distance_range", type="array", require=true, desc="距离范围配置", children={
     *          @Apidoc\Param("end", type="int", require=true, desc="结束距离"),
     *          @Apidoc\Param("is_unlimited", type="boolean", require=true, desc="是否最大范围"),
     *          @Apidoc\Param("price_per_km", type="decimal", require=true, desc="每公里价格"),
     *      }),
     * })
     */
    public function index(DeliveryValidate $validate)
    {
        if ($this->request->isGet()) {
            return $this->fetchData();
        }
        $param = $validate->goCheck('edit');
        $deliverySetting = SettingModel::where(['key' => SettingEnum::DELIVERY_CONFIG])->find();

        if (is_null($deliverySetting)) {
            SettingModel::insert([
                "key" => SettingEnum::DELIVERY_CONFIG,
                "describe" => "外送设置",
                "values" => json_encode($param['channels'])
            ]);
        } else {
            $deliverySetting->save([
                'describe' => SettingEnum::data()[SettingEnum::DELIVERY_CONFIG]['describe'],
                'values' =>  $param['channels'],
            ]);
            // 获取 company_setting 表，如果设置了自动同步，需要更新，以及对应的商家数据库
            $companySettings = Supplier::select();
            foreach ($companySettings as $companySetting) {
                if (!$companySetting || !$companySetting->delivery_config) {
                    continue;
                }
                $companyChannels = json_decode($companySetting->delivery_config, true);
                foreach ($companyChannels as $k => $companyChannel) {
                    foreach ($param['channels'] as $paramChannel) {
                        if ($companyChannel["channel"] == $paramChannel["channel"] && $companyChannel["config_type"] == "auto_sync") {
                            $companyChannels[$k] = $paramChannel + ["config_type" => "auto_sync"];
                        }
                    }
                }
                $companySetting->save(['delivery_config' => json_encode($companyChannels)]);
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
     * @Apidoc\Returned(type="array", desc="外送渠道选项", children={
     *      @Apidoc\Returned("name", type="string", desc="渠道名称"),
     *      @Apidoc\Returned("value", type="string", desc="渠道值"),
     * })
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
     * @Apidoc\Returned("delivery_config", type="array", desc="外送渠道配置列表", children={
     * @Apidoc\Returned("channel", type="string", desc="配送渠道名称"),
     *      @Apidoc\Returned("basic_fee", type="decimal", desc="基础服务费"),
     *      @Apidoc\Returned("base_delivery_fee", type="decimal", desc="基础配送费"),
     *      @Apidoc\Returned("rider_acceptance_timeout", type="int", desc="骑手接单超时时间（分钟）"),
     *      @Apidoc\Returned("distance_range", type="array", desc="距离区间配置", children={
     *          @Apidoc\Returned("end", type="int", desc="区间结束公里数"),
     *          @Apidoc\Returned("price_per_km", type="decimal", desc="每公里价格"),
     *          @Apidoc\Returned("is_unlimited", type="bool", desc="是否为无限区间"),
     *      }),
     *     @Apidoc\Returned("config_type", type="string", desc="配置类型：auto_sync-自动同步；manual手动设置"),
     *     @Apidoc\Returned("channel_names", type="array", desc="渠道名称，分号分隔", children={}),
     *     @Apidoc\Returned("channel_config_types", type="array", desc="配置类型：auto_sync-自动同步；manual手动设置，分号分隔", children={}),
     * })
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
     * @Apidoc\Param("keyword", type="string", default="", desc="商家名称/uuid")
     * @Apidoc\Param("channel", type="string", default="", desc="外送渠道：SKootar") 
     * @Apidoc\Param("configured", type="string", default="", desc="是否只获取已配置外送渠道：0-否；1-是") 
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned(type="array", desc="商家列表", children={
     *      @Apidoc\Returned("uuid", type="biginteger", desc="商家ID"),
     *      @Apidoc\Returned("name", type="string", desc="商家名称"),
     *      @Apidoc\Returned("channel_names", type="array", desc="商家已配置的外送渠道名称列表，如果长度不为空，表示已配置"),
     * })
     */
    public function companySelect()
    {
        $param = $this->getData();
        $param = array_merge($param, ["configured" => ($param['configured'] ?? '0') == '1']);
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
     * @Apidoc\Param("channels", type="array", require=true, desc="外送渠道配置列表", children={
     *      @Apidoc\Param("channel", type="string", require=true, desc="配送渠道名称"),
     *      @Apidoc\Param("config_type", type="string", require=true, desc="配置类型，如 auto_sync-自动同步；manual-手动设置"),
     *      @Apidoc\Param("basic_fee", type="decimal", require=true, desc="基础服务费"),
     *      @Apidoc\Param("base_delivery_fee", type="decimal", require=true, desc="起步配送费"),
     *      @Apidoc\Param("rider_acceptance_timeout", type="int", require=true, desc="骑手未接单取消时间：分钟，1-60"),
     *      @Apidoc\Param("distance_range", type="array", require=true, desc="距离范围配置", children={
     *          @Apidoc\Param("end", type="int", require=true, desc="结束距离"),
     *          @Apidoc\Param("is_unlimited", type="boolean", require=true, desc="是否最大范围"),
     *          @Apidoc\Param("price_per_km", type="decimal", require=true, desc="每公里价格"),
     *      }), 
     * }) 
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
     * @Apidoc\Param("month", type="string", require=true, desc="统计月份，格式：2025-01")
     * 
     * @Apidoc\Returned("list", type="object", desc="分页订单列表", children={
     *      @Apidoc\Returned("total", type="int", desc="总记录数"),
     *      @Apidoc\Returned("per_page", type="int", desc="每页数量"),
     *      @Apidoc\Returned("current_page", type="int", desc="当前页码"),
     *      @Apidoc\Returned("last_page", type="int", desc="最后一页页码"),
     *      @Apidoc\Returned("data", type="array", desc="订单数据列表", children={
     *          @Apidoc\Returned("uuid", type="biginteger", desc="商家ID"),
     *          @Apidoc\Returned("company_name", type="string", desc="商家名称"),
     *          @Apidoc\Returned("channel", type="string", desc="外送渠道"),
     *          @Apidoc\Returned("channel_order_no", type="string", desc="外送渠道订单号"),
     *          @Apidoc\Returned("order_no", type="string", desc="订单号"),
     *          @Apidoc\Returned("delivery_fee", type="decimal", desc="订单配送费"),
     *          @Apidoc\Returned("distance", type="decimal", desc="距离（公里）"),
     *          @Apidoc\Returned("basic_fee", type="decimal", desc="基础服务费"),
     *          @Apidoc\Returned("base_delivery_fee", type="decimal", desc="起步配送费"),
     *          @Apidoc\Returned("price_per_km", type="decimal", desc="距离单价"),
     *      }),
     *      @Apidoc\Returned("aggregate", type="object", desc="汇总信息", children={
     *          @Apidoc\Returned("order_count", type="int", desc="总订单数"),
     *          @Apidoc\Returned("delivery_fee_amount", type="decimal", desc="总配送费"),
     *          @Apidoc\Returned("channel_data", type="array", desc="各渠道配送费", children={
     *              @Apidoc\Returned("channel", type="string", desc="渠道名称"),
     *              @Apidoc\Returned("amount", type="decimal", desc="渠道配送费"),
     *          }),
     *          @Apidoc\Returned("is_settle", type="int", desc="结清状态：1-已结清；0-未结清"),    
     *      }),
     * })
     */
    public function ledger()
    {
        $param = $this->getData();
        if (!isset($param['uuid'])) {
            return $this->renderError('请选择商家');
        }
        if (!isset($param['month'])) {
            return $this->renderError('请选择月份');
        }
        if (!preg_match('/^\d{4}-\d{2}$/', $param['month'])) {
            return $this->renderError('月份格式不正确，格式：2025-01');
        }

        $param['start_time'] = strtotime(date('Y-m-01 00:00:00', strtotime($param['month'])));
        $param['end_time'] = strtotime(date('Y-m-t 23:59:59', strtotime($param['month'])));

        $companyUuid = $param['uuid'];
        $company = AppModel::where("uuid", $companyUuid)->find();
        if (!$company) {
            return $this->renderError('商家不存在');
        }
        $list = (new MemberSaleOrder([], $companyUuid))->setAppId($companyUuid)->getList($param)?->toArray();
        $uuids = [];
        $data = [];
        foreach ($list['data'] as $k => $item) {
            $uuids[] = $item['uuid'];
            $data[] = [
                'uuid' => $item['uuid'], // 外送订单ID
                'company_name' => $company->name, // 商家名称
                'channel' => $item['related_order_type'], // 外送渠道
                'channel_order_no' => $item['related_order_no'], // 外送渠道订单号
                'delivery_fee' => $item['delivery_fee_amount'], // 配送费
                'distance' => $item['delivery_distance'], // 距离
                'basic_fee' => $item['delivery_fee_base_fee'], // 基础服务费
                'base_delivery_fee' => $item['delivery_fee_base_fee'], // 起步配送费
                'price_per_km' => $item['delivery_fee_per_km'], // 距离单价
                'order_no' => '', // 订单号
            ];
        }

        $saleBills = Order::whereIn('member_sale_order_uuid', $uuids)->select();

        $saleOrderData = [];
        foreach ($saleBills as $saleBill) {
            $saleOrderData[$saleBill->member_sale_order_uuid] = $saleBill->order_no;
        }

        foreach ($data as $k => $item) {
            if (isset($saleOrderData[$item['uuid']])) {
                $data[$k]['order_no'] = $saleOrderData[$item['uuid']];
            }
        }
        $statistics = (new MemberSaleOrder())->getMonthStatistics($param);

        request()->appId = 0;
        $statistics['is_settle'] =  DeliveryLedgerSettle::where('month', $param['month'])->where('company_uuid', $companyUuid)->find() ? 1 : 0;

        $list['data'] = $data;

        return $this->renderSuccess('', compact('list', 'statistics'));
    }

    /**
     * @Apidoc\Title("结清上月外送台账")
     * @Apidoc\Desc("")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/delivery/settle")
     * @Apidoc\Param("uuid", type="biginteger", require=true, desc="商家ID") 
     * @Apidoc\Param("month", type="string", require=true, desc="统计月份，格式：2025-01")
     */
    public function settle()
    {
        $param = $this->postData();
        if (!isset($param['uuid'])) {
            return $this->renderError('请选择商家');
        }
        if (!isset($param['month'])) {
            return $this->renderError('请选择月份');
        }
        if (!preg_match('/^\d{4}-\d{2}$/', $param['month'])) {
            return $this->renderError('月份格式不正确，格式：2025-01');
        }
        // 月份不能大于当前月份
        if (strtotime($param['month']) > strtotime(date('Y-m-d'))) {
            return $this->renderError('月份不能大于当前月份');
        }
        $param['start_time'] = strtotime(date('Y-m-01 00:00:00', strtotime($param['month'])));
        $param['end_time'] = strtotime(date('Y-m-t 23:59:59', strtotime($param['month'])));

        $statistics = (new MemberSaleOrder([], $param['uuid']))->setAppId($param['uuid'])->getMonthStatistics($param);
        request()->appId = 0;
        $deliveryLedgerSettle =  DeliveryLedgerSettle::where('month', $param['month'])->where('company_uuid', $param['uuid'])->find();

        if (!$deliveryLedgerSettle) {
            $deliveryLedgerSettle = new DeliveryLedgerSettle();
            $deliveryLedgerSettle->month = $param['month'];
            $deliveryLedgerSettle->company_uuid = $param['uuid'];
            $deliveryLedgerSettle->order_count = $statistics['order_count'];
            $deliveryLedgerSettle->delivery_fee_amount = $statistics['delivery_fee_amount'];
            $deliveryLedgerSettle->channel_data = json_encode($statistics['channel_data']);
            $deliveryLedgerSettle->save();
        }

        return $this->renderSuccess('');
    }

    /**
     * @Apidoc\Title("导出外送台账")
     * @Apidoc\Desc("")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/delivery/export")
     * @Apidoc\Param("uuid", type="biginteger", require=true, desc="商家ID") 
     * @Apidoc\Param("month", type="string", require=true, desc="统计月份，格式：2025-01")
     */
    public function export()
    {
        $param = $this->postData();
        if (!isset($param['uuid'])) {
            return $this->renderError('请选择商家');
        }
        if (!isset($param['month'])) {
            return $this->renderError('请选择月份');
        }
        if (!preg_match('/^\d{4}-\d{2}$/', $param['month'])) {
            return $this->renderError('月份格式不正确，格式：2025-01');
        }
        // 月份不能大于当前月份
        if (strtotime($param['month']) > strtotime(date('Y-m-d'))) {
            return $this->renderError('月份不能大于当前月份');
        }
        $param['start_time'] = strtotime(date('Y-m-01 00:00:00', strtotime($param['month'])));
        $param['end_time'] = strtotime(date('Y-m-t 23:59:59', strtotime($param['month'])));

        $companyUuid = $param['uuid'];
        $company = AppModel::where("uuid", $companyUuid)->find();
        if (!$company) {
            return $this->renderError('商家不存在');
        }

        $statistics = (new MemberSaleOrder([], $companyUuid))->setAppId($companyUuid)->getMonthStatistics($param);

        $list = MemberSaleOrder::where('status', MemberSaleOrderStatusEnum::FINISHED)->order('create_time', 'desc')->select(); 
        $listArr = $list?$list->toArray():[]; 
        $saleBills = Order::whereIn('member_sale_order_uuid',  array_column($listArr, 'uuid'))->select();
        $saleOrderData = [];
        foreach ($saleBills as $saleBill) {
            $saleOrderData[$saleBill->member_sale_order_uuid] = $saleBill->order_no;
        }
  
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet(); 
        // 列宽
        $sheet->getColumnDimension('A')->setWidth(18);
        $sheet->getColumnDimension('B')->setWidth(18);
        $sheet->getColumnDimension('C')->setWidth(18);
        $sheet->getColumnDimension('D')->setWidth(18);
        $sheet->getColumnDimension('E')->setWidth(18);
        $sheet->getColumnDimension('F')->setWidth(18);
        $sheet->getColumnDimension('G')->setWidth(18);
        $sheet->getColumnDimension('H')->setWidth(18);
        $sheet->getColumnDimension('I')->setWidth(18); 
        $sheet->setCellValue('A1', date(__('商家名称')));
        $sheet->setCellValue('B1', date(__('外送渠道')));
        $sheet->setCellValue('C1', date(__('渠道订单编号')));
        $sheet->setCellValue('D1', date(__('订单编号')));
        $sheet->setCellValue('E1', date(__('订单配送费')));
        $sheet->setCellValue('F1', date(__('距离(公里)')));
        $sheet->setCellValue('G1', date(__('基础服务费')));
        $sheet->setCellValue('H1', date(__('起步配送费')));
        $sheet->setCellValue('I1', date(__('距离单价'))); 
        foreach ($list as $key => $item) {
            $sheet->setCellValue('A' . ($key + 2), $company->name); // 商家名称
            $sheet->setCellValue('B' . ($key + 2), $item->related_order_type); // 外送渠道
            $sheet->setCellValue('C' . ($key + 2), $item->related_order_no); // 渠道订单编号
            $sheet->setCellValueExplicit('D' . ($key + 2), (string)$saleOrderData[$item->uuid] ?? '', DataType::TYPE_STRING); // 订单编号
            $sheet->setCellValue('E' . ($key + 2), $item->delivery_fee_amount); // 订单配送费
            $sheet->setCellValue('F' . ($key + 2), $item->delivery_distance); // 距离(公里)
            $sheet->setCellValue('G' . ($key + 2), $item->delivery_fee_base_fee); // 基础服务费
            $sheet->setCellValue('H' . ($key + 2), $item->delivery_fee_min_fee); // 起步配送费
            $sheet->setCellValue('I' . ($key + 2), $item->delivery_fee_per_km); // 距离单价
        }

        $n = count($list) + 3;
        $sheet->setCellValue('A'.($n+0), date(__('总订单数')));
        $sheet->setCellValue('B'.($n+0), $statistics['order_count']);

        $sheet->setCellValue('A'.($n+1), date(__('总配送费')));
        $sheet->setCellValue('B'.($n+1), $statistics['delivery_fee_amount']); 

        foreach ($statistics['channel_data'] as $k => $item) {
            $sheet->setCellValue('A'.($n+2+$k), $item['channel']);
            $sheet->setCellValue('B'.($n+2+$k), $item['amount']);
        } 
        $sheet->setCellValue('A'.($n+count($statistics['channel_data'])+2), date(__('结清状态'))); 
        request()->appId = 0;
        $isSettle =  DeliveryLedgerSettle::where('month', $param['month'])->where('company_uuid', $companyUuid)->find() ? __("是") : __("否");
        $sheet->setCellValue('B'.($n+count($statistics['channel_data'])+2), $isSettle);

         // 保存文件
         $writer = new Xlsx($spreadsheet);
         $filename = __('外送账单') . '.xlsx';
         header('Content-Type: application/vnd.ms-excel');
         header('Content-Disposition: attachment;filename="' . $filename . '"');
         header('Cache-Control: max-age=0');
         $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
         $writer->save('php://output');

    }
}

<?php

namespace app\common\model\settings;

use help\RabbitmqHelp;
use app\job\queue\PrintJob;
use app\common\model\BaseModel;
use app\job\service\QueueService;
use app\common\library\printer\Driver;
use app\common\model\supplier\Supplier;
use app\common\enum\settings\SettingEnum;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\model\order\Order as OrderModel;
use app\common\model\settings\Setting as SettingModel;
use app\common\library\printer\Driver as PrinterDriver;

/**
 * 打印机日志
 */
class PrinterLog extends BaseModel
{
    protected $name = 'printer_log';
    protected $pk = 'id';

    const DATA_TYPE = [
        1 => ['value' => 1, 'text' => "预结账单"],
        2 => ['value' => 2, 'text' => "结账单"],
        3 => ['value' => 3, 'text' => "一菜一单"],
        4 => ['value' => 4, 'text' => "整单打印"],
        5 => ['value' => 5, 'text' => "打印发票"],
        6 => ['value' => 6, 'text' => "打印营业数据"],
        7 => ['value' => 7, 'text' => "打印交班单"],
        8 => ['value' => 8, 'text' => "充值单"],
        9 => ['value' => 9, 'text' => "退菜单"],
    ];

    /**
     * 获取打印机类型列表
     */
    public static function getPrinterTypeList()
    {
        static $printerTypeEnum = [];
        if (empty($printerTypeEnum)) {
            $printerTypeEnum = PrinterTypeEnum::getTypeNames();
        }
        return $printerTypeEnum;
    }

    /**
     * 获取原因
     */
    public function getReasonAttr($value)
    {
        return $value == '打印成功' ? '' : __($value ?: '');
    }

    /**
     * 格式化打印数据
     */
    public function setDataAttr($value)
    {
        if (ctype_xdigit($value)) {
            $binaryData = hex2bin($value);                  // 将16进制数据转换为二进制数据
            $compressedData = gzcompress($binaryData);      // 压缩二进制数据
            return bin2hex($compressedData);                // 将压缩后的二进制数据转换为16进制数据
        }
        return $value;
    }

    /**
     * 格式化打印数据
     */
    public function getDataAttr($value)
    {
        try {
            $compressedData = hex2bin($value);                  // 将16进制数据转换为二进制数据
            $uncompressedData = gzuncompress($compressedData);  // 解压缩二进制数据
            return bin2hex($uncompressedData);                  // 将还原后的二进制数据转换为16进制数据
        } catch (\Throwable $th) {
            return $value;
        }
    }

    /**
     *打印时间格式化
     */
    public function getPrinterTimeAttr($value)
    {
        return date('Y-m-d H:i:s', $value);
    }

    /**
     * 打印机名称
     */
    public function getPrinterNameAttr($value, $data)
    {
        if (!$value && ($data['printer_id'] ?? 1) <= 0) {
            return "-";
        }
        return $value ?: "-";
    }

    /**
     * 打印机类型
     */
    public function getPrinterTypeAttr($value, $data)
    {
        if (($data['printer_id'] ?? 1) <= 0) {
            return "CASHIER";
        }
        return $value['value'] ?? 'CASHIER';
    }

    /**
     * 获取状态文本
     */
    public function getStatusTextAttr($value, $data)
    {
        $status = $data['status'] ?? 0;
        $num = $data['num'] ?? 1;
        switch ($status) {
            case 0:
                return $num > 1 ? (__('补打失败') . " ($num)") : __('失败');
                break;
            case 1:
                return __('进行中') . " ($num)";
                break;
            case 2:
                return  $num > 1 ? (__('补打成功') . " ($num)") : __('成功');
                break;
        }
    }

    /**
     * 获取打印类型文本
     */
    public function getDataTypeTextAttr($value, $data)
    {
        return ($data['data_type'] ?? '') ? __(self::DATA_TYPE[$data['data_type']]['text']) : '';
    }


    /**
     * 关联打印配置
     */
    public function printer()
    {
        return $this->hasOne('app\\common\\model\\settings\\Printer', 'printer_id', 'printer_id');
    }

    /**
     * 关联打印机名称
     */
    public function printerName()
    {
        return $this->hasOne('app\\common\\model\\settings\\Printer', 'printer_id', 'printer_id')->bind(['printer_name']);
    }

    /**
     * 关联打印机配置相关
     */
    public function printerConfig()
    {
        return $this->hasOne('app\\common\\model\\settings\\Printer', 'printer_id', 'printer_id')->bind(['printer_name', 'printer_config', 'printer_type', 'print_times']);
    }

    /**
     * 关联订单取餐号
     */
    public function noName()
    {
        return $this->hasOne('app\\common\\model\\order\\Order', 'order_id', 'order_id')->field(['order_id', "if(table_no!='', table_no, call_no) as no"])->bind(['no']);
    }

    /**
     * 关联子级
     */
    public function printerLogsGroup()
    {
        return $this->hasMany('app\\common\\model\\settings\\PrinterLog', 'printer_id', 'printer_id');
    }

    /**
     * 添加打印日志
     */
    public static function addPrinterLog($printer, array $printerLogData = [], $controls_device_id = '')
    {
        $printerLogData['status'] = 1;
        $printerLogData['reason'] = '';
        $printerLogData['type'] = $printerLogData['type'] ?? 0;
        $printerLogData['first_execution'] = $printerLogData['first_execution'] ?? 0;
        $printerLogData['cashier_bind_key'] = $printerLogData['cashier_bind_key'] ?? '';

        // 是否本机
        $isLocal = $controls_device_id == $printerLogData['cashier_bind_key'];

        // 如果是点餐助手操作的，就永远不是本机
        if (app('http')->getName() == 'assistant') {
            $isLocal = false;
        }

        // 是否队列服务
        $isQueueService = false;

        // 如果是局域网部署 - 就都下放打印
        if (env('IS_CLOUD_DEPLOY', false)) {
            $printerLogData['type'] = 1;
        } else if (!$printerLogData['cashier_bind_key'] || (!$isLocal && ($printerLogData['printer_id'] ?? 0))) {
            $isQueueService = true;
        }

        // 如果是商米云打印 -  就都队列打印
        if (($printer?->printer_type['value'] ?? '') == PrinterTypeEnum::SUNMI_CLOUD && Supplier::where('shop_supplier_id', $printerLogData['shop_supplier_id'])->value('is_open_local_print') == 0) {
            $isQueueService = true;
            $printerLogData['type'] = 0;
            $printerLogData['first_execution'] = 0;
        }

        // 逻辑处理
        if (!$printer) {
            $printerLogData['status'] = 0;
            $printerLogData['reason'] =  "打印机不存在";
        } else {
            // 记录打印方式  = 1 文本打印，2图片打印
            $printerConfig = SettingModel::getSupplierItem(SettingEnum::PRINTER, $printerLogData['shop_supplier_id'], $printerLogData['app_id']);
            if ($printerLogData['data_type'] == self::DATA_TYPE[3]['value'] || $printerLogData['data_type'] == self::DATA_TYPE[4]['value'] || $printerLogData['data_type'] == self::DATA_TYPE[9]['value']) {
                $printerLogData['print_method'] = ($printerConfig['kitchen_print_method'] ?? 1) ?: 1;
            } else {
                $printerLogData['print_method'] = ($printerConfig['print_method'] ?? 1) ?: 1;
            }
            //
            if (!$isQueueService) {
                // 直接打印
                if ($printerLogData['printer_id'] && $printerLogData['type'] == 0 && $printerLogData['first_execution'] == 1) {
                    $driver = new PrinterDriver($printer);
                    if ($driver->printTicket($printerLogData['data'], $printerLogData['print_method'])) {
                        $printerLogData['status'] = 2;
                        $printerLogData['reason'] = '打印成功';
                    } else {
                        $printerLogData['status'] = 0;
                        $printerLogData['reason'] = $driver->getError() ?: '打印失败';
                    }
                    $printerLogData['num'] = 1;
                }
                // 返回数据下放给本地收银机去打印，都是成功状态
                else if ($printerLogData['type'] == 1 && $printerLogData['first_execution'] == 1 && $isLocal) {
                    $printerLogData['status'] = 2;
                }
            }
        }

        // 保存数据
        $printerLogData['printer_time'] = time();
        $printerLog = new self;
        $printerLog->save($printerLogData, 'id');

        // 只保留7天的数据
        (new self)->where('create_time', '<', strtotime('-7 days'))->delete();

        // 添加队列打印
        if ($printer && $isQueueService) {
            QueueService::push(PrintJob::class, $printerLog->app_id, [
                'id' => $printerLog->id,
                'printer_id' => $printerLog->printer_id,
                'print_method' => $printerLog->print_method,
            ]);
            return $printerLog->id;
        }
        // 返回给前端打印
        else if ($printer && $printerLogData['first_execution'] == 1 && $printerLogData['type'] != 0 && $isLocal) {
            $printerLogData['id'] = $printerLog->id;
            $printerLogData['print_times'] = $printer?->print_times ?? 1;
            $printerLogData['printer_type'] = $printer?->printer_type['value'] ?? 'CASHIER';
            $printerLogData['printer_config'] = $printer?->printer_config ?? '{}';
            $printerLogData['create_time'] = date('Y-m-d H:i:s',  $printerLogData['printer_time']);
            $printerLogData['no'] = ($printerLogData['order_id'] ?? '') ? OrderModel::where('order_id', $printerLogData['order_id'])->value("if(table_no!='',table_no, call_no)") : '';
            return $printerLogData;
        }
        // 执行前端定时获取打印
        else {
            return (!$isLocal || $printerLog->status == 2) ? $printerLog->id : false;
        }
    }

    /**
     * 获取静态打开钱箱配置
     */
    public static function getStaticPrinterConfig($appId, $shopSupplierId, $deviceId)
    {
        $setting = SettingModel::getAll($appId, $shopSupplierId);
        // 获取打印信息
        $printerInfo = SettingModel::getPrinterInfo($setting[SettingEnum::PRINTER]['values'], $deviceId);
        $printer = $printerInfo['printer'];
        if (!$printer) {
            return false;
        }
        if (!($printer?->printer_type['value'] ?? '')) {
            return false;
        }

        // 是否商米打印机
        $isSunmi = in_array($printer['printer_type']['value'] ?? '', [PrinterTypeEnum::SUNMI_LAN, PrinterTypeEnum::SUNMI_CLOUD]);

        // 打印
        return [
            'id' => time(),
            "no" => '',
            "reason" => '',
            "data" => $isSunmi ? '1b700019fa' : '1014010001',
            'print_times' => $printer?->print_times ?? 1,
            'printer_type' => $printer?->printer_type['value'] ?? 'CASHIER',
            'printer_config' => $printer?->printer_config ?? '{}',
            'print_method' => 1,
        ];
    }

    /**
     * 添加到打印队列
     */
    public function addPrinterPush()
    {
        if ($this->getData('type') == 0) {
            $printers = Printer::where('is_delete', 0)->order('printer_id')->select();
            foreach ($printers as $key => $printer) {
                if ($printer->printer_id == $this->printer_id) {
                    RabbitmqHelp::push('print-data-system-' . ($key > 10 ? 0 : $key), $this->toArray());
                }
            }
        } else {
            RabbitmqHelp::push('print-data-' . $this->shop_supplier_id . '-' . $this->cashier_bind_key, $this->toArray());
        }
    }

    /**
     * 获取列表记录
     */
    public function getList($params, int $shopSupplierId = 0)
    {
        return $this->withoutGlobalScope()
            ->with(['printerName', "noName"])
            ->field('id, reason, printer_id, order_id, create_time')
            ->where('status', 0)
            ->where('type', 0)
            ->order('create_time', 'desc')
            ->paginate($params);
    }

    /**
     * 未处理数量
     */
    public function getUnprocessedCount(int $shopSupplierId = 0)
    {
        return $this->withoutGlobalScope()
            ->where('status', 0)
            ->where('type', 0)
            ->count();
    }

    /**
     * 获取失败列表
     */
    public function getErrorList(int $shopSupplierId = 0)
    {
        return $this->withoutGlobalScope()
            ->with(['printerName', "noName"])
            ->field('id, reason, printer_id, cashier_bind_key, order_id, create_time')
            ->where('status', 0)
            ->where('first_execution', 0)
            ->where('type', 0)
            ->limit(10)
            ->select();
    }

    /**
     * 获取客户端打印数据
     */
    public static function getClientPrintDataList(int $shopSupplierId = 0, string $deviceId = '')
    {
        $result = self::alias('a')
            ->with(['printerConfig', 'noName'])
            ->field('a.id, a.data, a.reason, a.printer_id, a.order_id, a.create_time, a.print_method')
            ->leftJoin('printer_read_log read', 'a.id = read.log_id')
            ->whereRaw('read.id is null')
            ->where('a.create_time', '>', strtotime('-1 days'))
            ->where('a.type', 1)
            ->where('a.status', 1)
            ->where('a.num', 'in', [0, 1])
            ->where('a.shop_supplier_id', $shopSupplierId)
            ->where(function ($q) use ($deviceId) {
                $q->where('a.cashier_bind_key', $deviceId);
                $q->whereOr('a.cashier_bind_key', '');
            })
            ->limit(50)
            ->order('a.id')
            ->select();
        //
        $results = $result->toArray();
        //
        $logs = [];
        foreach ($result as $val) {
            $logs[] = [
                'log_id' => $val['id'],
                'device_id' => $deviceId
            ];
            //
            $val->num = $val->num + 1;
            $val->status = $val->num > 5 ? 0 : 2;
            // 去除开箱指令
            if (substr($val->data, -10) === '1b700019fa') {
                $val->data = substr($val->data, 0, -10);
            } else if (substr($val->data, -10) === '1014010001') {
                $val->data = substr($val->data, 0, -10);
            }
            //
            $val->save();
        }
        if (!empty($logs)) {
            (new PrinterReadLog)->saveAll($logs);
        }
        //
        return $results;
    }

    /**
     * 获取列表打印日志
     * @Param("status", type="int", require=false, default="0", desc="状态：0全部, 1成功, 2失败, 3补打成功, 4补打失败")
     * @Param("printer_id", type="int", require=false, default="0", desc="打印机id，0全部")
     * @Param("printer_type", type="int", require=false, default="0", desc="打印类型，0全部")
     * @Param("times", type="array", require=false, default="0", desc="按指定时间查询 ["2024-07-01", "2024-07-17"]")
     */
    public function getPrintLogList($params, int $shopSupplierId = 0)
    {
        $status = $params['status'] ?? 0;
        $printerId = $params['printer_id'] ?? 0;
        $printerType = $params['printer_type'] ?? 0;
        $startTime = 0;
        $endTime = 0;
        // 按指定时间查询
        if (isset($params['times']) && $params['times'] && ($params['times'][0] ?? '') && ($params['times'][1] ?? '')) {
            $startTime = strtotime($params['times'][0]);
            $endTime = strtotime($params['times'][1]) + 86399;
        }
        //
        return $this->withoutGlobalScope()->alias('a')
            ->leftJoin('order o', 'a.order_id = o.order_id')
            ->leftJoin('printer p', 'a.printer_id = p.printer_id')
            ->leftJoin('supplier_printing sp', 'a.printer_rule_id = sp.id')
            ->leftJoin('shop_bind_record sbr', 'a.cashier_bind_key = sbr.key')
            ->field('a.id, a.reason, a.status, a.num, a.printer_id, a.order_id, a.data_type, a.printer_time, a.create_time, a.print_method')
            ->field("sp.name as printer_rule_name")
            ->field("ifnull(p.printer_name, if(sbr.remark = '', sbr.key, sbr.remark)) printer_name")
            ->field("o.order_no, ifnull(if(o.table_no !='' , o.table_no, o.call_no), '') as no")
            ->where('a.shop_supplier_id', $shopSupplierId)
            // ->where('a.type', 0)
            // 按状态
            ->when($status == 1, function ($q) {
                $q->where('a.status', 2)->where('a.num', '<=', 1);
            })
            ->when($status == 2, function ($q) {
                $q->where('a.status', 0)->where('a.num', '<=', 1);
            })
            ->when($status == 3, function ($q) {
                $q->where('a.status', 2)->where('a.num', '>', 1);
            })
            ->when($status == 4, function ($q) {
                $q->where('a.status', 0)->where('a.num', '>', 1);
            })
            //
            ->when($printerId, function ($q) use ($printerId) {
                $q->where(function ($q) use ($printerId) {
                    $q->where('a.printer_id', $printerId);
                    $q->whereOr('a.cashier_bind_key', $printerId);
                });
            })
            //
            ->when($printerType, function ($q) use ($printerType) {
                $q->where('a.data_type', $printerType);
            })
            //
            ->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
                $q->where('a.printer_time', 'between', [$startTime, $endTime]);
            })
            //
            ->order('a.create_time', 'desc')
            ->paginate($params)
            ->append(['status_text', 'data_type_text']);
    }

    /**
     * 打印
     */
    public function print($deviceId = '')
    {
        if ($this->printer_id > 0 && (!$this->printer || $this->printer->is_delete == 1)) {
            $this->error = '打印失败，打印机已不存在';
            return false;
        }
        if (!$this->data) {
            $this->error = '打印失败，打印机已不存在';
            return false;
        }

        // 去除开箱指令
        if (substr($this->data, -10) === '1b700019fa') {
            $this->data = substr($this->data, 0, -10);
            $this->save();
        } else if (substr($this->data, -10) === '1014010001') {
            $this->data = substr($this->data, 0, -10);
            $this->save();
        }

        if ($this->printer_id > 0 && $this->getData('type') == 0) {
            $driver = new Driver($this->printer);
            $result = $driver->printTicket($this->data, $this->print_method);
            if (!$result) {
                $this->error = $driver->getError() ?: '打印失败，未连接打印机';
                return false;
            }
        } else if ($this->getData('type') == 1 && $this->cashier_bind_key && $this->cashier_bind_key != $deviceId) {
            PrinterReadLog::where('device_id', $this->cashier_bind_key)->where('log_id', $this->id)->delete();
            $this->status = 1;
            $this->save();
            return $this->id;
        } else {
            $result = [
                "data" => $this->data,
                "print_method" => $this->print_method,
                "printer_id" => $this->printer_id,
                "printer_time" => $time = time(),
                "create_time" => date('Y-m-d H:i:s', $time),
                "printer_name" => $this->printer?->printer_name ?: '',
                "printer_type" => $this->printer?->getData('printer_type') ?: 'CASHIER',
                "printer_config" => $this->printer?->printer_config ?: '{}',
                "print_times" => 1,
            ];
        }
        if (!$result) {
            $this->error = '打印失败，未连接打印机';
            return false;
        }
        //
        return $result;
    }
}

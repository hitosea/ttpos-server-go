<?php

namespace app\common\tasks;

use app\common\service\order\OrderPrinterService;

/**
 * 打印结账单
 * 使用：Task::deliver(new ImgBillTask('tmp','tmp2'));
 */
class ImgBillTask extends Task
{
    protected $order;
    protected $isQueue;
    protected $deviceId;
    protected $paramData;
    protected $isPrePrint;

    public function __construct($order, $isQueue, $deviceId, $paramData, $isPrePrint)
    {
        $this->order = $order;
        $this->isQueue = $isQueue;
        $this->deviceId = $deviceId;
        $this->paramData = $paramData;
        $this->isPrePrint = $isPrePrint;
    }
    
    public function start()
    {
        (new OrderPrinterService)->printTicket($this->order, $this->isQueue, $this->deviceId, $this->paramData, false, $this->isPrePrint);
    }

    public function end()
    {

    }
}

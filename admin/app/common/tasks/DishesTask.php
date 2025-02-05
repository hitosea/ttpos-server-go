<?php

namespace app\common\tasks;

use app\common\service\order\OrderPrinterService;

/**
 * 送餐任务 - 打印送餐单
 * 使用：Task::deliver(new DishesTask('tmp','tmp2'));
 */
class DishesTask extends Task
{
    protected $order;
    protected $printType;

    public function __construct($order, $printType)
    {
        $this->order = $order;
        $this->printType = $printType;
    }
    
    public function start()
    {
        (new OrderPrinterService)->printProductTicket($this->order, $this->printType, false);
    }

    public function end()
    {

    }
}

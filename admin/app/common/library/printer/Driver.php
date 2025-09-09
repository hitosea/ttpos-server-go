<?php

namespace app\common\library\printer;

use app\common\exception\BaseException;
use app\common\enum\settings\PrinterTypeEnum;

/**
 * 小票打印机驱动
 */
class Driver
{
    private $printer;    // 当前打印机
    private $engine;     // 当前打印机引擎类

    // 打印机引擎列表
    private static $engineList = [
        PrinterTypeEnum::FEI_E_YUN => 'Feie',
        PrinterTypeEnum::FEI_E_YUN_TAG => 'FeieTag',
        PrinterTypeEnum::PRINT_CENTER => 'PrintCenter',
        PrinterTypeEnum::SUNMI_LAN => 'Sunmi',
        PrinterTypeEnum::SUNMI_CLOUD => 'Sunmi',
        PrinterTypeEnum::XPRINTER_LAN => 'Xprinter',
        PrinterTypeEnum::XPRINTER_WIFI => 'Xprinter',
        PrinterTypeEnum::CODESOFT_LAN => 'Xprinter',
        PrinterTypeEnum::CODESOFT_WIFI => 'Xprinter',
        PrinterTypeEnum::GP_CLOUD => 'Xprinter',
    ];

    /**
     * 构造方法
     */
    public function __construct($printer = null)
    {
        // 当前打印机
        $this->printer = $printer;
        // 实例化当前打印机引擎
        if ($printer) {
            $this->engine = $this->getEngineClass();
        }
    }

    /**
     * 执行打印请求
     * @param string $content 打印内容-16进制
     * @return void
     */
    public function timingPrintTicket($content, $printMethod = 1)
    {
        $res = false;
        for ($i = 0; $i < ($this->printer?->print_times ?: 1); $i++) {
            $res = $this->engine->printTicket($content, $printMethod);
        }
        return $res;
    }

    /**
     * 执行打印请求
     * @param string $content 打印内容-16进制
     * @param bool $isQueue 是否加入队列打印，还是直接打印
     * @param bool $orderId 订单id，打印订单相关时需要
     * @param bool $printRule 打印规则
     * @return bool|array
     */
    public function printTicket($content, $printMethod = 1)
    {
        $res = false;
        for ($i = 0; $i < ($this->printer?->print_times ?: 1); $i++) {
            $res = $this->engine->printTicket($content, $printMethod);
        }
        return $res;
    }

    /**
     * 获取错误信息
     */
    public function getError()
    {
        return $this->engine->getError();
    }

    /**
     * 获取当前的打印机引擎类
     */
    private function getEngineClass()
    {
        $engineName = self::$engineList[$this->printer['printer_type']['value']];
        $classSpace = __NAMESPACE__ . "\\engine\\{$engineName}";
        if (!class_exists($classSpace)) {
            throw new BaseException("未找到打印机引擎类: {$engineName}");
        }
        return new $classSpace($this->printer['printer_config'], $this->printer['print_times'], $this->printer);
    }
}

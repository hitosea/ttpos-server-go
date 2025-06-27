<?php

declare(strict_types=1);

namespace app\common\command;

use help\FileHelp;
use help\HttpHelp;
use help\QueueHelp;
use GuzzleHttp\Pool;
use think\facade\Db;
use help\LicenseHelp;
use base\imgs\ImgFont;
use GuzzleHttp\Client;
use think\facade\Lang;
use app\job\event\Sync;
use help\SnowflakeHelp;
use think\facade\Cache;
use help\EncryptionHelp;
use think\console\Input;
use think\console\Output;
use think\console\Command;
use GuzzleHttp\Psr7\Stream;
use GuzzleHttp\Psr7\Request;
use app\shop\model\shop\User;
use app\common\model\shop\Access;
use app\job\service\QueueService;
use app\common\enum\sync\SyncEnum;
use Predis\Client as PredisClient;
use GuzzleHttp\Psr7\MultipartStream;
use app\admin\model\Shop as ShopUser;
use app\common\model\product\Product;
use app\job\service\WebSocketService;
use app\common\library\printer\Driver;
use app\common\model\shop\UserShiftLog;
use app\common\service\sync\SyncService;
use app\common\model\settings\PrinterLog;
use app\common\model\websock\WebSocketMsg;
use app\common\enum\settings\DeliveryTypeEnum;
use app\shop\model_old\order\Order as OrderModel;
use app\cashier\model\order\Order as CashierOrderModel;
use app\common\library\printer\party\SunmiCloudPrinter;
use app\common\model\shop\UserShiftLog as UserShiftLogModel;


// 语言翻译
// ./cmd think Test
class Test extends Command
{

    const PROT_KEY = "dsdd";

    protected function configure()
    {
        // 指令配置
        $this->setName('test')->setDescription('测试');
    }

    protected function execute(Input $input, Output $output)
    {

      
        $s = microtime(true);
        $res = Db::table('shop8609817471094784.ttpos_sale_order_product')->where('sale_bill_uuid', 3672098170077185)->select();

        dump( $res );
        $e = microtime(true);
        dump( $e - $s );
        die;
        request()->appId = 1724054105;
        
        // 订单列表
        $model = new OrderModel([], request()->appId);
        $data = [];
        // 时间模式
        if (!isset($data['time_mode']) || !is_array($data['time_mode'])) {
            $data['time_mode'] = [0]; // 默认开台时间
        }
        // 订单类型
        $dataType = 'all';
        //
        $data['time'] = [];
        $data['order_type'] = 1;
        $data['parent_id'] = 0;
        $data['shop_supplier_id'] = 1724054105;
        $list = $model->getList($dataType, $data);
        foreach ($list as $key => $item) {
            // 是否显示退款按钮 1-显示 0-隐藏
            /** @var OrderModel $item */
            [$list[$key]['is_refund_button'], $list[$key]['is_cancel_button']] = $item->getButtonStatus($item);
            if ($item['subOrder']) {
                foreach ($item['subOrder'] as $subKey => $subItem) {
                    /** @var OrderModel $subItem */
                    [$list[$key]['subOrder'][$subKey]['is_refund_button'], $list[$key]['subOrder'][$subKey]['is_cancel_button']] = $subItem->getButtonStatus($subItem);
                }
            }
            // 拆单主单支付方式去重
            if ($item['parent_id'] == 0 && count($item['subOrder']) > 0) {
                $payTypes = $item['payType']->toArray();
                $uniquePayTypes = [];
                foreach ($payTypes as $payType) {
                    $uniquePayTypes[$payType['value']] = $payType;
                }
                $item['payType'] = new \think\Collection(array_values($uniquePayTypes));
            }
        }
        $order_count = [
            'order_count' => [
                'all' => $model->getCount('all', $data),
                'payment' => $model->getCount('payment', $data),
                'process' => $model->getCount('process', $data),
                'complete' => $model->getCount('complete', $data),
                'cancel' => $model->getCount('cancel', $data),
            ],
        ];
        $ex_style = DeliveryTypeEnum::store();
        dump(compact('list', 'ex_style', 'order_count'));

        die;
        //
        $imageSrc = root_path('runtime/storage') . 'generated_image.png';
        $text = "Address\n發布掰個世界人權宣言发票\n2023-12-15(星期二) 現金引き出し 유효하지 ¥100 เวลาการชำระเงินไม่ถูกต้อง1 Address 發布金引き出し 유효하지 ฿100 9เวลาการชำระเงินไม่ถูกต้อง1";
        //
        $im = new ImgFont(568);
        //
        $im->setImagePadding(30);
        $im->lineFeed(1);
        $im->setAlignment(ImgFont::ALIGN_CENTER);
        $im->appendText("***shop名称***");
        $im->lineFeed(1);
        //
        $im->setTextLineHeight(30);
        $im->setFontSize(30);
        $im->setFontWeight(10);
        $im->lineFeed(1);
        $im->appendText("A01");
        $im->lineFeed(2);
        //
        $im->restoreDefault();
        $im->setImagePadding(30);
        //
        $im->setAlignment(ImgFont::ALIGN_LEFT);
        $im->appendText("订单号");
        $im->setAlignment(ImgFont::ALIGN_RIGHT);
        $im->appendText("N123129172224241", 250);
        $im->lineFeed(1);
        //
        $im->setAlignment(ImgFont::ALIGN_LEFT);
        $im->appendText("收银员");
        $im->setAlignment(ImgFont::ALIGN_RIGHT);
        $im->appendText("管理员");
        $im->lineFeed(1);
        //
        $im->setAlignment(ImgFont::ALIGN_LEFT);
        $im->appendText("时间");
        $im->setAlignment(ImgFont::ALIGN_RIGHT);
        $im->appendText(date('Y/m/d H:i:s', time()) . " 至 " . date('Y/m/d H:i:s', time()), 160);
        $im->lineFeed(1);
        //
        $im->restoreDefault();
        $im->lineFeed(1, 30);
        $im->setFontSize(20);
        $im->setFontWeight(5);
        $im->setTextLineHeight(40);
        $im->printInColumns(
            ["商品", 280, ImgFont::ALIGN_LEFT],
            ["数量", 0, ImgFont::ALIGN_CENTER],
            ["金额", 190, ImgFont::ALIGN_RIGHT]
        );
        $im->appendText("--------------------------------------------------------");
        $im->lineFeed(1);
        $im->setFontSize(20);
        $im->setFontWeight(1);
        $im->setTextLineHeight(40);
        $im->printInColumns(
            ["abcd", 280, ImgFont::ALIGN_LEFT],
            ["2", 0, ImgFont::ALIGN_CENTER, 10, 33],
            ["$11000.21", 190, ImgFont::ALIGN_RIGHT]
        );
        $im->lineFeed(1, 10);
        $im->lineFeed(2);
        //
        // $im->setAlignment(ImgFont::ALIGN_LEFT);
        // $im->appendText($text, 300);
        // $im->lineFeed(2);
        //
        $im->setAlignment(ImgFont::ALIGN_LEFT);
        $im->appendText("管理员");
        $im->setAlignment(ImgFont::ALIGN_RIGHT);
        $im->appendText("abcd", 156);
        $im->lineFeed(1);
        //
        $im->lineFeed(4);
        $orderData = $im->save($imageSrc);

        // 图片打印
        $fp = @fsockopen('192.168.100.180', 9100, $errno, $errstr, 3);
        if ($fp === false) { //连接打印机出错
            dump("连接打印机出错");
            die;
        }
        // 初始化打印机
        fwrite($fp, "\x1B\x40");
        // 文本打印
        fwrite($fp, hex2bin($orderData));
        //
        fwrite($fp, "\x1d\x56\x00");
        //
        // 关闭打印机连接
        fclose($fp);



        // 文本打印
        $printer = new SunmiCloudPrinter(567);
        $printer->appendText("asdjasgdasdasdasd");
        $printer->appendText("asdjasgdasdasdasd");
        $printer->appendText("asdjasgdasdasdasd");
        $printer->lineFeed();
        $printer->printAndExitPageMode();
        $printer->lineFeed(6);
        $printer->cutPaper(false);
        //
        $fp = @fsockopen('192.168.100.180', 9100, $errno, $errstr, 3);
        if ($fp === false) { //连接打印机出错
            dump("连接打印机出错");
            die;
        }
        $content = hex2bin($printer->orderData);
        $content = str_replace("ー", "-", $content);
        $content = iconv("UTF-8", "UTF-8//IGNORE", $content);
        $segments = preg_split('/([\p{Thai}\p{Hangul}฿]+)/u', $content, -1, PREG_SPLIT_DELIM_CAPTURE | PREG_SPLIT_NO_EMPTY);
        foreach ($segments as $segment) {
            if (preg_match('/[\p{Thai}]/u', $segment) || strpos($segment, "฿") !== false) {
                fwrite($fp, "\x1C\x2E");
                fwrite($fp, iconv("UTF-8", "CP874//IGNORE",  $segment));
            } else if (preg_match('/[\p{Hangul}]/u', $segment)) {
                fwrite($fp, "\x1C\x26");
                fwrite($fp, iconv("UTF-8", "CP949//IGNORE",  $segment));
            } else {
                fwrite($fp, "\x1C\x26");
                fwrite($fp, iconv("UTF-8", "GBK//IGNORE",  $segment));
            }
        }
    }

    function convertSqlFormat($input)
    {
        // 读取输入SQL
        $lines = explode(";\n", trim($input));
        $output = [];

        foreach ($lines as $line) {
            if (empty(trim($line))) continue;

            // 提取VALUES部分
            preg_match('/\((.*?)\) VALUES \((.*?)\)/', $line, $matches);
            if (empty($matches)) continue;

            $columns = array_map('trim', explode(',', $matches[1]));
            $values = str_getcsv(str_replace('\'', '"', $matches[2]));

            // 创建新的数据映射
            $data = array_combine($columns, $values);

            // 构建新的SQL
            $newColumns = [
                'id' => $data['`access_id`'],
                'uuid' => $data['`access_id`'],
                'name' => $data['`name`'],
                'path' => $data['`path`'],
                'api_path' => $data['`api_path`'],
                'parent_uuid' => $data['`parent_id`'],
                'sort' => $data['`sort`'],
                'icon' => $data['`icon`'],
                'redirect_name' => $data['`redirect_name`'],
                'is_route' => $data['`is_route`'],
                'is_menu' => $data['`is_menu`'],
                'is_show' => $data['`is_show`'],
                'plus_category_uuid' => $data['`plus_category_id`'],
                'remark' => $data['`remark`'],
                'is_supplier' => $data['`is_supplier`'],
                'create_time' => $data['`create_time`'],
                'update_time' => $data['`update_time`']
            ];

            // 构建新的INSERT语句
            $newSql = "INSERT INTO `ttpos_access` (`" . implode("`, `", array_keys($newColumns)) . "`) VALUES (";

            // 根据字段类型决定是否添加引号
            $formattedValues = array_map(function ($key, $value) {
                // 数字类型字段
                $numericFields = ['id', 'uuid', 'parent_uuid', 'sort', 'is_route', 'is_menu', 'is_show', 'plus_category_uuid', 'is_supplier', 'create_time', 'update_time'];
                if (in_array($key, $numericFields)) {
                    return trim($value, "' ");
                }
                // 字符串类型字段
                return "'" . trim($value, "'") . "'";
            }, array_keys($newColumns), array_values($newColumns));

            $newSql .= implode(", ", $formattedValues);
            $newSql .= ");";

            $output[] = $newSql;
        }

        return implode("\n", $output);
    }
}

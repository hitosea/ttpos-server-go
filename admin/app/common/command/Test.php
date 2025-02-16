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

        dump(signToken(8609817483677696, 'shop', 1, 'admin', 8609817471094784));
        die;

        // 测试雪花ID生成是否重复 - 真实并发测试
        $workerIds = range(1, 3); // 使用3个不同的worker ID
        $iterations = 1000; // 每个worker生成1000个ID
        $processes = [];
        $tempFiles = [];

        // 为每个worker创建一个临时文件存储生成的ID
        foreach ($workerIds as $workerId) {
            $tempFiles[$workerId] = tempnam(sys_get_temp_dir(), 'snowflake_');
        }

        // 创建多个并发进程
        foreach ($workerIds as $workerId) {
            $pid = pcntl_fork();

            if ($pid == -1) {
            die('无法创建进程');
            } else if ($pid == 0) {
            // 子进程
            $snowflake = new SnowflakeHelp($workerId);
            $ids = [];

            for ($i = 0; $i < $iterations; $i++) {
                $ids[] = $snowflake->next();
            }

            // 将生成的ID写入临时文件
            file_put_contents($tempFiles[$workerId], implode("\n", $ids));
            exit(0);
            } else {
            // 父进程记录进程ID
            $processes[] = $pid;
            }
        }

        // 等待所有子进程完成
        foreach ($processes as $pid) {
            pcntl_waitpid($pid, $status);
        }

        // 收集并检查所有生成的ID
        $allIds = [];
        foreach ($tempFiles as $file) {
            $ids = explode("\n", file_get_contents($file));
            foreach ($ids as $id) {
            if (empty($id)) continue;
            if (in_array($id, $allIds)) {
                dump("发现重复ID: " . $id);
                die;
            }
            $allIds[] = $id;
            }
            unlink($file); // 删除临时文件
        }
        dump($allIds);
        dump("并发生成 " . count($allIds) . " 个ID,未发现重复");
        dump("测试通过!");
        die;

        // 使用项目根目录路径
        $rootPath = root_path();
        // 读取输入SQL文件
        $inputSql = file_get_contents($rootPath . 'jjjfood_shop_access.sql');
        $outputSql = $this->convertSqlFormat($inputSql);
        // 保存到数据库初始化文件目录
        $outputPath = $rootPath . 'out_jjjfood_shop_access.sql';
        file_put_contents($outputPath, $outputSql);
        die;

        WebSocketService::publish('1724054105', 'cashier', WebSocketService::MSG_TYPE_UPDATE_PRODUCT, [[
            'ddd' => '1',
            'dddd' => '51325121233',
            'ddddddd' => 'test',
        ]]);
        WebSocketService::publish('1724054105', 'cashier', WebSocketService::MSG_TYPE_UPDATE_PRODUCT, [[
            'ddd' => '1',
            'dddd' => '51325121231',
            'ddddddd' => 'test',
        ]]);
        WebSocketService::publish('1724054105', 'cashier', WebSocketService::MSG_TYPE_UPDATE_PRODUCT, [[
            'ddd' => '1',
            'dddd' => '51325121235',
            'ddddddd' => 'test',
        ]]);

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

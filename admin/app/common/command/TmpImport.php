<?php

declare(strict_types=1);

namespace app\common\command;

use help\StringHelp;
use think\facade\Cache;
use think\console\Input;
use think\console\Output;
use think\console\Command;
use app\common\model\product\Product;
use app\common\model\product\ProductFeed;
use PhpOffice\PhpSpreadsheet\Reader\Xlsx;
use app\common\library\language\engine\OpenAi;

// 语言翻译
// ./cmd think tmpimport
class TmpImport extends Command
{

    protected function configure()
    {
        // 指令配置
        $this->setName('tmpimport')->setDescription('临时导入');
    }

    protected function execute(Input $input, Output $output)
    {
        $appId = 1724054092;
        $filename = '/var/www/sass-tmp-2024-10-01.xlsx';
        $reader = new Xlsx();
        $spreadsheet = $reader->load($filename);
        $worksheet = $spreadsheet->getActiveSheet();
        $values = [];
        // 
        foreach ($worksheet->getRowIterator() as $row) {
            $cellIterator = $row->getCellIterator();
            $cellIterator->setIterateOnlyExistingCells(false); // 遍历所有单元格，即使为空
            foreach ($cellIterator as $key => $cell) {
                if (!($values[$cell->getRow()] ?? 0)) {
                    $values[$cell->getRow()] = [];
                }
                if (in_array($key, ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I'])) {
                    $values[$cell->getRow()][] = $cell->getValue();
                }
            }
        }
        // 
        foreach($values as $k => $value) {
            $values[$k][] = $k;
            if ( empty(array_filter($value, function($value) { return $value !== null;})) || $k == 1) {
                unset($values[$k]);
            }
        }
        // 按商品进行分组
        $groupeds = [];
        foreach ($values as $k => $item) {
            $key = $item[0] . '---' . $item[1];
            if (!isset($groupeds[$key])) {
                $groupeds[$key] = [];
            }
            $groupeds[$key][] = $item;
        }
        // 翻译 
        $productModel = new Product([], $appId);
        // 
        $notData = [];
        foreach ($groupeds as $key => $grouped) {
            $spTh = trim($grouped[0][0] ?? '');
            $spEn = trim($grouped[0][1] ?? '');
            // 获取商品
            $product = $productModel->where("JSON_EXTRACT(product_name, '$.1') = '$spTh'")->find();
            if (!$product) {
                $product = $productModel->where("JSON_EXTRACT(product_name, '$.2') = '$spEn'")->find();
            }
            if (!$product) {
                $notData[] = $grouped;
                continue; // 跳出本次循环
            } 
            // 
            if (Cache::get('tmp-2024-10-01-v4-' . $product->product_id) ) {
                continue; // 已经成功的 - 跳出本次循环
            }
            // 
            foreach ($grouped as $key => $item) {
                $plTh = $item[4];
                $plEn = $item[5];
                $plMy = $item[6];
                $plZh = $item[7];
                // 翻译 
                if (!$plEn || !$plMy || !$plZh) {
                    if (!($langs = Cache::get('tmp-2024-10-01-v3-' . $plTh))) {
                        $res = (new OpenAi)->forward([ "data" => [[
                            'lang' => 'th',
                            'content'=> $plTh,
                        ]]]);
                        if (($res['code'] ?? 0) !== 200) {
                            dump("翻译中断，请重新运行");
                            dump($res);
                            die;
                        }
                        $langs = json_decode($res['data'], true);
                        if (!$langs || !$langs[0]) {
                            dump("翻译中断，请重新运行");
                            dump($res);
                            die;
                        }
                        $langs = $langs[0];
                        // 
                        if (!$langs['en'] || !$langs['my'] || !$langs['zh']) {
                            dump("翻译不全，请重新运行");
                            dump($res);
                            die;
                        }
                        // 
                        Cache::tag('tmp-2024-10-01')->set('tmp-2024-10-01-v3-' . $plTh, $langs);
                    }
                    // 
                    if (!$plEn) {
                        $plEn = $langs['en'];
                    }
                    if (!$plMy) {
                        $plMy = $langs['my'];
                    }
                    if (!$plZh) {
                        $plZh = $langs['zh'];
                    }
                }
                // 
                $pls = json_encode([
                    '1' => $plTh,
                    '2' => $plEn,
                    '3' => $plMy,
                    '4' => $plZh,
                ], JSON_UNESCAPED_UNICODE);
                // 
                $is_exit = false;
                $productFeeds = $product->product_feed;
                foreach($productFeeds as $productFeed) {
                    if ($productFeed['feed_name'] == $pls) {
                        $is_exit = true;
                    }
                }
                if (!$is_exit) {
                    usleep(10000); // 延时10毫秒
                    $productFeeds[] = $feed = [
                        "feed_name" => $pls,
                        "price" => 0 ,
                        "stock_num" => 9999999,
                        "material"=> [],
                        "uuid" => StringHelp::getGuidV4()
                    ];
                    $product->product_feed = $productFeeds;
                    $product->save();
                    // 
                    if (!ProductFeed::where('product_id', $product->product_id)->where('feed_name', $pls)->find()) {
                        $feed['product_id'] = $product->product_id;
                        $feed['app_id'] = $product->app_id;
                        $feed['shop_supplier_id'] = $product->shop_supplier_id;
                        (new ProductFeed)->save($feed);
                    }
                    
                }
            }
            // 
            Cache::tag('tmp-2024-10-01')->set('tmp-2024-10-01-v4-' . $product->product_id, 1);
            dump("该商品成功 - ". $product->product_id .": " . $product->product_name);
        }
        // 找不到的数据
        dump("恭喜你，不用再按回车了，已经全部导入完成");
        die;
    }
}

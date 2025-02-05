<?php

declare(strict_types=1);

namespace app\common\command;

use think\facade\Cache;
use think\console\Input;
use think\console\Output;
use think\console\Command;
use think\console\input\Option;
use app\common\model\settings\Setting;
use app\common\enum\settings\SettingEnum;

// 更新信息
// ./cmd think renewal:info --platform=ttpos
class RenewalInfo extends Command
{
    protected function configure()
    {
        // 指令配置
        $this->setName('renewal:info')
            ->addOption('platform', null, Option::VALUE_REQUIRED, '需要翻译的文本')
            ->addOption('title', null, Option::VALUE_REQUIRED, '设置浏览器的title')
            ->addOption('logo', null, Option::VALUE_REQUIRED, '设置登录页的默认logo')
            ->setDescription('语言翻译');
    }

    protected function execute(Input $input, Output $output)
    {
        $paramPlatform = $input->getOption('platform');
        $paramTitle = $input->getOption('title');
        $paramLogo = $input->getOption('logo');
        //
        $array = [
            'jbc' => 'JBCレジ',
            'tiger' => 'Tiger',
            'TTPOS' => 'TTPOS',
            'ttpos' => 'TTPOS',
        ];
        //
        foreach (['admin', 'shop', 'cashier', 'kitchen', 'tablet', 'assistant'] as $value) {
            //
            $path = public_path($value) . (($value != 'admin' && $value != 'shop') ? 'static/' : '') . 'config.js';
            //
            if (!file_exists($path)) {
                copy(root_path() . 'config.js', $path);
            }
            //
            $fileContent = file_get_contents($path);
            //
            $fileContent = preg_replace("/mode: '.*?'/", "mode: 'saas'", $fileContent);
            $fileContent = preg_replace("/mode:'.*?'/", "mode: 'saas'", $fileContent);
            //
            $fileContent = preg_replace("/saasApiUrl: '.*?'/", "saasApiUrl: ''", $fileContent);
            $fileContent = preg_replace("/saasApiUrl:'.*?'/", "saasApiUrl: ''", $fileContent);
            //
            if (!$paramPlatform) {
                $paramPlatform = Cache::get('RENEWALINFO-PLATFORM');
                if (!$paramPlatform) {
                    $paramPlatform = Setting::where('key', SettingEnum::PLATFORM_BRAND)->value('describe') ?: '';
                }
            }
            //
            if ($paramPlatform || $paramTitle || $paramLogo) {
                if ($paramPlatform) {
                    $fileContent = preg_replace("/brand: '.*?'/", "brand: '$paramPlatform'", $fileContent);
                    $fileContent = preg_replace("/brand:'.*?'/", "brand: '$paramPlatform'", $fileContent);
                    //
                    if ($t = $array[$paramPlatform] ?? '') {
                        $fileContent = preg_replace("/title:'.*?'/", "title: '$t'", $fileContent);
                        $fileContent = preg_replace("/title: '.*?'/", "title: '$t'", $fileContent);
                    }
                    //
                    Cache::set('RENEWALINFO-PLATFORM', $paramPlatform);
                    Setting::where('key', SettingEnum::PLATFORM_BRAND)->findOrEmpty()?->save([
                        'describe' => $paramPlatform,
                        'key' => SettingEnum::PLATFORM_BRAND,
                        'values' => ''
                    ]);
                }
                //
                if ($paramTitle) {
                    $fileContent = preg_replace("/title:'.*?'/", "title: '$paramTitle'", $fileContent);
                    $fileContent = preg_replace("/title: '.*?'/", "title: '$paramTitle'", $fileContent);
                }
                //
                if ($paramLogo) {
                    $fileContent = preg_replace("/webLogo:'.*?'/", "webLogo: '$paramLogo'", $fileContent);
                    $fileContent = preg_replace("/webLogo: '.*?'/", "webLogo: '$paramLogo'", $fileContent);
                }
                //
            }
            file_put_contents($path, $fileContent);
            //
            // $indexPath  = public_path($value) . 'index.html';
            // if (file_exists($indexPath)) {
            //     $fileContent = file_get_contents($indexPath);
            //     $fileContent = str_replace("/config.js", "/config.js". '?t='. time(), $fileContent);
            //     file_put_contents($indexPath, $fileContent);
            // }

            //
            $indexPath = public_path($value) . 'index.html';
            if ($paramPlatform && file_exists($indexPath) ) {
                $fileContent = file_get_contents($indexPath);
                // <title>JBCレジ</title>
                $newTitle = $array[$paramPlatform] ?? '';
                $fileContent = preg_replace(
                    '/<title>[^<]*<\/title>/',
                    '<title>' . $newTitle . '</title>',
                    $fileContent
                );
                if ($value == 'shop') {
                    // <link rel="icon" href="./static/ico/jbc-bc29ac15.ico">
                    if ($paramPlatform == 'jbc') {
                        $paramPlatformIco = $paramPlatform.'-bc29ac15.ico';
                    } else {
                        $paramPlatformIco = 'TTPOS-6aac1dca.ico';
                    }
                    $newIconPath = "./static/ico/{$paramPlatformIco}";
                    $fileContent = preg_replace(
                        '/<link rel="icon" href="\.\/static\/ico\/[^"]+\.ico">/',
                        '<link rel="icon" href="' . $newIconPath . '">',
                        $fileContent
                    );
                }else {
                    // <link rel="shortcut icon" type="image/ico" href="./static/jbc.ico" />
                    $newIconPath = "./static/{$paramPlatform}.ico";
                    $fileContent = preg_replace(
                        '/<link rel="shortcut icon" type="image\/ico" href="\.\/static\/[^"]+\.ico" \/>/',
                        '<link rel="shortcut icon" type="image/ico" href="' . $newIconPath . '" />',
                        $fileContent
                    );
                }
                file_put_contents($indexPath, $fileContent);
            }
        }
        //
        if ($paramPlatform && $paramTitle && $paramLogo) {
            $output->writeln('变更成功');
        }
    }
}

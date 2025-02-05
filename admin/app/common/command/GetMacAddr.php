<?php

declare(strict_types=1);

namespace app\common\command;

use help\LicenseHelp;
use think\console\Input;
use think\console\Output;
use think\console\Command;

// 获取mac 地址
// ./cmd think get-mac-addr

class GetMacAddr extends Command
{
    protected function configure()
    {
        // 指令配置
        $this->setName('get-mac-addr')->setDescription('获取mac地址');
    }

    protected function execute(Input $input, Output $output)
    {
        $res = (new LicenseHelp)->getMacAddress();
        if (!isset($res['mac']) || !$res['mac']) {
            $output->writeln('获取mac地址失败');
        }
    }
}

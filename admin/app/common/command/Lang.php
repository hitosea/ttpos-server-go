<?php

declare(strict_types=1);

namespace app\common\command;

use think\console\Input;
use think\console\Output;
use think\console\Command;
use think\console\input\Option;
use app\common\library\language\Language;

// 语言翻译
// ./cmd think lang
// ./cmd think lang --text=萨达撒萨达撒萨达 --channel=ai
class Lang extends Command
{
    protected function configure()
    {
        // 指令配置
        $this->setName('lang')
            ->addOption('text', null, Option::VALUE_REQUIRED, '需要翻译的文本')
            ->addOption('channel', null, Option::VALUE_REQUIRED, 'ai|google|youdao')
            ->setDescription('语言翻译');
    }

    protected function execute(Input $input, Output $output)
    {
        // 指令输出
        $output->writeln('#####开始提取所有中文#####');
        $channel = $input->getOption('channel') ?: env('TRANSLATORS_CHANNEL', 'ai');

        // 提取所有中文
        $lang = new Language();
        $paramText = $input->getOption('text');
        $texts = $lang->extractTexts($paramText ? [$paramText] : []);

        // 执行待翻译中文
        $output->writeln('#####开始翻译#####');

        // 新语言先执行翻译
        $newTargets = $lang->getTranslatedLang();
        if (!empty($newTargets)) {
            $lang->commandExecute($newTargets, $lang->getTranslatedTexts(), $channel);
        }
        // 语言翻译
        $targets = $lang->getTargets($channel);
        $output->writeln("count: " . count($texts));
        $lang->commandExecute($targets, $texts, $channel);
        $output->writeln('#####翻译完成#####');
    }
}

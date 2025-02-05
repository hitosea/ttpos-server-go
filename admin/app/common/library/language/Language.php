<?php

namespace app\common\library\language;

use think\facade\Filesystem;
use app\common\library\language\engine\Google;
use app\common\library\language\engine\OpenAi;
use app\common\library\language\engine\YouDao;

/**
 * 多语言管理类
 */
class Language
{
    // 当前文件语言集
    // zh: 中文 | zhtw: 繁体中文 | en: 英文 | th: 泰文 | ja: 日文 | tr: 土耳其文 | ko：韩语 | my: 缅甸语
    private $targetFiles = ['zh', 'zhtw', 'en', 'th', 'ja', 'tr', 'ko', 'my'];

    // 谷歌语言集
    private $googleTargetFiles = ['zh', 'zh-TW', 'en', 'th', 'ja', 'tr', 'ko', 'my'];

    // 有道语言集
    private $youdaoTargetFiles = ['zh-CHS', 'zh-CHT', 'en', 'th', 'ja', 'tr', 'ko', 'my'];

    // Ai语言集
    private $aiTargetFiles = ['zh', 'zh-TW', 'en', 'th', 'ja', 'tr', 'ko', 'my'];

    // 当前语言文件路径
    private $langFilePath = 'lang/';

    // 有道翻译参数
    private $youdaoAppKey;
    private $youdaoSecKey;


    public function __construct()
    {
        $this->youdaoAppKey = env('YOUDAO_APP_KEY') ?? '';
        $this->youdaoSecKey = env('YOUDAO_SEC_KEY') ?? '';
    }

    /**
     * 语言提取
     *
     * @param array $texts
     * @return array
     */
    public function extractTexts($texts = [])
    {
        if (!empty($texts)) {
            return $texts;
        }
        $arr = [];
        $files = self::getDirAllFile(app_path());
        foreach ($files as $file) {
            $contents = file_get_contents($file);
            $contents = preg_replace('/\/\*\*.*\n(.*\*.*\n)+/', '', $contents); // 去掉注释
            preg_match_all("/(\'|\")((auto\.)?(([a-zA-Z0-9\/\-]+)?[\x{4e00}-\x{9fa5}]+.*?))(\'|\")/u", $contents, $match); // 匹配中文
            $arr = array_merge($arr, $match[4] ?? []);
        }
        $texts = array_values(array_filter(array_unique($arr)));
        return $texts;
    }

    /**
     * 获取已翻译中文
     *
     * @return array
     */
    public function getTranslatedTexts()
    {
        $translated = !is_file('lang/zh/auto.php') ? [] : include('lang/zh/auto.php');
        return $translated;
    }

    /**
     * 过滤已翻译中文
     *
     * @param array $texts
     * @param array $translated
     * @return array
     */
    public function filterTexts($texts, $translated)
    {
        $texts = array_filter($texts, function ($text) use ($translated) {
            return empty(($translated[$text] ?? '')) && !str_contains($text, '$') && !str_contains($text, '","');
        });
        return $texts;
    }

    /**
     * 获取目标语言
     *
     * @param string $channel
     * @return array
     */
    public function getTargets($channel)
    {
        $targets = [];
        if ($channel == 'google') {
            $targets = array_combine($this->targetFiles, $this->googleTargetFiles);
        } else if ($channel == 'youdao') {
            $targets = array_combine($this->targetFiles, $this->youdaoTargetFiles);
        } else if ($channel == 'ai') {
            $targets = array_combine($this->targetFiles, $this->aiTargetFiles);
        }
        return $targets;
    }

    /**
     * 获取翻译引擎
     *
     * @param string $channel
     * @return Google|YouDao|null
     */
    public function getTranslator($channel)
    {
        if ($channel == 'google') {
            $tr = new Google(); // Translates to 'en' from auto-detected language by default
        } else if ($channel == 'youdao') {
            $tr = new YouDao($this->youdaoAppKey, $this->youdaoSecKey);
        } else if ($channel == 'ai') {
            $tr = new OpenAi();
        } else {
            return;
        }
        return $tr;
    }

    /**
     * 获取保存路径
     *
     * @param string $file
     * @return string
     */
    public function getSavePath($file)
    {
        return $this->langFilePath . $file . '/auto.php';
    }

    /**
     * 初始化保存路径
     *
     * @param string $file
     * @return string
     */
    public function initSavePath($file)
    {
        $savePath = $this->getSavePath($file);
        $this->makeDir($this->langFilePath . $file);
        if (!is_file($savePath)) {
            $emptyContent = <<<EOF
                <?php
                return [];
                EOF;
            file_put_contents($savePath, $emptyContent);
        }
        return $savePath;
    }

    /**
     * 获取新翻译语言
     */
    public function getTranslatedLang()
    {
        $newTargets = [];
        foreach ($this->targetFiles as $target) {
            $savePath = $this->getSavePath($target);
            if (!file_exists($savePath)) {
                $newTargets[$target] = $target;
            }
        }
        return empty($newTargets) ? [] : array_merge($newTargets, ['zh' => 'zh']);
    }

    /**
     * 执行翻译
     *
     * @param array $targets
     * @param array $texts
     * @param [type] $tr
     * @return void
     */
    public function commandExecute($targets = [], $texts = [], $channel = 'google')
    {
        $tr = $this->getTranslator($channel); // 获取翻译器
        if (empty($targets) || empty($texts) || empty($tr)) {
            return;
        }
        // ai翻译
        if ($channel == 'ai') {
            $this->aiTranslate($targets, $texts, $tr);
            return;
        }
        // 常规翻译
        foreach ($targets as $file => $target) {
            $savePath = $this->initSavePath($file); // 不同语言保存路径
            $translation = include($savePath); // 不同语言已翻译信息
            $newTexts = $this->filterTexts($texts, $translation); // 过滤已翻译词
            $textChunks = array_chunk(array_values($newTexts), 200); // 新词分块
            foreach ($textChunks as $chunk) {
                if ($file == 'zh') {
                    // 简体中文不用翻译
                    foreach ($chunk as $item) {
                        $translation[$item] = $item;
                    }
                } else {
                    // 其他语言翻译
                    $q = implode("\n", $chunk);
                    $trans = $tr->translate($q, $targets['zh'], $target);
                    $trans = is_string($trans) ? $trans : '';
                    $translation = array_merge($translation, array_combine($chunk, explode("\n", $trans)));
                }
            }
            // 保存翻译结果
            $translationContent = '';
            foreach ($translation as $k => $v) {
                // 保持原有的换行符位置
                $v = str_replace("\n", "\\n", $v);
                $translationContent .= sprintf('    "%s" => "%s",' . PHP_EOL, $k, $v);
            }
            $filledContent = <<<EOF
            <?php
            return [
            $translationContent];
            EOF;
            file_put_contents($savePath, $filledContent);
        }
    }

    /**
     * ai翻译
     *
     * @param array $targets
     * @param array $texts
     * @param [type] $tr
     * @return void
     */
    public function aiTranslate($targets = [], $texts = [], $tr = null, $from = 'cn')
    {
        if (empty($targets) || empty($texts) || empty($tr)) {
            return;
        }

        $newTexts = [];
        $langTexts = [];
        $translationsByLang = [];
        $existingTranslations = [];
        foreach ($targets as $lang => $target) {
            $savePath = $this->initSavePath($lang); // 不同语言保存路径
            $existingTranslations[$lang] = include($savePath); // 不同语言已翻译信息
            $newTexts = $this->filterTexts($texts, $existingTranslations[$lang]); // 过滤已翻译词
            // 给新词每个元素加上语言标识
            foreach ($newTexts as $text) {
                $langTexts[$target][] = $text;
            }
        }

        $uniqueLangTexts = array_values(array_unique(array_merge(...array_values($langTexts))));
        $textChunks = array_chunk($uniqueLangTexts, 10); // 新词分块，一次请求多少个词
        foreach ($textChunks as $chunk) {
            $trans = $tr->translate($chunk, $from); // 翻译
            // 解析翻译结果
            foreach ($trans as $translation) {
                if ((count($translation) - 1) != count($this->targetFiles)) {
                    throw new \Exception("ai translate return error");
                }
                foreach ($targets as $lang => $target) {
                    if (isset($translation[$target])) {
                        $translationsByLang[$lang][$translation['key']] = $translation[$target];
                    }
                }
            }
        }

        // 更新语言文件
        foreach ($translationsByLang as $lang => $translations) {
            $savePath = $this->initSavePath($lang);
            $mergedTranslations = array_merge($existingTranslations[$lang], $translations);
            // 保存翻译结果
            $translationContent = '';
            foreach ($mergedTranslations as $k => $v) {
                // 保持原有的换行符位置
                $v = str_replace("\n", "\\n", $v);
                $translationContent .= sprintf('    "%s" => "%s",' . PHP_EOL, $k, $v);
            }
            $filledContent = <<<EOF
            <?php
            return [
            $translationContent];
            EOF;
            file_put_contents($savePath, $filledContent);
        }
    }

    /**
     * 新建目录
     * @param $path
     * @return mixed
     */
    public function makeDir($path)
    {
        try {
            Filesystem::createDirectory($path);
        } catch (\Exception $e) {
        }
        if (!file_exists($path)) {
            $this->makeDir(dirname($path));
            @mkdir($path, 0777);
            @chmod($path, 0777);
        }
        return $path;
    }

    /**
     * 获取目录下所有文件
     * @param $path
     * @param array $arr
     * @return array
     */
    public function getDirAllFile($path, $arr = [])
    {
        $arr = array();
        if (is_file($path)) {
            if (str_ends_with($path, '.php')) {
                $arr[] = $path;
            }
        } else {
            if (is_dir($path)) {
                $data = scandir($path);
                if (!empty($data)) {
                    foreach ($data as $value) {
                        if ($value != '.' && $value != '..') {
                            $sub_path = $path . "/" . $value;
                            $temp = self::getDirAllFile($sub_path);
                            $arr = array_merge($temp, $arr);
                        }
                    }
                }
            }
        }
        return $arr;
    }
}

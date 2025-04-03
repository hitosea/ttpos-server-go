<?php

/**
 * 字体生成图像
 * @wfs: 2024/04/19
 */

namespace base\imgs;

use Endroid\QrCode\QrCode;
use Endroid\QrCode\Writer\PngWriter;


class ImgFont
{
    // 图像对象
    protected $image;

    // 方向
    protected $direction = 0;

    // 图片宽度
    protected $imageWidth = 567;

    // 图片高度
    protected $imageHeight = 0;

    // 图片内边距
    protected $imagePadding = 24;

    // 文本行高
    protected $defaultTextLineHeight = 45;
    protected $textLineHeight = 45;

    // 文本间距
    protected $textSpacing = 1;

    // 文本总高度
    protected $textTotalHeight = 0;

    // 文本最后一行已使用的宽度
    protected $textLastLineUsedWidth = 0;

    // 文本对齐方向
    protected $alignment = 1;

    // 字体大小
    protected $fontSize = 20;

    // 字体粗细
    protected $fontWeight = 1;

    // 缅甸语的特殊字体
    protected $mySpecialFonts = [];

    // 字体
    const FONT_EN = "/fonts/en/NotoSans-Regular.ttf";
    const FONT_JA = "/fonts/ja/NotoSansJP-Regular.ttf";
    const FONT_KO = "/fonts/ko/NotoSansKR-Regular.ttf";
    const FONT_MY = "/fonts/my/NotoSansMyanmar-Regular.ttf";
    const FONT_MY2 = "/fonts/my/Zawgyi-One_v4.ttf";
    const FONT_TH = "/fonts/th/NotoSansThai-Regular.ttf";
    const FONT_TR = "/fonts/tr/NotoSans-Regular.ttf";
    const FONT_ZH = "/fonts/zh/NotoSansSC-Regular.ttf";
    const FONT_ALL = "/fonts/NotoSans-Regular-SC_KR_TH.ttf";

    // 图片地址
    const IMAGE_SRC = "";

    // 横向排列
    const DIRECTION_X = 0;

    // 竖向排列
    const DIRECTION_Y = 1;

    // 文本对齐方向
    const ALIGN_LEFT = 1;
    const ALIGN_CENTER = 2;
    const ALIGN_RIGHT = 3;

    /**
     * 架构方法 设置参数
     * @access public
     * @param int  $imageWidth | height
     * @param int $direction 方向 
     */
    public function __construct(int $imageWidth = 0, $defaultTextLineHeight = 0, $direction = self::DIRECTION_X)
    {
        // 
        if ($imageWidth) {
            $this->imageWidth = $imageWidth;
        }
        if ($defaultTextLineHeight) {
            $this->defaultTextLineHeight = $defaultTextLineHeight;
            $this->textLineHeight = $defaultTextLineHeight;
        }
        // 
        if ($direction == self::DIRECTION_Y) {
            $this->direction = 90;
            $this->imageHeight = $imageWidth;
        }
        // 
        $this->mySpecialFonts = [
            'ကြို' => Rabbit::uni2zg('ကြို'), 
            'က္တြ' => Rabbit::uni2zg('က္တြ'), 
            'ကြွ' => Rabbit::uni2zg('ကြွ'),
            'ပြု' => Rabbit::uni2zg('ပြု'),
            'ကြော' => Rabbit::uni2zg('ကြော'),
            'မှု' => Rabbit::uni2zg('မှု'),
            'က္ခ' => Rabbit::uni2zg('က္ခ'),
            'ဏ္ဍ​' => Rabbit::uni2zg('ဏ္ဍ​'),
            'ဒ္ဒ​' => Rabbit::uni2zg('ဒ္ဒ'),
            'န္တ' => Rabbit::uni2zg('န္တ​'),
            'န္အ' => Rabbit::uni2zg('န္အ'),
        ];
        // 
        $this->createImg();
    }

    /**
     * 创建img
     */
    private function createImg()
    {
        // 创建一个空白画布
        $this->image = imagecreatetruecolor($this->imageWidth, 20000);
        // 设置背景颜色为白色
        imagefill($this->image, 0, 0, imagecolorallocate($this->image, 255, 255, 255));
    }

    /**
     * 获取字体文件地址
     * @access private
     * @param string $text
     * @return string
     */
    private function getFontPath($char)
    {
        if (preg_match('/[\p{Thai}]/u', $char)  || strpos($char, "฿") !== false) {
            return dirname(__FILE__) . self::FONT_TH;
        } 
        else if (preg_match('/[\p{Hangul}]/u', $char)) {
            return dirname(__FILE__) . self::FONT_KO;
        } 
        else if (preg_match('/[\x{4E00}-\x{9FFF}\x{FF01}\x{FF0C}\x{FF08}\x{FF09}\x{FF1A}\x{FF5E}\x{2014}]/u', $char) || $char == '￥' || $char == '；' || $char == '？' || $char == '＋') {
            return dirname(__FILE__) . self::FONT_ZH;
        } 
        else if (preg_match('/[\p{Hiragana}\p{Katakana}\p{Han}★]+/u', $char)) {
            return dirname(__FILE__) . self::FONT_JA;
        } 
        else if (preg_match('/[\x{1000}-\x{109F}\x{AA60}-\x{AA7F}\x{A9E0}-\x{A9FF}\x{AA20}-\x{AA3F}]/u', $char)) {
            if ( in_array($char, array_values($this->mySpecialFonts))) {
                return dirname(__FILE__) . self::FONT_MY2;
            }
            return dirname(__FILE__) . self::FONT_MY;
        } 
        else if (preg_match('/[\x{011E}-\x{0130}\x{0131}\x{015E}-\x{015F}\x{00C7}-\x{00E7}\x{011F}]/u', $char)) {
            return dirname(__FILE__) . self::FONT_TR;
        } 
        else {
            return dirname(__FILE__) . self::FONT_EN;
        }
    }

    /**
     * 获取字体宽度
     * @access private
     * @param string $text
     * @return string
     */
    private function getFontWeight($fontSize, $char)
    {
        $fontPath = $this->getFontPath($char);
        if ($char =='ြ') {
            $charWidth = 5;
        } else {
            $charWidth = imagettfbbox($fontSize, 0, $fontPath, $char)[2];
        }
        if ($fontPath == dirname(__FILE__) . self::FONT_MY || $fontPath == dirname(__FILE__) . self::FONT_MY2) {
            if ($charWidth < 0) {
                $charWidth = 0;
            }
        }
        return $charWidth;
    }

    /**
     * 获取字体数组
     * @access private
     * @return array
     */
    private function getFontArrays($texts)
    {
        $segments = [];
        $texts = preg_split('/(' . implode('|', array_keys($this->mySpecialFonts)) . ')/u', $texts, -1, PREG_SPLIT_DELIM_CAPTURE | PREG_SPLIT_NO_EMPTY);
        foreach ($texts as $text) {
            if (in_array($text, array_keys($this->mySpecialFonts))) {
                $segments[] = $this->mySpecialFonts[$text];
            } else {
                $pattern = '/(?!' . implode('|', ['မှု']) . ')/u';
                $segment = preg_split($pattern, $text, -1, PREG_SPLIT_DELIM_CAPTURE | PREG_SPLIT_NO_EMPTY);
                // 调整缅甸语的书写顺序
                for ($i = 1; $i < count($segment); $i++) {
                    if ($segment[$i] === 'ြ') {
                        $temp = $segment[$i - 1] ?? '';
                        $segment[$i - 1] = $segment[$i];
                        $segment[$i] = $temp;
                    }
                }
                for ($i = 1; $i < count($segment); $i++) {
                    if ($segment[$i] === 'ေ') {
                        $temp = $segment[$i - 2] ?? '';
                        if ($temp == 'ြ') {
                            $segment[$i - 2] = $segment[$i];
                            $segment[$i] = $temp;
                        } 
                        $temp = $segment[$i - 1];
                        $segment[$i - 1] = $segment[$i];
                        $segment[$i] = $temp;
                    }
                }
                $segments = array_merge($segments, $segment);
            }
        }
        // 
        return $segments;
    }

    /**
     * 添加文本
     * @access public
     * @param string $text
     * @return bool
     */
    private function addText(string $text, float $height, float $fixedWidth = 0, float $deviationWidth = 0): array
    {
        // 字体大小
        $fontSize = $this->fontSize;
        // 字体粗
        $fontWeight = $this->fontWeight;
        // 设置文本颜色为黑色
        $black = imagecolorallocate($this->image, 0, 0, 0);
        // 分隔字体为数组
        $segments = $this->getFontArrays($text);
        // 获取已使用高度  1.居右并固定宽度的计算 ，2 正常 
        if ($fixedWidth && $this->alignment == self::ALIGN_RIGHT) {
            $useWidth = $this->imageWidth - $fixedWidth - $deviationWidth - $this->imagePadding;
            $canWidth = $this->imageWidth - $useWidth - $this->imagePadding * 3;
            [$tmpw, $tmpStr] = [0, ''];
            foreach ($segments as $key => $char) {
                $tmpw = $this->getFontWeight($fontSize, $tmpStr .= $char) * $this->textSpacing;
                if ($tmpw > $canWidth) {
                    break;
                }
            }
            // 
            $useWidth = $this->imageWidth - $fixedWidth - $deviationWidth - $this->imagePadding - ($tmpw + $this->imagePadding);
        } else {
            $useWidth = $this->textLastLineUsedWidth + $deviationWidth;
        }
        // 回归宽度
        $this->textLastLineUsedWidth = 0;
        // 
        $isDeviation = $deviationWidth;
        // 执行
        foreach ($segments as $key => $char) {
            // 记录
            $oldUseWidth = $useWidth;
            // 换行
            if ($char == "\n") {
                $useWidth = $deviationWidth;
                // 剧右并固定宽度的计算
                if ($fixedWidth && $this->alignment == self::ALIGN_RIGHT) {
                    $useWidth = $this->imageWidth - $fixedWidth - $this->imagePadding;
                }
                // 
                $height += $this->textLineHeight;
                continue;
            }
            // 当前字体的宽度
            $charWidth = $this->getFontWeight($fontSize, $char) * $this->textSpacing;
            // 文本排列方向
            if ($this->alignment != self::ALIGN_LEFT) {
                $lastText = array_slice($segments, $key);
                // 最后一行的宽度
                $lastTextWidth = 0;
                foreach ($lastText as $c) {
                    $lastTextWidth += $this->getFontWeight($fontSize, $c);
                }
                $lastTextWidth = $lastTextWidth * $this->textSpacing;
                // 文本居中
                if (($useWidth == 0 || $isDeviation) && $this->alignment == self::ALIGN_CENTER) {
                    $oldUseWidth = $useWidth = ($this->imageWidth + $deviationWidth - $lastTextWidth - $this->imagePadding * 2) / 2;
                    $isDeviation = 0;
                }
                // 文本居右
                // 最后一行能够正常显示的宽度计算
                if ($this->alignment == self::ALIGN_RIGHT) {
                    $calculateWidth = $this->imageWidth - $useWidth - $lastTextWidth;
                    if ($calculateWidth > $this->imagePadding * 2) {
                        $oldUseWidth = ($this->imageWidth - $lastTextWidth - $this->imagePadding * 2) - 2;
                    }
                }
                // 小于就从0开始
                if ($useWidth <= 0) $useWidth = 0;
                if ($oldUseWidth <= 0) $oldUseWidth = 0;
            }
            // 累加宽度
            $useWidth += $charWidth;
            $isLinebreak = false;
            // 正常计算
            if ($useWidth >= ($this->imageWidth - $this->imagePadding * 2)) {
                $isLinebreak = true;
            }
            // 居左并固定宽度的计算
            if ($fixedWidth && $this->alignment == self::ALIGN_LEFT) {
                if ($useWidth - $deviationWidth >= $fixedWidth + $this->imagePadding * 2) {
                    $isLinebreak = true;
                }
            }
            // 输入或换行
            if (!$isLinebreak) {
                $fontPath = $this->getFontPath($char);
                for ($i = 1; $i <= $fontWeight; $i++) {
                    imagettftext($this->image, $fontSize, 0, (int)($oldUseWidth + $this->imagePadding), (int)$height, $black, $fontPath, $char);
                }
            } else {
                $newArray = array_slice($segments, $key);
                $mySpecialFonts = array_flip($this->mySpecialFonts);
                foreach ($newArray as &$c) {
                    if (isset($mySpecialFonts[$c])) {
                        $c = $mySpecialFonts[$c];
                    }
                }
                return $this->addText(implode('', $newArray), $height + $this->textLineHeight, $fixedWidth, $deviationWidth);
            }
        }
        // 
        return [
            "height" => $height,
            "width" => $useWidth,
        ];
    }

    /**
     * 添加文本
     * @access public
     * @param string $text
     * @return bool
     */
    public function appendPartingline(string $text, int $fixedWidth = 0, $deviationWidth = 0): self
    {
        $result = $this->addText($text, $this->textTotalHeight ?: $this->textLineHeight, $fixedWidth, $deviationWidth);
        // 
        $this->textTotalHeight = $result['height'];
        // 
        $this->textLastLineUsedWidth = $result['width'] + 1;
        //
        return $this;
    }

    /**
     * 添加文本
     * @access public
     * @param string $text
     * @return bool
     */
    public function appendText(string $text, int $fixedWidth = 0, $deviationWidth = 0): self
    {
        $texts = explode("\n", $text);
        foreach ($texts as $key => $value) {
            if ($value) {
                $result = $this->addText($value, $this->textTotalHeight ?: $this->textLineHeight, $fixedWidth, $deviationWidth);
                // 
                $this->textTotalHeight = $result['height'];
                // 
                $this->textLastLineUsedWidth = $result['width'] + 1;
            }
            // 
            if (count($texts) - 1 != $key) {
                $this->lineFeed(1);
            }
        }
        //
        return $this;
    }

    /**
     * 添加图片
     * @access public
     * @param string $text
     * @return bool
     */
    public function appendImg(string $imgPath, int $size = 300, $isRoundness = false, $topHeight = 0): self
    {
        $imageData = file_get_contents($imgPath);
        $image = imagecreatefromstring($imageData);
        $width = imagesx($image);
        $height = imagesy($image);
        // 调整图片大小
        if ($size > 0) {
            $ratio = $width / $height;
            $width = $size;
            $height = (int)($size / $ratio);
            $resizedImage = imagecreatetruecolor($width, $height);
            imagecopyresampled($resizedImage, $image, 0, 0, 0, 0, $width, $height, imagesx($image), imagesy($image));
            imagedestroy($image);
            $image = $resizedImage;
        }
        // 圆角
        if ($isRoundness) {
            $mask = imagecreatetruecolor($width, $height);
            $maskBg = imagecolorallocate($mask, 255, 255, 255);
            imagefill($mask, 0, 0, $maskBg);
            $maskFg = imagecolorallocate($mask, 0, 0, 0);
            imagefilledellipse($mask, $width / 2, $height / 2, $width, $height, $maskFg);
            imagecolortransparent($mask, $maskFg);
            imagecopymerge($image, $mask, 0, 0, 0, 0, $width, $height, 100);
            imagedestroy($mask);
        }
        // 计算图片位置
        $x = ($this->imageWidth - $width) / 2;
        imagecopy($this->image, $image, $x, $this->textTotalHeight + $topHeight, 0, 0, $width, $height);
        imagedestroy($image);
        // 更新文本总高度和最后一行已使用宽度
        $this->textTotalHeight += $height + $topHeight;
        $this->textLastLineUsedWidth += $width;
        // 换行
        $this->lineFeed(1);
        // 
        return $this;
    }

    /**
     * 添加二维码
     * @access public
     * @param string $rootPath
     * @param int $fixedWidth
     * @param int $deviationWidth
     * @return self
     */
    public function appendQrcode(string $data, int $size = 280, $margin = 10, $isBase64 = false): self
    {
        // 生成二维码图片
        if ($isBase64 && ($data = base64_decode($data, true))) {
            $qrCodeString = $data;
            $this->textTotalHeight = $this->textTotalHeight - ($margin * 2);
        } else {
            $qrCode = new QrCode($data);
            $qrCodeString = (new PngWriter())->write($qrCode)->getString();
        }
        // 
        $image = imagecreatefromstring($qrCodeString);
        $width = imagesx($image);
        $height = imagesy($image);
        // 调整图片大小
        if ($size > 0) {
            $ratio = $width / $height;
            $width = $size;
            $height = (int)($size / $ratio);
            $resizedImage = imagecreatetruecolor($width, $height);
            imagecopyresampled($resizedImage, $image, 0, 0, 0, 0, $width, $height, imagesx($image), imagesy($image));
            imagedestroy($image);
            $image = $resizedImage;
        }
        // 计算二维码位置
        $x = ($this->imageWidth - $width) / 2;
        imagecopy($this->image, $image, $x, $this->textTotalHeight, 0, 0, $width, $height);
        imagedestroy($image);
        // 更新文本总高度和最后一行已使用宽度
        $this->textTotalHeight += $height;
        $this->textLastLineUsedWidth += $width;
        // 换行
        $this->lineFeed(1);
        // 
        return $this;
    }

    /**
     * 添加分割行
     * @access public
     * @param string $text
     * @return bool
     */
    public function appendSplitline($islineFeed = false, $lineHeight=0, $fW = 2): self
    {
        $fontWeight = $this->fontWeight;
        // 
        $this->setFontWeight($fW);
        if ($this->imageWidth == 567) {
            $this->appendText("----------------------------------------------------------------------");
        }
        if ($this->imageWidth == 568) {
            $this->appendText("----------------------------------------------------------------------");
        }
        if ($islineFeed) {
            $this->lineFeed(1, $lineHeight);
        } 
        // 
        $this->setFontWeight($fontWeight);
        //
        return $this;
    }

    /**
     * 设置列
     * @access public
     * @param string  printInColumns(["商品", 320], ["数量", 0], ["金额", 100])
     * @return bool
     */
    public function setupColumns(): self
    { 
        $params = [];
        for ($i = 0; $i < func_num_args(); $i++) {
            $params[] = func_get_arg($i);
        }


        return $this;
    }

    /**
     * 添加列
     * @access public
     * @param string  printInColumns(["商品", $width, $align,  $fontWeight, $fontSize])
     * @return bool
     */
    public function printInColumns(): self
    {
        $params = [];
        for ($i = 0; $i < func_num_args(); $i++) {
            $params[] = func_get_arg($i);
        }
        // 
        $imageWidth = $this->imageWidth;
        $height = $this->textTotalHeight ?: $this->textLineHeight;
        $allColumnWidth = array_sum(array_column($params, '1'));
        $deviationWidth = 0;
        $oldFontWeight = $this->fontWeight;
        $oldFontSize = $this->fontSize;
        $results = [];
        // 
        $lineHeight = 0;
        foreach ($params as $key => $strrem) {
            $text = $strrem[0];
            $columnWidth = $strrem[1];
            if ($columnWidth == 0) {
                $columnWidth = $imageWidth - $allColumnWidth;
            }
            $align = $strrem[2] ?? '';
            $fontWeight = $strrem[3] ?? 0;
            $fontSize = $strrem[4] ?? 0;
            $lineHeight = $strrem[5] ?? 0;
            // 
            if ($align == self::ALIGN_CENTER || ($align == self::ALIGN_RIGHT && $key != count($params) - 1)) {
                $this->imageWidth = $deviationWidth + $columnWidth + $this->imagePadding;
            }
            if ($fontWeight) $this->setFontWeight($fontWeight);
            if ($fontSize) $this->setFontSize($fontSize);
            // 
            $this->setAlignment($strrem[2] ?? self::ALIGN_LEFT);
            $results[] = $this->addText($text, $height, $columnWidth - $this->imagePadding * 2, $deviationWidth);
            //    
            $deviationWidth += $columnWidth;
            $this->imageWidth = $imageWidth;
            if ($fontWeight) $this->setFontWeight($oldFontWeight);
            if ($fontSize) $this->setFontSize($oldFontSize);
        }
        // 
        $this->textTotalHeight = max(array_column($results, 'height'));
        // 
        $this->textLastLineUsedWidth = 0;
        //
        $this->setAlignment(self::ALIGN_LEFT);
        $this->lineFeed(1, $lineHeight);
        // 
        return $this;
    }

    /**
     * 设置文本行高
     * @access public
     * @param string $text
     * @return bool
     */
    public function setTextLineHeight(int $textLineHeight = 0): self
    {
        if ($textLineHeight > 0) {
            $this->textLineHeight = $textLineHeight;
        }
        return $this;
    }

    /**
     * 设置文本行高
     * @access public
     * @param string $text
     * @return bool
     */
    public function recoverDefaultTextLineHeight(): self
    {
        $this->textLineHeight = $this->defaultTextLineHeight;
        return $this;
    }

    /**
     * 设置对齐方式
     * @access public
     * @param string $text
     * @return bool
     */
    public function setAlignment(int $alignment = self::ALIGN_LEFT): self
    {
        $this->alignment = $alignment;
        return $this;
    }

    /**
     * 换行
     * @access public
     * @param int $num
     * @return bool
     */
    public function lineFeed(int $num = 1, $lineHeight = 0): self
    {
        $this->textLastLineUsedWidth = 0;
        $this->textTotalHeight += $lineHeight ?: ($this->textLineHeight * $num);
        return $this;
    }

    /**
     * setFontSize
     * @access public
     * @param int $num
     * @return bool
     */
    public function setFontSize(int $fontSize = 0): self
    {
        if ($fontSize > 0) {
            $this->fontSize = $fontSize;
        }
        return $this;
    }

    /**
     * setFontWeight
     * @access public
     * @param int $num
     * @return bool
     */
    public function setFontWeight(int $num = 1): self
    {
        if ($num >= 1) {
            $this->fontWeight = $num;
        }
        return $this;
    }

    /**
     * 设置文本间距
     * @access public
     * @param int $num
     * @return bool
     */
    public function setTextSpacing(float $num = 1): self
    {
        $this->textSpacing = $num;
        return $this;
    }

    /**
     * 恢复默认值
     * @access public
     * @param int $num
     * @return bool
     */
    public function restoreDefault(): self
    {
        $this->recoverDefaultTextLineHeight();
        $this->setFontWeight(1);
        $this->setFontSize(20);
        $this->setTextSpacing(1);
        $this->setAlignment(ImgFont::ALIGN_LEFT);
        return $this;
    }

    /**
     * 设置图像边距
     * @access public
     * @param string $text
     * @return bool
     */
    public function setImagePadding(int $num = 0): self
    {
        $this->imagePadding = $num;
        return $this;
    }

    /**
     * 保存
     * @access public
     * @param string $imageSrc
     * @return string
     */
    public function save(string $imageSrc = self::IMAGE_SRC, bool $reminderSound = true, bool $openMoneybox = false): string
    {
        $data = [];
        $maxHeight = 2200;
        $height = $this->textTotalHeight + $this->textLineHeight;
        $headHeight =(int)($this->textLineHeight / 2) - 10;
        // 
        $heights = [];
        while ($height > $maxHeight) {
            $heights[] = $maxHeight;
            $height -= $maxHeight;
        }
        $heights[] = $height;
        // 
        foreach ($heights as $key => $height) {
            $height = (int)$height;
            $cropped_image = imagecreatetruecolor($this->imageWidth + ($this->direction != 0 ? 180 : 0), $height);
            imagefill($cropped_image, 0, 0, imagecolorallocate($cropped_image, 255, 255, 255));
            // 
            if ($key == 0) {
                imagecopy($cropped_image, $this->image, 0, 0, 0, $headHeight, $this->imageWidth, $height);
            } else {
                imagecopy($cropped_image, $this->image, 0, 0, 0, $maxHeight * $key + $headHeight, $this->imageWidth, $height);
            }
            // 
            $rotated_image = imagerotate($cropped_image, -$this->direction, 0, true);
            // 输出图片
            if ($imageSrc) {
                $directory = dirname($imageSrc);
                if (!is_dir($directory)) {
                    mkdir($directory, 0777, true);
                }
                imagepng($rotated_image, $imageSrc);
            }
            // 释放内存
            imagedestroy($cropped_image);
            imagedestroy($rotated_image);
            // 二唯码图片打印数据
            $data[] = chr(29) . chr(118) . chr(48) . chr(0) . $this->getBytesFromBitMap($rotated_image);
        }
        // 
        $print_code = implode('', $data);
        // 提示音
        if ($reminderSound) {
            $print_code .= "\x1B\x42\x03\x02";
        }
        // 切纸
        $print_code .= "\x1d\x56\x00";
        // 打开钱箱
        if ($openMoneybox === 2) {
            $print_code .= chr(27) . chr(112) . chr(0) . chr(25) . chr(250);
        }else if ($openMoneybox) {
            $print_code .=  "\x10\x14\x01\x00\x01";
        }
        // 
        return bin2hex($print_code);
    }

    // 转 print_code
    public function toRasterFormat($image){
        $print_code = chr(29) . chr(118) . chr(48) . chr(0);
        $print_code .= $this->getBytesFromBitMap($image);
        $print_code .= "\x1d\x56\x00";
        return bin2hex($print_code);
    }

    /**
     * 将bitmap图转换为头四位有宽高的光栅位图
     */
    public static function getBytesFromBitMap($bitmap)
    {
        // 获取图像的宽度和高度
        $width = (int)imagesx($bitmap);
        $height = (int)imagesy($bitmap);
        $bw = (int)(($width - 1) / 8) + 1;

        // 初始化返回的字节数组
        $rv = array_fill(0, $height * $bw + 4, 0);
        // xL
        $rv[0] = $bw & 0xFF;
        // xH
        $rv[1] = ($bw >> 8) & 0xFF;
        $rv[2] = $height & 0xFF;
        $rv[3] = ($height >> 8) & 0xFF;

        // 获取图像的像素数据
        for ($i = 0; $i < $height; $i++) {
            for ($j = 0; $j < $width; $j++) {
                $clr = imagecolorat($bitmap, $j, $i);
                $red = ($clr >> 16) & 0xFF;
                $green = ($clr >> 8) & 0xFF;
                $blue = $clr & 0xFF;
                $gray = self::RGB2Gray($red, $green, $blue);
                $rv[(int)(($width * $i + $j) / 8) + 4] |= ($gray << (7 - (($width * $i + $j) % 8)));
            }
        }

        // 释放内存
        imagedestroy($bitmap);

        // 将数组转换为字节字符串
        return implode(array_map("chr", $rv));
    }

    /**
     * 将 RGB 转换为灰度值
     */
    private static function RGB2Gray($red, $green, $blue)
    {
        return ($red * 0.29900 + $green * 0.58700 + $blue * 0.11400) > 127 ? 0 : 1;
    }
}

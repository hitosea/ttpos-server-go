<?php

namespace help;

class PrintTexHelp
{
    /**
     * 获取文本宽度
     * @return int
     */
    public static function getTextWidth($text)
    {
        if (preg_match('/[\p{Thai}฿]/u', $text)) { //泰语
            preg_match_all('/[\p{Thai}฿]+/u', $text, $matches);
            $ttf = implode('', $matches[0]);
            $w = grapheme_strlen($ttf);
            $segments = preg_split('//u', $ttf, -1, PREG_SPLIT_DELIM_CAPTURE | PREG_SPLIT_NO_EMPTY);
            foreach($segments as  $segment) {
                if (strstr($segment, "ำ")) {
                    $w += 1;
                }
            }
            $w += strlen(iconv("UTF-8", "GBK//IGNORE", $text));
        } else if (preg_match('/[\x{AC00}-\x{D7A3}]/u', $text)) { // 韩语
            $w = strlen(iconv("UTF-8", "EUC-KR//IGNORE", $text));
        } else {
            $w = strlen(iconv("UTF-8", "GBK//IGNORE", $text));
        }
        return $w;
    }

    /**
     * 截取文本
     * @return array
     */
    public static function interceptText($text, $num, $intervalNum)
    {
        $afterText = "";
        if ($num > 0 && $text) {
            $nums = $num - $intervalNum;
            if (preg_match('/[\x{4e00}-\x{9fa5}\x{3040}-\x{309f}\x{30a0}-\x{30ff}]+/u', $text)) {
                $tmpText = mb_substr($text, 0, ceil($nums / 2), 'UTF-8');
                preg_match_all('/[^\x{4e00}-\x{9fa5}\x{3040}-\x{309f}\x{30a0}-\x{30ff}]/u', $tmpText, $matches);
                $tmpTextCount = count($matches[0]);
                if ($tmpTextCount > 1) {
                    $afterText = mb_substr($text, ceil($nums / 2 + $tmpTextCount / 2), 1000, 'UTF-8');
                    $text = mb_substr($text, 0, ceil($nums / 2 + $tmpTextCount / 2), 'UTF-8');
                } else {
                    $afterText = mb_substr($text, ceil($nums / 2), 1000, 'UTF-8');
                    $text = $tmpText;
                }
            } else if (preg_match('/[\p{Thai}]/u', $text)) { //泰语
                $afterText = mb_substr($text, ceil($nums * 1.2), 1000, 'UTF-8');
                $text = mb_substr(iconv("UTF-8", "UTF-8///IGNORE", $text), 0, ceil($nums * 1.2));
            } else if (preg_match('/[\x{AC00}-\x{D7A3}]/u', $text)) { // 韩语
                $afterText = mb_substr($text, ceil($nums / 1.7), 1000, 'UTF-8');
                $text = mb_substr(iconv("UTF-8", "UTF-8///IGNORE", $text), 0, ceil($nums / 1.7));
            } else {
                $afterText = mb_substr($text, $nums, 1000, 'UTF-8');
                $text = mb_substr(iconv("UTF-8", "UTF-8//IGNORE", $text), 0, $nums);
            }
        }
        if (strstr($text,"\n")) {
            $texts = explode("\n", $text);
            $text = $texts[0];
            $afterText = $texts[1] . $afterText;
        } 
        return [$text, $afterText];
    }

    /**
     * filterCharacter
     * @return string
     */
    public static function filterCharacter($text)
    {
        $text = str_replace("​​", "", $text);
        $text = str_replace("　", " ", $text);
        $text = str_replace("ー", "-", $text);
        $text = str_replace("グ", "ク", $text);
        $text = str_replace("・", "·", $text);
        return $text;
    }

    /**
     * 获取打印文本
     * @return string
     */
    public static function printText($leftText, $centerText = "", $rightText = "", $total = 32, $leftNum = 0, $centerNum=0, $rightNum=0, $intervalNum=4)
    {
        $leftText = $leftText == null ? '' : $leftText;
        $centerText = $centerText == null ? '' : $centerText;
        $rightText = $rightText == null ? '' : $rightText;
        // 
        $leftText = self::filterCharacter(trim($leftText));
        $centerText = self::filterCharacter(trim($centerText));
        $rightText = self::filterCharacter(trim($rightText));
        //
        [$leftText, $afterLeftText] = self::interceptText($leftText, $leftNum, $intervalNum);
        [$centerText, $afterCenterText] = self::interceptText($centerText, $centerNum, $intervalNum);
        [$rightText, $afterRightText] = self::interceptText($rightText, $rightNum, $intervalNum);
        // 
        $leftWidth = self::getTextWidth($leftText);
        //
        $leftPadding = $leftNum - $leftWidth > 0 ? str_repeat(" ", intval($leftNum - $leftWidth)) : "";
        $leftPaddingWidth = self::getTextWidth($leftPadding);
        $centerWidth = $centerText == '!' ? 0 :self::getTextWidth($centerText);
        $rightWidth = self::getTextWidth($rightText);
        $centerPaddingWidth = ($total - $leftWidth - $leftPaddingWidth - $centerWidth - $rightWidth);
        $centerPadding = $centerPaddingWidth > 0 ? str_repeat(" ", $centerPaddingWidth) : "";
        //
        $content = $leftText . $leftPadding . $centerText . $centerPadding . $rightText;
        // 
        if ($afterLeftText || $afterCenterText || $afterRightText) {
            return $content .= "\n" . printText($afterLeftText, $afterCenterText, $afterRightText, $total, $leftNum, $centerNum, $rightNum, $intervalNum);
        }
        //
        return $content;
    }
}

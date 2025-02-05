<?php

namespace app\common\library\language\engine;

use \Stichoza\GoogleTranslate\GoogleTranslate;

class Google
{
    /**
     * 翻译
     *
     * @param string $q
     * @param string $to
     * @return string
     * @throws \Exception
     */
    public function translate($q, $from = null, $to = null)
    {
        try {
            if ($from === null) {
                $from = 'auto';
            }
            if ($to === null) {
                $to = $from;
                $from = 'auto';
            }
            $tr = new GoogleTranslate();
            $tr->setTarget($to);
            return $tr->translate($q);
        } catch (\Exception $e) {
            throw new \Exception("Google Translation failed:" . $e->getMessage());
        }
    }
}

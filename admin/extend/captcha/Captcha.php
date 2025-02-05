<?php

namespace captcha;

use help\StringHelp;
use think\facade\Cache;
use think\captcha\Captcha as BaseCaptcha;

/**
 * Class Captcha
 */
class Captcha extends BaseCaptcha
{
    private $key = null;

    /**
     * 验证验证码是否正确
     * @access public
     * @param string $code 用户验证码
     * @return bool 用户验证码是否正确
     */
    public function check(string $code): bool
    {
        $sign = request()->param('sign', request()->header('sign', '') );
        if (Cache::get('captcha_' . $sign) != $code) {
            return false;
        }
        Cache::delete('captcha_' . $sign);
        return true;
    }

    /**
     * 创建验证码
     * @return string
     */
    public function create(string $config = null)
    {
        $v = request()->param('v', 0);
        if ($v != 0) {
            $imgBase64 = 'data:image/jpeg;base64,' . base64_encode(parent::create()->getData());
            //
            $sign = StringHelp::generatePassword(32);
            Cache::set('captcha_' . $sign, $this->key, config('captcha.expire'));
            //
            return json([
                'code' => 1,
                'data' => [
                    'sign' => $sign,
                    'base64'  => $imgBase64
                ],
                'msg' => ''
            ]);
        }
        return parent::create();
    }

    
    /**
     * 创建验证码
     * @return array
     * @throws Exception
     */
    protected function generate(): array
    {
        $bag = '';

        if ($this->math) {
            $this->useZh  = false;
            $this->length = 5;

            $x   = random_int(6, 10);
            $y   = random_int(1, 5);
            if (($x + $y) % 2 != 0) {
                $bag = "{$x} - {$y} = ";
                $key = $x - $y;
            } else {
                $bag = "{$x} + {$y} = ";
                $key = $x + $y;
            }
            $key .= '';
        } else {
            if ($this->useZh) {
                $characters = preg_split('/(?<!^)(?!$)/u', $this->zhSet);
            } else {
                $characters = str_split($this->codeSet);
            }

            for ($i = 0; $i < $this->length; $i++) {
                $bag .= $characters[random_int(0, count($characters) - 1)];
            }

            $key = mb_strtolower($bag, 'UTF-8');
        }

        $this->key = $key;

        $hash = password_hash($key, PASSWORD_BCRYPT, ['cost' => 10]);

        return [
            'value' => $bag,
            'key'   => $hash,
        ];
    }

}

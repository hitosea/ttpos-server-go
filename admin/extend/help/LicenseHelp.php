<?php

namespace help;

use think\facade\Cache;

class LicenseHelp
{
    protected $ffl;

    /**
     * 构造方法
     */
    public function __construct()
    {
        $licensePath = "license.so";
        $architecture = php_uname('m');
        if (strpos($architecture, 'arm') !== false || strpos($architecture, 'aarch64') !== false) {
            $licensePath = "license_arm.so";
        }
        $this->ffl = \FFI::cdef(<<<EOF
            char* NewLicense(char* a, char* b);
            char* SaveLicense(char* a);
            char* GetLicense();
            char* ValidateLicense(char* a, char* b);
            char* GetMacAddress();
        EOF, "/var/www/bin/" . $licensePath);
    }

    /**
     * get
     */
    public function newLicense(array $arr, string $cipher): string
    {
        return \FFI::string($this->ffl->NewLicense(json_encode($arr), $cipher));
    }

    /**
     * save
     */
    public function saveLicense(string $str): array
    {
        try {
            $res = \FFI::string($this->ffl->SaveLicense($str));
            if ($res) {
                Cache::set('licenses', null, 1);
                Cache::set('__SYNC_GET_PUBLICKEY_', 0);
                $res = json_decode($res, true) ?: [];
                if (isset($res['app_id'])) {
                    Cache::store('file')->set(md5($str), $res['app_id'], null);
                }
                return $res;
            }
            //
            return [];
        } catch (\Throwable $th) {
            return [$th->getMessage()];
        }
    }

    /**
     * getSave
     */
    public function getLicense(): string
    {
        return '';
        // return \FFI::string($this->ffl->GetLicense());
    }

    /**
     * get
     */
    public function validateLicense($str = ""): array
    {
        if (!($licenses = Cache::get('licenses')) || $str) {
            try {
                $licenses = \FFI::string($this->ffl->ValidateLicense($str, ''));
                if ($licenses) {
                    $licenses = json_decode($licenses, true);
                    if ($licenses === null) {
                        $licenses = [];
                    }
                } else {
                    $licenses = [];
                }
            } catch (\Throwable $th) {
                $licenses = [];
            }
            if ($str == "") {
                Cache::set('licenses', $licenses, 1 * 60);
            }
        }
        //
        $getLicenseMd5 = md5($this->getLicense());
        if (!Cache::store('file')->get($getLicenseMd5)) {
            if (isset($licenses['app_id'])) {
                Cache::store('file')->set($getLicenseMd5, $licenses['app_id'], null);
            }
        }
        return $licenses ?: [];
    }

    /**
     * get license app id
     */
    public function getLicenseAppId(): string
    {
        return Cache::store('file')->get(md5($this->getLicense())) ?? 0;
    }

    /**
     * get mac addar
     */
    public function getMacAddress(): array
    {
        $result = \FFI::string($this->ffl->GetMacAddress());
        return json_decode($result, true) ?? [];
    }

    
}

<?php

declare(strict_types=1);

namespace app\common\middleware;

use Closure;
use think\Response;
use help\LicenseHelp;
use app\common\exception\BaseException;

/**
 * License
 */
class License
{

    /**
     * @return Response
     */
    public function handle($request, Closure $next, ?array $header = [])
    {
        $licenses = (new LicenseHelp)->validateLicense();
        $request->licenses = $licenses['data'] ?? [];
        // 过滤接口102
        $filter102 = [
            'shop/index/public',
            'shop/passport/login',
            'shop/passport/captcha',
            'shop/index/base',
            'shop/auth.user/getRoleList',
            'shop/index/verifyAuthCode',
            'shop/index/bindList',
            'shop/index/unbind',
        ];
        // 过滤接口101
        $filter101 = [
            'shop/index/authCode',
            'apidoc/config',
            'apidoc/apiMenus',
            'apidoc/docMenus',
            'apidoc/apiDetail',
        ];
        $pathInfo = $request->pathinfo();
        $licenseCode = $licenses['code'] ?? 1;
        if (in_array($pathInfo, $filter102) && ($licenses && $licenseCode != -101)) {
            return $next($request)->header($header);
        }
        //
        if ((!$licenses || $licenseCode != 0) && !in_array($pathInfo, $filter101)) {
            throw new BaseException([
                'msg' => "授权信息错误，请联系销售代表",
                'code' => $licenseCode,
                'data' => array_merge((new LicenseHelp)->getMacAddress(), ['error' => $licenses['msg'] ?? ""]),
            ]);
        }
        //
        return $next($request)->header($header);
    }
}

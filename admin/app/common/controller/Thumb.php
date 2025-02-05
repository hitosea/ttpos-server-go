<?php

namespace app\common\controller;

use image\Imagick;
use think\Request;
use think\Response;
use hg\apidoc\annotation as Apidoc;

/**
 * 缩略图相关
 * @Apidoc\Group("home")
 * @Apidoc\Sort(1)
 */
class Thumb
{
    /**
     * @Apidoc\Title("缩略图地址")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/product/thumbs")
     * @Apidoc\Param("w", type="", require=true, default="", desc="宽度")
     * @Apidoc\Param("h", type="", require=true, default="", desc="高度")
     * @Apidoc\Param("x", type="", require=true, default="2", desc="倍数")
     * @Apidoc\Returned()
     */
    public function index(Request $request)
    {
        $param = $request->param();
        $x = (int)($param['x'] ?? 2);
        $width = (int)($param['w'] ?? 0);
        $height = (int)($param['h'] ?? 0);
        $defaultPath = public_path('/image/product') . 'default.png';
        // 设置缓存有效期为7天
        header('Cache-Control: max-age=604800');
        // 参数错误-返回默认图
        if (!isset($param['catalogue']) || !isset($param['date']) || !isset($param['name'])) {
            header('Content-Type: image/jpeg');
            return Response::create(readfile($defaultPath))->header(['Content-Type' => 'image/jpeg']);
        }
        // 文件不存在-返回默认图
        if ($param['shop'] ?? '') {
            $path = public_path('/uploads/' . $param['shop'] . '/' . $param['date']) . $param['name'];
        } else {
            $path = public_path('/uploads/' . $param['date']) . $param['name'];
        }
        if (!file_exists($path)) {
            return Response::create(readfile($defaultPath))->header(['Content-Type' => 'image/jpeg']);
        }
        // 没有传宽高-返回原图
        if (!$width || !$height) {
            return Response::create(readfile($path))->header(['Content-Type' => 'image/' . pathinfo($param['name'], PATHINFO_EXTENSION)]);
        }
        if ($width < 10) $width = 50;
        if ($height < 10) $height = 50;
        if ($x < 0) $x = 2;
        // 已存在缩略图则直接返回
        $thumbPath = preg_replace('/(\.\w+)$/', '_' . ($width * $x) . 'X' . ($height * $x) . '$1', $path);
        if (file_exists($thumbPath)) {
            return Response::create(readfile($thumbPath))->header(['Content-Type' => 'image/' . pathinfo($param['name'], PATHINFO_EXTENSION)]);
        }
        //
        Imagick::thumb($path, ($width * $x), ($height * $x), $thumbPath);
        //
        return Response::create(readfile($thumbPath))->header(['Content-Type' => 'image/' . pathinfo($param['name'], PATHINFO_EXTENSION)]);
    }
}

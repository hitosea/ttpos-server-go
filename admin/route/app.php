<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2018 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------
use think\facade\Route;
use app\common\controller\Thumb;

// 缩略图 - shop分割
Route::get('product/thumb/:catalogue/:shop/:date/:name', Thumb::class . '@index');
Route::get('product/thumbs/:catalogue/:shop/:date/:name', Thumb::class . '@index');

// 缩略图
Route::get('product/thumb/:catalogue/:date/:name', Thumb::class . '@index');
Route::get('product/thumbs/:catalogue/:date/:name', Thumb::class . '@index');

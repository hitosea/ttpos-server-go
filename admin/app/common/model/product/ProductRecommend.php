<?php

namespace app\common\model\product;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 商品推荐
 */
class ProductRecommend extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_package_recommend';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $append = [];
}

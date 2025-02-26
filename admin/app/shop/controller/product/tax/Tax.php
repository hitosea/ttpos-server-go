<?php

namespace app\shop\controller\product\tax;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\tax\TaxCategory;

/**
 * 税类
 * @Apidoc\Group("product")
 * @Apidoc\Sort(4)
 */
class Tax extends Controller
{
    /**
     * @Apidoc\Title("税类列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.tax.tax/list")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\common\model\tax\TaxCategory\getList")
     */
    public function list()
    {
        $data = $this->request->param();
        $model = new TaxCategory();
        $list = $model->getList($data);
        foreach ($list as $item) {
            $item['id'] = $item['uuid'];
        }
        return $this->renderSuccess('', compact('list'));
    }
}

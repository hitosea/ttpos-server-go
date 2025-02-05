import request from '@/utils/request';

const ProductImportsApi = {
  // 导入文件，获取预览商品
  getProductImportsList(data, errorback) {
    return request._post('/shop/product.store.product/imports', { ...data, mode: 'get' }, errorback);
  },
  // 预览保存数据
  fetchProductImportsList(data, errorback) {
    return request._post('/shop/product.store.product/imports', { ...data, mode: 'save' }, errorback);
  },
};

export default ProductImportsApi;

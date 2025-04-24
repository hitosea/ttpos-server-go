<?php

namespace app\shop\controller\file;

use app\common\service\example\ExampleService;
use app\shop\model\file\UploadFile;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\settings\Setting as SettingModel;
use app\shop\controller\Controller as BaseController;
use app\common\library\storage\Driver as StorageDriver;
use app\shop\model\product\Material;
use app\shop\model\product\Product;
use think\facade\Db;
use think\facade\Log;
use think\file\UploadedFile;

/**
 * 文件库管理
 * @Apidoc\Group("home")
 * @Apidoc\Sort(5)
 */
class Upload extends BaseController
{
    /**
     * @Apidoc\Title("图片上传接口")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/shop/file.Upload/image")
     * @Apidoc\Param("iFile", type="file", default="", desc="文件")
     * @Apidoc\Returned("id", type="string", default="1", desc="文件的id")
     * @Apidoc\Returned("file_path", type="string", default="http://127.0.0.1/uploads/20240327/8018f81f7285a7403937635c338e747c.png", desc="使用这个字段去预览保存")
     */
    public function image($group_id = -1)
    {
        $fileSource = request()->param('source') ?? '';
        // 实例化存储驱动
        $config = SettingModel::getItem('storage');
        $storageDriver = new StorageDriver($config);
        // 设置上传文件的信息
        $fileInfo = $storageDriver->setUploadFile('iFile');
        // 图片信息
        $fileType = request()->param('file_type') ?? 'image';
        // 检查文件类型
        $extension = strtolower($fileInfo->extension());
        // 图片上传
        if ($fileType == 'image') {
            if (!in_array($extension, ['jpg', 'jpeg', 'png', 'webp'])) {
                return json(['code' => 0, 'msg' => '仅支持JPG、JPEG、PNG格式']);
            }
        }
        // 视频上传
        if ($fileType == 'video') {
            $size = request()->param('size') ?? 30;
            if (!in_array($extension, ['avi', 'mpeg', 'mov', 'mp4'])) {
                return json(['code' => 0, 'msg' => '仅支持AVI、MPEG、MOV、MP4格式']);
            }
            // 检查文件大小
            $maxSize = $size * 1024 * 1024;
            if ($fileInfo->getSize() > $maxSize) {
                return json(['code' => 0, 'msg' => $size > 30 ? '文件大小不能超过30MB' : '文件大小不能超过10MB']);
            }
        }
        //
        $saveName = $storageDriver->upload($fileSource == '' ? 500 : 5000);
        if ($saveName == '') {
            return json(['code' => 0, 'msg' => '上传失败' . $storageDriver->getError()]);
        }
        $saveName = str_replace('\\', '/', $saveName);
        // 图片上传路径
        $fileName = $storageDriver->getFileName();
        // 访问参数
        $urlParam = $storageDriver->getUrlParam();
        // 添加文件库记录
        $uploadFile = $this->addUploadFile($group_id, $fileName, $fileInfo, $fileType, $saveName, $urlParam);
        // 图片上传成功
        return json(['code' => 1, 'msg' => '图片上传成功', 'data' => $uploadFile]);
    }

    /**
     * 添加文件库上传记录
     */
    private function addUploadFile($group_id, $fileName, $fileInfo, $fileType, $saveName, $urlParam = '')
    {
        // 存储引擎
        $config = SettingModel::getItem('storage');
        $storage = $config['default'];
        // 存储域名
        $fileUrl = isset($config['engine'][$storage]['domain']) ? $config['engine'][$storage]['domain'] : '';
        // 添加文件库记录
        $model = new UploadFile;
        $real_name = $fileInfo->getOriginalName();
        $model->save([
            'uuid'=> createUuid(),
            'group_uuid' => $group_id > 0 ? (int)$group_id : 0,
            'storage' => $storage,
            'file_url' => $fileUrl,
            'file_name' => $fileName,
            'save_name' => $saveName,
            'file_size' => $fileInfo->getSize(),
            'file_type' => $fileType,
            'extension' => $fileInfo->getOriginalExtension(),
            'real_name' => $real_name,
            'index_file_name' => pathinfo($real_name, PATHINFO_FILENAME),
            'url_param' => $urlParam,
        ]);
        return $model;
    }

    /**
     * 批量移动文件分组
     */
    public function moveFiles($group_id, $fileIds)
    {
        $model = new UploadFile;
        if ($model->moveGroup($group_id, $fileIds) !== false) {
            return $this->renderSuccess('移动成功');
        }
        return $this->renderError($model->getError() ?: '移动失败');
    }

    public function repeatList()
    {
        $page = request()->param('page') ?? 1;
        $listRows = request()->param('list_rows') ?? 10;
        $product_list = request()->param('list') ?? '[]';
        $list = [];
        $result = [
            'current_page' => $page,
            'data' => $list,
            'per_page' => $listRows,
            'total' => 0,
            'last_page' => 0,
        ];

        $product_list_assoc = array_column($product_list, null, 'img_name');
        $img_names = array_keys($product_list_assoc);
        if (empty($img_names)) {
            return $this->renderSuccess('', $result);
        }

        // 成品
        $productSql = Product::alias('p')
            ->field(implode(',', [
                'file.real_name',
                'file.index_file_name',
                'p.create_time',
                'p.uuid as product_id', 
                'p.name as product_name', 
                'p.sort as product_sort',
                'c1.name as category_name', 
                'c2.name as category_parent_name', 
                'c1.uuid as category_uuid',
                'c2.uuid as category_parent_uuid',
                'p.sort as sort',
                'p.id as id'
            ]))
            ->leftJoin('product_category c1', 'p.category_uuid = c1.uuid')
            ->leftJoin('product_category c2', 'c1.parent_uuid = c2.uuid')
            ->leftJoin('file', 'p.image_file_uuid = file.uuid')
            ->where('p.delete_time', 0)
            ->where('file.delete_time', 0)
            ->whereIn('file.index_file_name', $img_names)
            ->group('p.uuid')
            ->order('p.sort', 'asc')
            ->order('p.id', 'desc')
            ->buildSql();

            // 材料
        $materialSql = Material::alias('m')
        ->field(implode(',', [
            'file.real_name',
            'file.index_file_name',
            'm.create_time',
            'm.uuid as product_id', 
            'm.name as product_name', 
            '0 as product_sort',
            'c1.name as category_name', 
            'c2.name as category_parent_name', 
            'c1.uuid as category_uuid',
            'c2.uuid as category_parent_uuid',
            '0 as sort',
            'm.id as id'
        ]))
        ->leftJoin('product_category c1', 'm.category_uuid = c1.uuid')
        ->leftJoin('product_category c2', 'c1.parent_uuid = c2.uuid')
        ->leftJoin('file', 'm.image_uuid = file.uuid')
        ->where('m.delete_time', 0)
        ->whereIn('file.index_file_name', $img_names)
        ->group('m.uuid')
        ->order('m.id', 'desc')
        ->buildSql();

        // 分页
        $offset = ($page - 1) * $listRows;
        $limit = $listRows;

        $querySql = "SELECT " . implode(',', [
            'create_time',
            'category_name', 
            'category_parent_name', 
            'real_name',
            'index_file_name',
            'product_id', 
            'product_name', 
            'product_sort',
            'category_uuid',
            'category_parent_uuid',
            'sort',
            'id'
        ]) . " FROM ($productSql UNION ALL $materialSql) AS all_product";
        $orderSql = ' ORDER BY sort ASC, id DESC';
        $pageSql = " LIMIT {$offset}, {$limit}";

        // 执行查询
        $countSql = "SELECT COUNT(*) AS total_count". " FROM ($productSql UNION ALL $materialSql) AS all_product";
        $count = Db::connect('shop' . request()->appId)->query($countSql);
        $total = $count[0]['total_count'];
        $rows = Db::connect('shop' . request()->appId)->query($querySql . $orderSql . $pageSql);

        foreach ($rows as $row) {
            // 分类
            $pathNameText = extractLanguage($row['category_name']);
            if ($row['category_parent_name']) {
                $pathNameText = extractLanguage($row['category_parent_name']) . '-' . $pathNameText;
            }
            $list[] = [
                'product_id' => $row['product_id'],
                'product_name_text' => extractLanguage($row['product_name']),
                'category_name_text' => $pathNameText,
                'img_name' => $row['index_file_name'],
                'file_name' => $row['real_name'],
            ];
        }

        $result['total'] = $total;
        $result['last_page'] = ceil($total / $listRows);
        $result['data'] = $list;

        return $this->renderSuccess('', $result);
    }

    /**
     * @Apidoc\Title("批量替换商品图片")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/shop/file.Upload/batchReplaceProductImage")
     * @Apidoc\Param("iFile", type="file", default="", desc="文件")
     * @Apidoc\Param("list", type="array", default="", desc="替换图片数组")
     * @Apidoc\Param("iFile", type="file", default="", desc="文件")
     */
    public function batchReplaceProductImage()
    {
        //
        $fileInfo = request()->file('iFile');
        $isOverlay = request()->param('is_overlay') ?? 0;
        $fileSource = request()->param('source') ?? '';
        $request_cache = md5(mt_rand(0, 9999));
        $product_list = request()->param('list') ?? '[]';
        $product_list_arr = json_decode($product_list, true) ?: [];
        $insert_list = $product_list_arr;

        // 判断文件是否上传成功
        if ($fileInfo) {
            /**
             * step1 解压保存图片到临时目录
             */
            $zip = new \ZipArchive;
            $extractPath = root_path() . "/runtime/storage/tmp_zip_img_{$request_cache}/";
            if (!file_exists($extractPath)) {
                mkdir($extractPath, 0777, true);
            }
            if ($zip->open($fileInfo->getPathname()) === TRUE) {
                $zip->extractTo($extractPath);
                $zip->close();
            } else {
                return $this->renderError('解压失败');
            }

            /**
             * step2 存入图片库
             */
            $insert_file_res_arr = [];    // 已入库图片
            // 实例化存储驱动
            $config = SettingModel::getItem('storage');
            $storageDriver = new StorageDriver($config);
            // 读取解压后的文件内容
            $files = array_diff(scandir($extractPath), ['.', '..']);
            $total = 0;
            $insert = 0;
            $model = new UploadFile();
            $maxRetries = 3;
            $attempt = 0;

            while ($attempt < $maxRetries) {
                $model->startTrans();
                try {
                    foreach ($files as $file) {
                        $total++;
                        $tmp_img_path = $extractPath . $file;
                        $imgFileInfo = new UploadedFile($tmp_img_path, $file);
                        // 检查文件类型
                        $extension = strtolower($imgFileInfo->extension());
                        // 图片上传
                        if (!in_array($extension, ['jpg', 'jpeg', 'png', 'webp'])) {
                            // 格式不对先跳过
                            continue;
                        }
                        // 设置上传文件的信息
                        $storageDriver->setLocalFile($imgFileInfo);
                        //
                        $saveName = $storageDriver->upload($fileSource == '' ? 500 : 5000);
                        if ($saveName == '') {
                            // 图片上传失败跳过
                            continue;
                        }
                        $saveName = str_replace('\\', '/', $saveName);
                        // 图片上传路径
                        $fileName = $storageDriver->getFileName();
                        // 访问参数
                        $urlParam = $storageDriver->getUrlParam();
                        // 添加文件库记录
                        $uploadFile = $this->addUploadFile(0, $fileName, $imgFileInfo, 'image', $saveName, $urlParam);
                        if ($uploadFile) {
                            // 覆盖删除同名数据
                            if ($isOverlay) {
                                $imageUuidList = UploadFile::where('index_file_name', $uploadFile->index_file_name)->where('id', '<', $uploadFile->id)->column('uuid');
                                UploadFile::where('index_file_name', $uploadFile->index_file_name)->where('id', '<', $uploadFile->id)->delete();
                                // 用到该图的商品都图换
                                Product::whereIn('image_file_uuid', $imageUuidList)->update(['image_file_uuid' => $uploadFile->uuid]);
                                Material::whereIn('image_uuid', $imageUuidList)->update(['image_uuid' => $uploadFile->uuid]);
                            }
                            $insert++;
                            // 插入商品对应的 image_id
                            foreach ($insert_list as $key => &$item_p) {
                                if ($item_p['img_name'] == $uploadFile->index_file_name) {
                                    $item_p['image_id'] = $uploadFile->uuid;
                                    $insert_file_res_arr[] = $item_p;
                                    unset($insert_list[$key]);
                                }
                            }
                        }
                    }

                    /**
                     * step3 替换新图片
                     */
                    if (!empty($insert_file_res_arr)) {
                        // 商品
                        $productExistingRecords = [];
                        $productData = [];
                        // 材料
                        $materialExistingRecords = [];
                        $materialData = [];
                        // 
                        $productIds = array_unique(array_column($insert_file_res_arr, 'product_id'));
                        //
                        foreach (array_chunk($productIds, 1000) as $chunk) {
                            $productRecords = Product::whereIn('uuid', $chunk)->select()->toArray();
                            $productExistingRecords = array_merge($productExistingRecords, $productRecords);

                            $materialRecords = Material::whereIn('uuid', $chunk)->select()->toArray();
                            $materialExistingRecords = array_merge($materialExistingRecords, $materialRecords);
                        }

                        $existingProductIds = array_column($productExistingRecords, 'uuid');
                        $existingMaterialIds = array_column($materialExistingRecords, 'uuid');

                        foreach ($insert_file_res_arr as $item) {
                            if (in_array($item['product_id'], $existingProductIds)) {
                                $productDataItem = [
                                    'uuid' => $item['product_id'],
                                    'image_file_uuid' => $item['image_id']
                                ];
                                $productData[] = $productDataItem;
                            }
                            if (in_array($item['product_id'], $existingMaterialIds)) {
                                $materialDataItem = [
                                    'uuid' => $item['product_id'],
                                    'image_uuid' => $item['image_id']
                                ];
                                $materialData[] = $materialDataItem;
                            }
                        }

                        // 批量更新
                        if (!empty($productData)) {
                            foreach (array_chunk($productData, 1000) as $list) {
                                foreach ($list as $item) {
                                    Product::where('uuid', $item['uuid'])->update(['image_file_uuid' => $item['image_file_uuid']]);
                                }
                            }
                        }
                        if (!empty($materialData)) {
                            foreach (array_chunk($materialData, 1000) as $list) {
                                foreach ($list as $item) {
                                    Material::where('uuid', $item['uuid'])->update(['image_uuid' => $item['image_uuid']]);
                                }
                            }
                        }
                    }

                    $model->commit();
                    break;
                } catch (\PDOException $e) {
                    $model->rollback();
                    if ($e->getCode() == 1213 && $attempt < $maxRetries - 1) {
                        $attempt++;
                        sleep(1); // 等待后重试
                    } else {
                        return $this->renderError('数据库错误: ' . $e->getMessage());
                    }
                }
            }

            // 清除临时图片文件
            (new ExampleService)->deleteDirectory($extractPath);

            return $this->renderSuccess('上传成功', compact('total', 'insert'));
        } else {
            return $this->renderError('上传文件失败');
        }
    }
}

<?php

namespace app\shop\model\supplier;

use app\common\model\supplier\Printing as PrintingModel;
use think\facade\Cache;

/**
 * 菜品打印模型
 */
class Printing extends PrintingModel
{
    /**
     * 获取列表数据
     */
    public function getLists($params)
    {
        $model = $this;
        // 查询列表数据
        $paginate = $model->with([
            'printingItem' => [ 'printer' ],
        ])->order(['create_time' => 'desc'])->paginate($params);

        $list = [];
        foreach ($paginate->items() as $item) {
            $list[] = [
                'id' => $item['id'],
                'name' => $item['name'],
                'printer_name_text' => self::getPrinterNameText($item['printingItem']),
                'printer_list' => self::getPrinterList($item['printingItem']),
                'print_type' => self::PRINT_MODE_REVERSE_MAP[$item['print_mode']],
                'print_method' => self::PRINT_METHOD_REVERSE_MAP[$item['print_method']],
                'is_open' => $item['status'],
                'create_time' => $item['create_time'],
                'is_usb' => $item['is_usb'],
            ];
        }

        return [
            'current_page' => $paginate->currentPage(),
            'last_page' => $paginate->lastPage(),
            'per_page' => $paginate->listRows(),
            'total' => $paginate->total(),
            'data' => $list,
        ];
    }

    /**
     * 添加
     */
    public function add($data)
    {
        // 开启事务
        $this->startTrans();
        try {
            // 添加商品打印(档口)
            $this->save([
                'name' => $data['name'], // 名称
                'copies' => $data['copies'] ?? 1, // 打印份数
                'status' => intval($data['is_open']), // 是否开启: 0开启 1关闭
                'print_mode' => self::PRINT_MODE_MAP[intval($data['print_type'])], // 打印模式
                'print_method' => self::PRINT_METHOD_MAP[intval($data['print_method'])], // 打印方式
                'print_product_select' => self::PRINT_PRODUCT_SELECT_MAP[intval($data['product_method'])], // 打印商品选择
                'print_mode_scene' => self::PRINT_MODE_SCENE_MAP[intval($data['print_select'])], // 打印场景
            ]);
            // 添加商品打印详情
            $itemList = [];
            foreach ($data['printer_id'] ?: [] as $id) {
                $itemList[$id] = [
                    'product_printer_uuid' => intval($this->uuid),
                    'printer_uuid' => intval($id),
                    'create_time' => time(),
                    'update_time' => time(),
                ];
            }
            (new PrintingItem())->saveAll($itemList);

            // 添加打印的商品
            $productList = [];
            foreach ($data['product_ids'] ?: [] as $id) {
                $productList[$id] = [
                    'product_printer_uuid' => $this->uuid,
                    'product_package_uuid' => $id,
                    'create_time' => time(),
                    'update_time' => time(),
                ];
            }
            (new PrintingProduct())->saveAll($productList);
            
            // 添加打印区域
            $areaList = [];
            foreach ($data['area_id'] ?: [] as $id) {
                $areaList[$id] = [
                    'product_printer_uuid' => $this->uuid,
                    'desk_region_uuid' => $id,
                    'create_time' => time(),
                    'update_time' => time(),
                ];
            }
            (new PrintingRegion())->saveAll($areaList);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 修改
     */
    public function edit($data, $index = 0)
    {  
        $lockKey = sprintf("PRINTING_EDIT_LOCK_%s_%d", self::$app_id, $this['id']);
        $lockValue = uniqid('lock_', true);

        // 开启事务
        $this->startTrans();
        try {
            // 检查锁是否已存在
            if (Cache::has($lockKey)) {
                $this->error = '数据处理中，请稍后重试';
                return false;
            }

            // 设置锁，120秒过期
            Cache::set($lockKey, $lockValue, 120);

            // 物理删除打印机
            $this->printingItem()->withTrashed()->chunk(500, function ($items) {
                foreach ($items as $item) {
                    $item->force(true)->delete();
                }
            });

            // 物理删除打印的区域
            $this->printingRegion()->withTrashed()->chunk(500, function ($items) {
                foreach ($items as $item) { 
                    $item->force(true)->delete();
                }
            });

            // 物理删除打印的商品
            $this->printingProductItem()->withTrashed()->chunk(500, function ($items) {
                foreach ($items as $item) {
                    $item->force(true)->delete();
                }
            });

            // 添加商品打印详情
            $itemList = [];
            foreach ($data['printer_id'] as $id) {
                $itemList[$id] = [
                    'product_printer_uuid' => intval($this->uuid),
                    'printer_uuid' => intval($id),
                    'create_time' => time(),
                    'update_time' => time(),
                ];
            }
            (new PrintingItem())->saveAll($itemList);

            // 添加打印的商品
            $productList = [];
            foreach ($data['product_ids'] ?: [] as $id) {
                $productList[$id] = [
                    'product_printer_uuid' => $this->uuid,
                    'product_package_uuid' => $id,
                    'create_time' => time(),
                    'update_time' => time(),
                ];
            }
            (new PrintingProduct())->saveAll($productList);
            
            // 添加打印区域
            $areaList = [];
            foreach ($data['area_id'] ?: [] as $id) {
                $areaList[$id] = [
                    'product_printer_uuid' => $this->uuid,
                    'desk_region_uuid' => $id,
                    'create_time' => time(),
                    'update_time' => time(),
                ];
            }
            (new PrintingRegion())->saveAll($areaList);

            $this->save([
                'name' => $data['name'], // 名称
                'copies' => $data['copies'] ?? 1, // 打印份数
                'status' => intval($data['is_open']), // 是否开启: 0开启 1关闭
                'print_mode' => self::PRINT_MODE_MAP[intval($data['print_type'])], // 打印模式
                'print_method' => self::PRINT_METHOD_MAP[intval($data['print_method'])], // 打印方式
                'print_product_select' => self::PRINT_PRODUCT_SELECT_MAP[intval($data['product_method'])], // 打印商品选择
                'print_mode_scene' => self::PRINT_MODE_SCENE_MAP[intval($data['print_select'])], // 打印场景
            ]);
            $this->commit();
            Cache::delete($lockKey);
            return true;
        } catch (\Exception $th) {
            $this->rollback();
            Cache::delete($lockKey);
            trace($th->getMessage());
            trace($th->getTraceAsString()); 
            // 
            if (strpos($th->getMessage(), 'Duplicate entry') !== false) {
                if ($index > 10) {
                    $this->error = "更新失败";
                    return false;
                }
                return $this->edit($data, $index + 1);
            }
            $this->error = "更新失败";
            return false;
        }
    }

    /**
     * 设置状态
     */
    public function setStatus($status)
    {
        return $this->save(['status' => $status ? 1 : 0]);
    }

    /**
     * 软删除
     */
    public function setDelete()
    {
        // 开启事务
        $this->startTrans();
        try {
            $this['printingItem']->delete();
            $this['printingProductItem']->delete();
            $this['printingRegion']->delete();
            $this->delete();

            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }
}

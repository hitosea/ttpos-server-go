<?php

namespace app\common\model\erp;

use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\Material;
use app\common\model\product\ProductBom;
use app\shop\model\product\RelatedMaterial;

/**
 * 报损记录模型
 */
class ErpDamagedProductRecord extends BaseModel
{
    use SoftDelete;
    protected $name = 'loss_report_form';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['number', 'type', 'product_id', 'product_sku_id', 'review_status', 'operator_id', 'refused'];

    /**
     * 报损类型, 0-损耗 1-丢失
     */
    const SCENE_LOSS = 0;
    const SCENE_LOST = 1;
    const OLD_SCENE_LOSS = [
        1 => self::SCENE_LOST,
        2 => self::SCENE_LOSS
    ];

    /**
     * 状态,0-pending待审核 1-approved已通过 2-rejected已驳回
     */
    const STATUS_PENDING = 0;
    const STATUS_APPROVED = 1;
    const STATUS_REJECTED = 2;

    /**
     * 兼容字段
     */
    public function getNumberAttr($value, $data = [])
    {
        return $this?->form_no ?: '';
    }
    public function getTypeAttr($value, $data = [])
    {
        return $this?->scene ?: 0;
    }
    public function getProductIdAttr($value, $data = [])
    {
        return $this?->material_uuid ?: 0;
    }
    public function getProductSkuIdAttr($value, $data = [])
    {
        return $this?->material_uuid ?: 0;
    }
    public function getReviewStatusAttr($value, $data = [])
    {
        return $this?->status ?: 0;
    }
    public function getOperatorIdAttr($value, $data = [])
    {
        return $this?->operator_uuid ?: 0;
    }
    public function getRefusedAttr($value, $data = [])
    {
        return $this?->reject_reason ?: 0;
    }

    /**
     * 报损数量
     */
    public function getNumAttr($value)
    {
        return floatval($value);
    }

    /**
     * 操作人
     */
    public function operator()
    {
        return $this->belongsTo(User::class, 'operator_uuid', 'uuid')->field(['uuid', 'uuid as shop_user_id', 'username as user_name', 'real_name']);
    }

    /**
     * 关联商品规格
     */
    public function sku()
    {
        return $this->belongsTo(ProductBom::class, 'product_bom_uuid', 'uuid');
    }

    public function material()
    {
        return $this->belongsTo(Material::class, 'material_uuid', 'uuid');
    }


    /**
     * 获取列表
     *
     * @param [type] $params
     * @return object
     */
    public function getList($params)
    {
        $model = new self;
        $type = null;
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        }

        if (isset($params['type']) && $params['type']) {
            $type = $params['type'];
        }

        $paginate = $model
            ->with([
                'sku' => function($q) {
                    return $q->with([ 'product' => function($q) {
                        return $q->with([ 'image', 'category' ])->withTrashed();
                    }])->withTrashed();
                }, 
                'material' => function($q) {
                    return $q->with([ 'image', 'category' ])->withTrashed();
                }, 
                'operator',
            ])
            ->when($type && $type > 0, function ($q) use ($type) {
                $q->where('scene', self::OLD_SCENE_LOSS[$type]);
            })
            ->when($start_time && $end_time, function ($q) use ($start_time, $end_time) {
                $q->where('create_time', 'between', [strtotime($start_time), strtotime($end_time)]);
            })
            ->order('create_time desc')->paginate($params);

        $list = [];
        foreach ($paginate->items() as $item) {
            $productId = 0;
            $sku = [];
            // 规格
            if ($item['sku']) {
                $productId = $item['product_bom_uuid'];
                $product = [
                    'type' => 10,
                    'product_name_text' => extractLanguage($item['sku']['product']['name']),
                    'image' => $item['sku']['product']['image'],
                    'category' => [
                        'path_name_text' => $item['sku']['product']['category']['path_name_text'],
                    ],
                ];
                $sku['product'] = $product;
                $sku['spec_name_text'] = extractLanguage($item['sku']['name']);
                $sku['product_price'] = $item['sku']['product_price'];
                $sku['stock_num'] = floatval($item['sku']['stock_num']);
                $sku['material_stock'] = 0;
                $sku['product_sales'] = floatval($item['sku']['actual_sale_num']);
                $sku['create_time'] = $item['sku']['create_time'];
            }
            // 材料
            if ($item['material']) {
                $productId = $item['material_uuid'];
                $product = [
                    'type' => 20,
                    'product_name_text' => extractLanguage($item['material']['name']),
                    'image' => [$item['material']['image']],
                    'category' => [
                        'path_name_text' => $item['material']['category']['path_name_text'],
                    ],
                ];
                $sku['product'] = $product;
                $sku['spec_name_text'] = '';
                $sku['product_price'] = '0.00';
                $sku['product_stock'] = 0;
                $sku['material_stock'] = floatval($item['material']['stock_num']);
                $sku['product_sales'] = floatval($item['material']['actual_sale_num']);
                $sku['create_time'] = $item['material']['create_time'];
            }
            $list[] = [
                'id' => $item['uuid'],
                'product_id' => $productId,
                'number' => $item['form_no'],
                'type' => $item['scene'] == 0 ? 2 : 1,
                'num' => $item['num'],
                'remark' => $item['remark'],
                'operator' => $item['operator'],
                'sku' => $sku,
                'create_time' => $item['create_time'],
                'review_status' => $item['review_status'],
                'approved_time' => $item['approved_time'] > 0 ? date('Y-m-d H:i:s', $item['approved_time']) : '',
                'rejected_time' => $item['revoke_time'] > 0 ? date('Y-m-d H:i:s', $item['revoke_time']) : '',
                'refused' => $item['reject_reason'],
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
     * 报损记录柱状图
     * @param $params
     * @return mixed
     */
    public function getChartList($params)
    {
        $type = null;
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
            $startTime = date($year . "-" . $month . "-01");
            $endTime = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $startTime = date($year . "-" . $month . "-01");
            $endTime = date($year . "-" . $month . "-t 23:59:59");
        }
        if (isset($params['type']) && $params['type']) {
            $type = $params['type'];
        }

        $records = self::field('uuid,product_bom_uuid,material_uuid,num')
            ->with([
                'sku' => function($q) use ($year, $month, $startTime, $endTime) {
                    return $q->field('uuid,product_package_uuid,name,barcode_value')->with([
                            'product' => function($q) {
                                return $q->field('uuid,category_uuid,name')->with([
                                        'category' => function($q) {
                                            return $q->field('uuid,parent_uuid,name')->with([
                                                'parent' => function($q) {
                                                    return $q->field('uuid,name')->withTrashed();
                                                }
                                            ])->withTrashed();
                                        }
                                    ])->withTrashed();
                            },
                            'erpMonthlyProductStatistics' => function($q) use ($year, $month) {
                                return $q->where('year', $year)
                                    ->where('month', $month)
                                    ->where('scene', ErpMonthlyProductStatistics::MONTH_START);
                            },
                            'erpInventoryRecord' => function($q) use ($startTime, $endTime) {
                                return $q->where('status', 0)
                                    ->whereIn('scene', [0, 1, 2])
                                    ->where('create_time', 'between', [strtotime($startTime), strtotime($endTime)]);
                            }
                        ])->withTrashed();
                },
                'material' => function($q) use ($year, $month, $startTime, $endTime) {
                    return $q->field('uuid,category_uuid,name')->with([
                        'category' => function($q) {
                            return $q->field('uuid,parent_uuid,name')->with([
                                'parent' => function($q) {
                                    return $q->field('uuid,name')->withTrashed();
                                }
                            ])->withTrashed();
                        },
                        'erpMonthlyMaterialStatistics' => function($q) use ($year, $month) {
                            return $q->where('year', $year)
                                ->where('month', $month)
                                ->where('scene', ErpMonthlyProductStatistics::MONTH_START);
                        },
                        'erpInventoryRecord' => function($q) use ($startTime, $endTime) {
                            return $q->where('status', 0)
                                ->whereIn('scene', [0, 1, 2])
                                ->where('create_time', 'between', [strtotime($startTime), strtotime($endTime)]);
                        }
                    ])->withTrashed();
                }
            ])
            ->when($type && $type > 0, function($q) use ($type) {
                $q->where('scene', self::OLD_SCENE_LOSS[$type]);
            })
            ->where('status', self::STATUS_APPROVED)
            ->where('create_time', 'between', [strtotime($startTime), strtotime($endTime)])
            ->select();
        
        $list = [];
        foreach ($records as $record) {
            $monthStartStock = 0; // 月初库存数
            $monthEntryStock = 0; // 月入库存数
            if ($record->sku) {
                foreach ($record->sku->erpMonthlyProductStatistics as $item) {
                    $monthStartStock = helper::bcadd($monthStartStock, $item->stock);
                }
                foreach ($record->sku->erpInventoryRecord as $item) {
                    $monthEntryStock = helper::bcadd($monthEntryStock, $item->num);
                }  
                
                $category = $record->sku->product->category;
            } else {
                foreach ($record->material->erpMonthlyMaterialStatistics as $item) {
                    $monthStartStock = helper::bcadd($monthStartStock, $item->stock);
                }
                foreach ($record->material->erpInventoryRecord as $item) {
                    $monthEntryStock = helper::bcadd($monthEntryStock, $item->num);
                }

                $category = $record->material->category;
            }
            $totalEntryStock = helper::bcadd($monthEntryStock, $monthStartStock, 4);

            if ($category->parent_uuid == 0) {
                $categoryUuid = $category->uuid;
                $categoryName = $category->name_text;
            } else {
                $categoryUuid = $category->parent->uuid;
                $categoryName = $category->parent->name_text;
            }
            $list[$categoryUuid]['category_id'] = $categoryUuid;
            $list[$categoryUuid]['name'] = $categoryName;
            $list[$categoryUuid]['damage_count'] = helper::bcadd($list[$categoryUuid]['damage_count'] ?? 0, $record['num'], 4);
            $list[$categoryUuid]['entry_stock'] = helper::bcadd($list[$categoryUuid]['entry_stock'] ?? 0, $totalEntryStock, 4);
        }
        $list = array_values($list);
        foreach ($list as $key => $item) {
            $item['damage_ratio'] = $item['entry_stock'] > 0 ? helper::bcdiv($item['damage_count'], $item['entry_stock'], 5) : 0;
            $item['damage_ratio'] = floatval(helper::bcmul($item['damage_ratio'], 100, 2)) . '%';
            $item['damage_count'] = floatval($item['damage_count']);
            $list[$key] = $item;
        }
        usort($list, function ($item1, $item2) {
            return $item2['damage_count'] <=> $item1['damage_count'];
        });

        return $list;
    }

    /**
     * 详情
     */
    public function detail($id): self
    {
        $model = new self;
        return $model->with(['operator'])
            ->where('uuid', $id)
            ->find();
    }

    /**
     * 新增
     */
    public function add($params)
    {
        $model = new self;
        $num = $params['num'] ?? 0;
        if (!$num) {
            $this->error = '数量不能为0';
            return false;
        }
        $productBom = ProductBom::where('uuid', $params['product_sku_id'])->find();
        $material = Material::where('uuid', $params['product_sku_id'])->find();

        $productBomUuid = 0;
        $materialUuid = 0;

        $stock_num = 0;
        if ($productBom) {
            $stock_num = ProductBom::where('uuid', $params['product_sku_id'])->value('stock_num');
            $productBomUuid = $params['product_sku_id'];
        }

        // 原料
        if ($material) {
            $stock_num = Material::where('uuid', $params['product_sku_id'])->value('stock_num');
            $materialUuid = $params['product_sku_id'];
        }

        if ($num > $stock_num) {
            $this->error = '不能大于库存数量';
            return false;
        }
        //
        $data['form_no'] = $model->generateNumber();
        $data['scene'] = self::OLD_SCENE_LOSS[$params['type'] ?? 1];
        $data['product_bom_uuid'] = $productBomUuid;
        $data['material_uuid'] = $materialUuid;
        $data['num'] = $num;
        $data['remark'] = $params['remark'] ?? 0;
        $data['status'] = self::STATUS_PENDING;
        $data['operator_uuid'] = $params['operator_id'];
        $model->save($data);
        return $model->id;
    }

    /**
     * 编辑更改
     */
    public function edit($params)
    {
        $detail = (new self)->detail($params['id'] ?? 0);
        if (!$detail) {
            $this->error = '记录不存在';
            return false;
        }
        $productId = $params['product_id'];
        //
        $productBom = ProductBom::where('uuid',  $productId)->find();
        $material = Material::where('uuid',  $productId)->find();

        $productBomUuid = 0;
        $materialUuid = 0;

        $num = $params['num'] ?? 0;
        $stock_num = 0;
        if ($productBom) {
            $stock_num = ProductBom::where('uuid',  $productId)->value('stock_num');
            $productBomUuid =  $productId;
        } else if ($material) {
            $stock_num = Material::where('uuid',  $productId)->value('stock_num');
            $materialUuid =  $productId;
        } else {
            $this->error = '商品不存在';
            return false;
        }

        if ($num > $stock_num) {
            $this->error = '不能大于库存数量';
            return false;
        }
        //
        $updateArr['scene'] = self::OLD_SCENE_LOSS[$params['type'] ?? 1];
        $data['product_bom_uuid'] = $productBomUuid;
        $data['material_uuid'] = $materialUuid;
        $updateArr['num'] = $num;
        $updateArr['remark'] = $params['remark'] ?? '';
        $updateArr['operator_uuid'] = $params['operator_id'];

        return $detail->save($updateArr);
    }

    /**
     * 审核
     */
    public function review($params)
    {
        if ($this->status != 0) {
            $this->error = '当前状态不可操作';
            return false;
        }
        $updateArr = [];
        $this->startTrans();
        try {
            if (($params['review_status'] ?? 0) == 1) {
                $updateArr['status'] = 1;
                $updateArr['approved_time'] = time();
                // 减少库存
                $product = ProductBom::where('uuid', $this->product_bom_uuid)->find();
                $material = Material::where('uuid', $this->material_uuid)->find();
                // 成品
                if ($product) {
                    if ($this->num > $product->stock_num) {
                        $this->error = '报损数量大于剩余库存数量';
                        return false;
                    }
                    $skuStock = helper::bcsub($product->stock_num, $this->num);
                    $product->save(['stock_num' => $skuStock]);
                }
                // 材料
                else if ($material) {
                    if ($this->num > $material->stock_num) {
                        $this->error = '报损数量大于剩余库存数量';
                        return false;
                    }
                    $skuStock = helper::bcsub($material->stock_num, $this->num, 4);
                    $material->save(['stock_num' => $skuStock]);
                    $relatedMaterialUuidList = RelatedMaterial::where('material_uuid', $material->uuid)->column('uuid') ?: [];
                    RelatedMaterial::updateStock($relatedMaterialUuidList);
                } else {
                    $this->error = '记录不存在';
                    return false;
                }
            }
            if (($params['review_status'] ?? 0) == 2) {
                $product = ProductBom::where('uuid', $this->product_bom_uuid)->find();
                $material = Material::where('uuid', $this->material_uuid)->find();
                if (!$product && !$material) {
                    $this->error = '商品不存在';
                    return false;
                }
                $updateArr['status'] = 2;
                $updateArr['revoke_time'] = time();
                $updateArr['reject_reason'] = $params['refused'] ?? '';
            }
            //
            $this->save($updateArr);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 根据规格汇总报损数量
     */
    public function sumDamagedProductNum($product_sku_id)
    {
        return (new self())->where('product_sku_id', $product_sku_id)->sum('num');;
    }

    /**
     * 入库编号：18位纯数字（前2位WT，2-10位是年月日，中间位是0000，后4位随机生成）
     *
     * @return string
     */
    public function generateNumber()
    {
        $date = date('Ymd');
        $rand = rand(1000, 9999);
        $code = 'WT' . $date . '0000' . $rand;
        return $code;
    }
}

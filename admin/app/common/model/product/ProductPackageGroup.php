<?php

namespace app\common\model\product;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\store\MultiLanguageName;
use app\common\model\product\ProductPackageGroupItem as ProductPackageGroupItemModel;

/**
 * 套餐组
 */
class ProductPackageGroup extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_package_group';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $append = ['group_name_text'];

    /**
     * 规格名称
     */
    public static function getGroupNameTextAttr($value, $data)
    {
        return extractLanguage($data['name']);
    }

    /**
     * 关联套餐分组商品
     */
    public function productPackageGropItem()
    {
        return $this->hasMany(ProductPackageGroupItemModel::class, 'product_package_group_uuid', 'uuid');
    }

    /**
     * 关联套餐
     */
    public function product()
    {
        return $this->belongsTo(Product::class, 'product_package_uuid', 'uuid');
    }

    /**
     * 添加套餐商品组
     */
    public static function addPackageGroup($data, $product)
    {
        $insertGroups = [];
        $insertGroupItems = [];

        $packageGroup = $data['package_group'] ?? [];
        foreach ($packageGroup as $item) {
            $groupUuid = createUuid();
            $multiLanguageNameUuid = (new MultiLanguageName)->saveNames($item['group_name']);
            $insertGroups[] = [
                'uuid' => $groupUuid, // 套餐分组uuid
                'name' => $item['group_name'], // 套餐分组名称
                'multi_language_name_uuid' => $multiLanguageNameUuid, // 多语言名称uuid
                'product_package_uuid' => $product['uuid'], // 套餐uuid
                'create_time' => time(), // 创建时间
                'update_time' => time(), // 更新时间
            ];
            $productIds = array_column($item['product_list'], 'product_id');
            $productBoms = ProductBom::whereIn('uuid', $productIds)->column('product_package_uuid', 'uuid');
            foreach ($item['product_list'] as $productItem) {
                $insertGroupItems[] = [
                    'uuid' => createUuid(), // 套餐分组商品uuid
                    'product_package_group_uuid' => $groupUuid, // 套餐分组uuid
                    'related_uuid' => $productBoms[$productItem['product_id']], // product_package_uuid
                    'product_bom_uuid' => $productItem['product_id'], // product_bom_uuid
                    'num' => $productItem['num'] ?: 0, // 商品数量
                    'sort' => $productItem['sort'] ?: 0, // 排序
                    'create_time' => time(), // 创建时间
                    'update_time' => time(), // 更新时间
                ];
            }
        }
        if (!empty($insertGroups)) {
            (new self())->saveAll($insertGroups);
        }
        if (!empty($insertGroupItems)) {
            (new ProductPackageGroupItemModel())->saveAll($insertGroupItems);
        }
    }

    /**
     * 更新套餐分组
     */
    public static function updatePackageGroup($data, $product)
    {
        $groupUuidList = [];
        $groupItemUuidList = [];
        // 新增或编辑套餐分组
        $groupList = $data['package_group'];
        foreach ($groupList as $item) {
            $groupData = [
                'name' => $item['group_name'], // 套餐分组名称
                'product_package_uuid' => $product['uuid'], // 套餐uuid
            ];
            $groupUuid = $item['group_id'] ?? 0;
            if ($groupUuid == 0) {
                $multiLanguageNameUuid = (new MultiLanguageName)->saveNames($item['group_name']);
                $groupData['multi_language_name_uuid'] = $multiLanguageNameUuid;
                /** @var ProductPackageGroup $group */
                $group = self::create($groupData);
            } else {
                $group = self::where('uuid', $groupUuid)->find();
                if (!$group) {
                    $multiLanguageNameUuid = (new MultiLanguageName)->saveNames($item['group_name']);
                    $groupData['multi_language_name_uuid'] = $multiLanguageNameUuid;
                    /** @var ProductPackageGroup $group */
                    $group = self::create($groupData);
                } else {
                    $multiLanguageNameUuid = (new MultiLanguageName)->saveNames($item['group_name'], $group['multi_language_name_uuid']);
                    $groupData['multi_language_name_uuid'] = $multiLanguageNameUuid;
                    /** @var ProductPackageGroup $group */
                    $group->save($groupData);
                }
            }
            $groupItemList = $item['product_list'] ?? [];
            $productIds = array_column($groupItemList, 'product_id');
            $productBoms = ProductBom::whereIn('uuid', $productIds)->column('product_package_uuid', 'uuid');
            foreach ($groupItemList as $item) {
                $itemData = [
                    'product_package_group_uuid' => $group['uuid'], // 套餐分组uuid 
                    'related_uuid' => $productBoms[$item['product_id']], // product_package_uuid
                    'product_bom_uuid' => $item['product_id'], // product_bom_uuid
                    'num' => $item['num'] ?: 0, // 商品数量
                    'sort' => $item['sort'] ?: 0, // 排序
                ];
                $groupItemId = $item['item_id'] ?? 0;
                if ($groupItemId == 0) {
                    $item = ProductPackageGroupItemModel::create($itemData);
                } else {
                    $item = ProductPackageGroupItemModel::where('uuid', $groupItemId)->find();
                    if (!$item) {
                        $item = ProductPackageGroupItemModel::create($itemData);
                    } else {
                        $item->save($itemData);
                    }
                }
                $groupItemUuidList[$group['uuid']][] = $item['uuid'];
            }
            $groupUuidList[] = $group['uuid'];
        }
        // 删除套餐分组
        if (!empty($groupUuidList)) {
            $delGroupUuidList = self::whereNotIn('uuid', $groupUuidList)->where('product_package_uuid', $product['uuid'])->column('uuid');
            self::destroy(function ($query) use ($delGroupUuidList) {
                $query->whereIn('uuid', $delGroupUuidList);
            });
            ProductPackageGroupItemModel::destroy(function ($query) use ($delGroupUuidList) {
                $query->whereIn('product_package_group_uuid', $delGroupUuidList);
            });     
        }
        // 删除套餐分组商品
        if (!empty($groupItemUuidList)) {
            foreach ($groupItemUuidList as $groupUuid => $itemUuidList) {
                ProductPackageGroupItemModel::destroy(function ($query) use ($groupUuid, $itemUuidList) {
                    $query->where('product_package_group_uuid', $groupUuid)->whereNotIn('uuid', $itemUuidList);
                });
            };
        }
    }

    /**
     * 删除套餐分组
     */
    public static function deletePackageGroup($product)
    {
        $groupUuidList = self::where('product_package_uuid', $product['uuid'])->column('uuid');
        self::destroy(function ($query) use ($product) {
            $query->where('product_package_uuid', $product['uuid']);
        });
        ProductPackageGroupItemModel::destroy(function ($query) use ($groupUuidList) {
            $query->whereIn('product_package_group_uuid', $groupUuidList);
        });
    }
}

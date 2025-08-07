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
            foreach ($item['product_list'] as $productItem) {
                $insertGroupItems[] = [
                    'uuid' => createUuid(), // 套餐分组商品uuid
                    'product_package_group_uuid' => $groupUuid, // 套餐分组uuid
                    'related_uuid' => $product['uuid'], // product_package_uuid
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
            foreach ($groupItemList as $item) {
                $itemData = [
                    'product_package_group_uuid' => $group['uuid'],
                    'related_uuid' => $product['uuid'],
                    'product_bom_uuid' => $item['product_id'],
                    'num' => $item['num'] ?: 0,
                    'sort' => $item['sort'] ?: 0,
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
                $groupItemUuidList[] = $item['uuid'];
            }
            $groupUuidList[] = $group['uuid'];
        }
        // 删除套餐分组
        if (!empty($groupUuidList)) {
            self::destroy(function ($query) use ($groupUuidList, $product) {
                $query->whereNotIn('uuid', $groupUuidList)->where('product_package_uuid', $product['uuid']);
            });
            ProductPackageGroupItemModel::destroy(function ($query) use ($groupUuidList, $product) {
                $query->whereNotIn('product_package_group_uuid', $groupUuidList)->where('related_uuid', $product['uuid']);
            });     
        }
        // 删除套餐分组商品
        if (!empty($groupItemUuidList)) {
            ProductPackageGroupItemModel::destroy(function ($query) use ($groupItemUuidList, $product) {
                $query->whereNotIn('uuid', $groupItemUuidList)->where('related_uuid', $product['uuid']);
            });
        }
    }

    /**
     * 删除套餐分组
     */
    public static function deletePackageGroup($product)
    {
        self::destroy(function ($query) use ($product) {
            $query->where('product_package_uuid', $product['uuid']);
        });
        ProductPackageGroupItemModel::destroy(function ($query) use ($product) {
            $query->where('related_uuid', $product['uuid']);
        });
    }
}

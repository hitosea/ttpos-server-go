<template>
  <el-dialog
    class="product-selector"
    @close="handleClose"
    v-model="dialogVisible"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :title="$t('选择商品')"
    append-to-body
  >
    <!-- {{ categoriesTreeCurrentKey }} -->
    <!-- <pre style="font-size: 12px">{{ selectedProductsTmp }}</pre> -->
    <div class="product-selector-content">
      <div class="product-selector-tree">
        <template v-if="props.selectorType === 'label'">
          <el-tree-v2
            ref="printTagsTreeRef"
            :height="480"
            :data="printTagsTree"
            node-key="id"
            highlight-current
            :current-node-key="printTagsTreeCurrentKey"
            @current-change="handlePrintTagsTreeCurrentChange"
            auto-expand-parent
            :expand-on-click-node="false"
            :default-expanded-keys="printTagsTreeExpandedKeys"
            @node-expand="handlePrintTagsTreeExpand"
            @node-collapse="handlePrintTagsTreeExpand"
            :props="{ children: 'children', label: 'label' }"
          >
            <template #default="{ node }">
              <div v-show="false" :style="{ marginRight: '4px' }" @click.stop>
                <el-checkbox
                  v-model="printTagsProductIsAllSelected[`l-${node.key}`]"
                  @change="(v) => handlePrintTagCheck(node.key, v)"
                  :disabled="printTagsProductCount[`l-${node.key}`] === 0"
                />
              </div>
              <span>{{ node.label }}</span>
              <template v-if="printTagsSelectedProductCount[`l-${node.key}`] > 0">
                <span style="margin-left: 2px">({{ printTagsSelectedProductCount[`l-${node.key}`] }})</span>
              </template>
            </template>
          </el-tree-v2>
        </template>
        <template v-else>
          <el-tree-v2
            ref="categoriesTreeRef"
            :height="480"
            :data="categoriesTree"
            node-key="id"
            highlight-current
            :current-node-key="categoriesTreeCurrentKey"
            @current-change="handleCategoriesTreeCurrentChange"
            auto-expand-parent
            :expand-on-click-node="false"
            :default-expanded-keys="categoriesTreeExpandedKeys"
            @node-expand="handleCategoriesTreeExpand"
            @node-collapse="handleCategoriesTreeExpand"
            :props="{ children: 'children', label: 'label' }"
          >
            <template #default="{ node }">
              <div v-show="false" :style="{ marginRight: '4px' }" @click.stop>
                <el-checkbox
                  v-model="categoriesProductIsAllSelected[`c-${node.key}`]"
                  @change="(v) => handleCategoryCheck(node.key, v)"
                  :disabled="categoriesProductCount[`c-${node.key}`] === 0"
                />
              </div>
              <span>{{ node.label }}</span>
              <template v-if="categoriesProductSelectedCount[`c-${node.key}`] > 0">
                <span style="margin-left: 2px">({{ categoriesProductSelectedCount[`c-${node.key}`] }})</span>
              </template>
            </template>
          </el-tree-v2>
        </template>
      </div>
      <!-- <div class="product-selector-divider" /> -->
      <div class="product-selector-main">
        <div class="product-selector-form">
          <el-form size="small" ref="formRef" :model="form" :inline="true">
            <el-form-item :label="$t('商品名称')" :placeholder="$t('请输入商品名称')">
              <el-input size="small" v-model="form.product_name" @input="onDebounceSearch" />
            </el-form-item>
            <el-form-item>
              <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
                {{ $t('查询') }}
              </el-button>
            </el-form-item>
          </el-form>
        </div>

        <div class="product-selector-table">
          <el-auto-resizer>
            <template #default="{ height, width }">
              <el-table
                ref="productsTableRef"
                :height="height"
                :style="{ width }"
                fixed
                :data="productsTableData"
                size="small"
                border
                v-loading="loading"
                :row-class-name="rowClassName"
                @select="handleSelect"
                @select-all="handleSelectAll"
              >
                <!-- 添加 selectable 属性控制是否可选 -->
                <el-table-column type="selection" width="40" :selectable="selectable" />
                <el-table-column prop="product_name_text" :label="$t('商品名称')" />
                <el-table-column v-if="props.haveSku" prop="spec_name_text" :label="$t('规格')" />
                <el-table-column v-if="props.haveSku" prop="stock_num" :label="$t('库存')" />
              </el-table>
            </template>
          </el-auto-resizer>
        </div>
      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="onSubmit" :loading="loading || props.isLoading">{{ $t('确定') }}</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
  import { ref, reactive, computed, nextTick, getCurrentInstance } from 'vue';
  import IndexApi from '@/api/index.js';
  import InventoryApi from '@/api/inventory.js';

  defineOptions({
    name: 'ProductSelector',
  });

  const { proxy } = getCurrentInstance();

  const emit = defineEmits(['close']);

  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
    selectorType: {
      type: String,
      default: 'all',
      validator(value) {
        return ['all', 'category', 'label'].includes(value);
      },
    },
    type: {
      type: String,
      default: '',
    },
    // 计量类型：0-所有；1-整数；2-小数
    numType: {
      type: Number,
      default: 0,
    },
    showDeliveryRequired: {
      type: Number,
      default: 0,
    },
    selectedProductIds: {
      type: Array,
      default: () => [],
    },
    maxCount: {
      type: Number,
      default: Infinity,
    },
    haveSku: {
      type: Boolean,
      default: false,
    },
    isLoading: {
      type: Boolean,
      default: false,
    },
    // 库存为0的商品不可选
    stockZero: {
      type: Boolean,
      default: false,
    },
    haveStatusZero: {
      type: Boolean,
      default: false,
    },
  });

  const dialogVisible = ref(props.open);

  const loading = ref(false);

  const form = reactive({
    product_name: '',
  });

  const formRef = ref(null);

  const categories = ref([]);
  const categoriesProductCount = computed(() => {
    const _map = {};

    let count = 0;
    for (const item of categories.value) {
      const _count = products.value.filter((product) => product.category_id === item.category_id).length;
      let childCount = 0;
      for (const child of item.child) {
        const _childCount = products.value.filter((product) => product.category_id === child.category_id).length;
        _map[`c-${child.category_id}`] = _childCount;
        childCount += _childCount;
      }
      _map[`c-${item.category_id}`] = _count + childCount;
      count += _count + childCount;
    }

    _map[`c-0`] = count;
    return _map;
  });

  const categoriesTree = computed(() => {
    return [
      {
        id: 0,
        label: $t('全部'),
        children: loading.value
          ? undefined
          : Array.isArray(categories.value)
          ? categories.value.map((item) => {
              return {
                id: item.category_id,
                pid: 0,
                label: item.name_text,
                children: Array.isArray(item.child)
                  ? item.child.map((child) => {
                      return {
                        id: child.category_id,
                        pid: item.category_id,
                        label: child.name_text,
                      };
                    })
                  : undefined,
              };
            })
          : undefined,
      },
    ];
  });
  const categoriesTreeRef = ref(null);
  const categoriesTreeCurrentKey = ref(0);
  const categoriesTreeExpandedKeys = ref([0]);

  const printTags = ref([]);
  const printTagsProductCount = computed(() => {
    const _map = {};

    let count = 0;
    for (const item of printTags.value) {
      const _count = products.value.filter((product) => product.label_id === item.label_id).length;
      _map[`l-${item.label_id}`] = _count;
      count += _count;
    }

    _map[`l-0`] = count;
    return _map;
  });
  const printTagsTree = computed(() => {
    return [
      {
        id: 0,
        label: $t('全部'),
        children: loading.value
          ? undefined
          : Array.isArray(printTags.value)
          ? printTags.value.map((item) => {
              return {
                id: item.label_id,
                pid: 0,
                label: item.label_name_text,
              };
            })
          : undefined,
      },
    ];
  });
  const printTagsTreeRef = ref(null);
  const printTagsTreeCurrentKey = ref(0);
  const printTagsTreeExpandedKeys = ref([0]);

  const products = ref([]);
  const productsTableData = computed(() => {
    switch (props.selectorType) {
      case 'category':
        return products.value
          .filter((item) => {
            // 如果选择的是全部分类，显示所有商品
            if (categoriesTreeCurrentKey.value === 0) return true;

            // 直接匹配当前分类ID
            if (categoriesTreeCurrentKey.value === item.category_id) return true;

            // 检查是否是选中分类的子分类商品
            if (categoriesTreeCurrentKey.value === item.parent_category_id) return true;

            // 查找选中分类的所有子分类ID，判断商品是否属于这些子分类
            const selectedCategory = categories.value.find((cat) => cat.category_id === categoriesTreeCurrentKey.value);
            if (selectedCategory && selectedCategory.child) {
              return selectedCategory.child.some((child) => child.category_id === item.category_id);
            }

            return false;
          })
          .filter((item) => !searchValue.value || item.product_name_text.includes(searchValue.value));
      case 'label':
        return products.value
          .filter((item) => printTagsTreeCurrentKey.value === 0 || printTagsTreeCurrentKey.value === item.label_id)
          .filter((item) => !searchValue.value || item.product_name_text.includes(searchValue.value));
      default:
        return products.value
          .filter((item) => {
            // 如果选择的是全部分类，显示所有商品
            if (categoriesTreeCurrentKey.value === 0) return true;

            // 直接匹配当前分类ID
            if (categoriesTreeCurrentKey.value === item.category_id) return true;

            // 检查是否是选中分类的子分类商品
            if (categoriesTreeCurrentKey.value === item.parent_category_id) return true;

            // 查找选中分类的所有子分类ID，判断商品是否属于这些子分类
            const selectedCategory = categories.value.find((cat) => cat.category_id === categoriesTreeCurrentKey.value);
            if (selectedCategory && selectedCategory.child) {
              return selectedCategory.child.some((child) => child.category_id === item.category_id);
            }

            return false;
          })
          .filter((item) => !searchValue.value || item.product_name_text.includes(searchValue.value));
    }
  });

  const selectedProductsTmp = ref([]);

  const categoriesProductSelectedCount = computed(() => {
    const _map = {};

    let count = 0;
    for (const item of categories.value) {
      const _count = selectedProductsTmp.value.filter((product) => product.category_id === item.category_id).length;
      let childCount = 0;
      for (const child of item.child) {
        const _childCount = selectedProductsTmp.value.filter((product) => product.category_id === child.category_id).length;
        _map[`c-${child.category_id}`] = _childCount;
        childCount += _childCount;
      }
      _map[`c-${item.category_id}`] = _count + childCount;
      count += _count + childCount;
    }

    _map[`c-0`] = count;
    return _map;
  });
  const categoriesProductIsAllSelected = computed({
    get() {
      const _map = {};

      for (const item of categories.value) {
        _map[`c-${item.category_id}`] =
          categoriesProductCount.value[`c-${item.category_id}`] > 0 &&
          categoriesProductSelectedCount.value[`c-${item.category_id}`] === categoriesProductCount.value[`c-${item.category_id}`];
        for (const child of item.child) {
          _map[`c-${child.category_id}`] =
            categoriesProductCount.value[`c-${child.category_id}`] > 0 &&
            categoriesProductSelectedCount.value[`c-${child.category_id}`] === categoriesProductCount.value[`c-${child.category_id}`];
        }
      }

      return _map;
    },
    set(val) {
      console.log(val);
    },
  });

  const printTagsSelectedProductCount = computed(() => {
    const _map = {};

    let count = 0;
    for (const item of printTags.value) {
      const _count = selectedProductsTmp.value.filter((product) => product.label_id === item.label_id).length;
      _map[`l-${item.label_id}`] = _count;
      count += _count;
    }

    _map[`l-0`] = count;
    return _map;
  });

  const printTagsProductIsAllSelected = computed(() => {
    const _map = {};

    for (const item of printTags.value) {
      _map[`l-${item.label_id}`] =
        printTagsProductCount.value[`l-${item.label_id}`] > 0 && printTagsSelectedProductCount.value[`l-${item.label_id}`] === printTagsProductCount.value[`l-${item.label_id}`];
    }

    return _map;
  });

  const pagination = ref({
    page: 1,
    pageSize: 10000,
    total: 0,
    totalPage: 0,
  });

  const getData = async (isFirst = false) => {
    loading.value = true;
    try {
      const product_name = form?.product_name;
      const category_ids = categoriesTreeCurrentKey.value === 0 ? '' : categoriesTreeCurrentKey.value;
      const label_ids = printTagsTreeCurrentKey.value === 0 ? '' : printTagsTreeCurrentKey.value;
      const res = props.haveSku
        ? // 有sku
          await InventoryApi.getErpInventory(
            {
              material_type: 10,
              product_status: props.haveStatusZero ? 0 : 10,
              filter_having_material: 0,
              filter_having_decimal: 1,
              list_rows: 1000,
            },
            true
          )
        : // 无sku
          await IndexApi.getProductList(
            {
              page: pagination.value.page,
              list_rows: pagination.value.pageSize,
              mode: props.selectorType,
              product_name,
              category_ids,
              label_ids,
              type: props.type,
              num_type: props.numType,
              show_delivery_required: props.showDeliveryRequired,
              show_package: 1,
            },
            true
          );

      products.value = res.data.list.data;
      if (isFirst) {
        categories.value = res.data.category;
        printTags.value = res.data.label;
      }

      pagination.value.total = res.data.list.total;
      pagination.value.totalPage = res.data.list.total_page;
      pagination.value.page = res.data.list.current_page;
      pagination.value.pageSize = res.data.list.per_page;

      toggleRowSelection(isFirst);
    } catch (error) {
      //
    } finally {
      loading.value = false;
    }
  };

  getData(true);

  const onSubmit = () => {
    // 二次校验：当开启库存为0不可选时，确认前剔除库存不足的商品并提示
    if (props.stockZero) {
      const outOfStocks = selectedProductsTmp.value.filter((item) => Number(item?.stock_num ?? 0) <= 0);
      if (outOfStocks.length > 0) {
        // 取消这些行的选中并从已选中列表移除
        nextTick(() => {
          outOfStocks.forEach((row) => productsTableRef.value?.toggleRowSelection(row, false));
        });
        selectedProductsTmp.value = selectedProductsTmp.value.filter((item) => Number(item?.stock_num ?? 0) > 0);
        proxy.$ElMessage({ message: $t('商品库存不足，请调整。'), type: 'warning' });
        return; // 中断提交，等待用户调整
      }
    }

    emit('close', selectedProductsTmp.value, categories.value);
    reset();
  };

  const reset = () => {
    selectedProductsTmp.value = [];
  };

  const handleClose = () => {
    emit('close');
    reset();
  };

  const productsTableRef = ref(null);

  // 添加选择控制函数
  const selectable = (row) => {
    // 当开启库存为0不可选时，库存为0的商品禁止选择
    if (props.stockZero && Number(row?.stock_num ?? 0) <= 0) return false;
    // 如果当前行已选中，允许操作（保持行为与表格一致）
    if (selectedProductsTmp.value.some((item) => item.product_id === row.product_id)) return true;
    // 未选中时检查是否达到最大数量
    return selectedProductsTmp.value.length < props.maxCount;
  };

  // 行样式：库存为0置灰
  const rowClassName = ({ row }) => {
    if (props.stockZero && Number(row?.stock_num ?? 0) <= 0) return 'is-out-of-stock';
    return '';
  };

  const handleSelect = (data, node) => {
    const isChecked = data.some((item) => item.product_id === node.product_id);

    if (isChecked) {
      // 检查是否超过最大数量
      if (selectedProductsTmp.value.length >= props.maxCount) {
        // 超过则取消选中并提示
        productsTableRef.value.toggleRowSelection(node, false);
        proxy.$ElMessage({
          message: $t('最多只能选择' + props.maxCount + $t('个商品')),
          type: 'warning',
        });
        return;
      }

      // 未超过则添加
      if (selectedProductsTmp.value.every((item) => item.product_id !== node.product_id)) {
        selectedProductsTmp.value.push(node);
      }
    } else {
      // 取消选中
      selectedProductsTmp.value = selectedProductsTmp.value.filter((item) => item.product_id !== node.product_id);
    }
  };

  const handleSelectAll = (selection) => {
    if (selection.length === 0) {
      // 取消全选
      selectedProductsTmp.value = selectedProductsTmp.value.filter((item) => !productsTableData.value.some((p) => p.product_id === item.product_id));
      return;
    }

    // 计算可添加的数量
    const currentSelectedCount = selectedProductsTmp.value.length;
    const canSelectCount = props.maxCount - currentSelectedCount;

    if (canSelectCount <= 0) {
      // 已满则提示
      proxy.$ElMessage({
        message: $t('最多只能选择' + props.maxCount + $t('个商品')),
        type: 'warning',
      });
      // 清除全选状态
      nextTick(() => productsTableRef.value.clearSelection());
      return;
    }

    // 获取当前页未选中的商品
    const currentPageUnselected = productsTableData.value
      .filter((item) => !selectedProductsTmp.value.some((p) => p.product_id === item.product_id))
      .filter((item) => !(props.stockZero && Number(item?.stock_num ?? 0) <= 0));

    // 实际可添加的商品
    const toSelect = currentPageUnselected.slice(0, canSelectCount);

    // 添加选中的商品
    toSelect.forEach((item) => {
      if (!selectedProductsTmp.value.some((p) => p.product_id === item.product_id)) {
        selectedProductsTmp.value.push(item);
      }
    });

    // 如果实际添加数量小于当前页未选数量，需要调整选中状态
    if (toSelect.length < currentPageUnselected.length) {
      nextTick(() => {
        // 清除全选状态
        productsTableRef.value.clearSelection();
        // 重新设置选中状态
        productsTableData.value.forEach((row) => {
          if (selectedProductsTmp.value.some((item) => item.product_id === row.product_id)) {
            productsTableRef.value.toggleRowSelection(row, true);
          }
        });
      });
      proxy.$ElMessage({
        message: $t('最多只能选择' + props.maxCount + $t('个商品')),
        type: 'warning',
      });
    }
  };

  const toggleRowSelection = async (isFirst = false) => {
    if (isFirst) {
      selectedProductsTmp.value = productsTableData.value
        .filter((item) => props.selectedProductIds.includes(item.product_id))
        .filter((item) => !(props.stockZero && Number(item?.stock_num ?? 0) <= 0));
    }

    nextTick(() => {
      // 清除所有选择
      productsTableRef.value.clearSelection();

      // 只选中当前分类下的商品
      const currentProducts = productsTableData.value;
      selectedProductsTmp.value.forEach((item) => {
        if (currentProducts.some((p) => p.product_id === item.product_id)) {
          productsTableRef.value.toggleRowSelection(item, true, true);
        }
      });
    });
  };

  let searchTimer;
  const onDebounceSearch = () => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
      onSearch();
    }, 200);
  };

  const searchValue = ref('');

  const onSearch = () => {
    searchValue.value = form.product_name;

    nextTick(() => {
      toggleRowSelection(false);
    });
  };

  const handleCategoriesTreeCurrentChange = async ({ id }) => {
    if (id === categoriesTreeCurrentKey.value) return;
    categoriesTreeCurrentKey.value = id;
    categoriesTreeRef.value.setCurrentKey(categoriesTreeCurrentKey.value);

    // 清除当前分类下的选择
    productsTableRef.value.clearSelection();

    // 重新设置选中状态
    toggleRowSelection(false);
  };

  const handleCategoryCheck = (id, checked) => {
    const _products = products.value.filter((item) => item.category_id === id || item.parent_category_id === id);

    if (id === categoriesTreeCurrentKey.value) {
      _products.forEach((item) => {
        productsTableRef.value.toggleRowSelection(item, checked, true);
      });
    }

    if (checked) {
      _products.forEach((item) => {
        // 库存为0不可加入已选
        if (props.stockZero && Number(item?.stock_num ?? 0) <= 0) return;
        if (selectedProductsTmp.value.some((product) => product.product_id === item.product_id)) return;
        selectedProductsTmp.value.push(item);
      });
    } else {
      selectedProductsTmp.value = selectedProductsTmp.value.filter((item) => !_products.some((product) => product.product_id === item.product_id));
    }
  };

  const handleCategoriesTreeExpand = ({ id }, { expanded }) => {
    if (expanded) {
      categoriesTreeExpandedKeys.value = [...categoriesTreeExpandedKeys.value, id];
    } else {
      if (id === 0) {
        categoriesTreeExpandedKeys.value = [];
        return;
      }
      categoriesTreeExpandedKeys.value = categoriesTreeExpandedKeys.value.filter((item) => item !== id);
    }
  };

  const getCategoryPid = (id) => {
    let pid;

    const data = categoriesTree.value?.[0]?.children;
    if (!data) return pid;

    for (const treeItem of data) {
      const _pid = treeItem.children?.find((v) => v.id === id)?.pid;
      if (typeof _pid === 'number') {
        pid = _pid;
        break;
      }
      for (const treeItemChild of treeItem.children) {
        const _pid = treeItemChild.children?.find((v) => v.id === id)?.pid;
        if (typeof _pid === 'number') {
          pid = _pid;
          break;
        }
      }
    }

    return pid;
  };

  const handleCategoryTableSelectAll = (val) => {
    if (categoriesTreeCurrentKey.value === 0) {
      selectedProductsTmp.value = val;
      toggleRowSelection(false);
      return;
    }

    const pid = getCategoryPid(categoriesTreeCurrentKey.value);

    const filterCondition = pid
      ? (product) => product.category_id !== categoriesTreeCurrentKey.value
      : (product) => product.category_id !== categoriesTreeCurrentKey.value && product.parent_category_id !== categoriesTreeCurrentKey.value;

    const selectedProductsWithoutCurrent = selectedProductsTmp.value.filter(filterCondition);
    selectedProductsTmp.value = [...new Set([...selectedProductsWithoutCurrent, ...val].map((item) => item.product_id))].map((id) =>
      [...selectedProductsWithoutCurrent, ...val].find((item) => item.product_id === id)
    );

    toggleRowSelection(false);
  };

  const handlePrintTagsTreeCurrentChange = async ({ id }) => {
    if (id === printTagsTreeCurrentKey.value) return;
    printTagsTreeCurrentKey.value = id;
    printTagsTreeRef.value.setCurrentKey(printTagsTreeCurrentKey.value);

    toggleRowSelection(false);
  };

  const handlePrintTagCheck = (id, checked) => {
    const _products = products.value.filter((item) => item.label_id === id);

    if (id === printTagsTreeCurrentKey.value) {
      _products.forEach((item) => {
        productsTableRef.value.toggleRowSelection(item, checked, true);
      });
    }

    if (checked) {
      _products.forEach((item) => {
        // 库存为0不可加入已选
        if (props.stockZero && Number(item?.stock_num ?? 0) <= 0) return;
        if (selectedProductsTmp.value.some((product) => product.product_id === item.product_id)) return;
        selectedProductsTmp.value.push(item);
      });
    } else {
      selectedProductsTmp.value = selectedProductsTmp.value.filter((item) => !_products.some((product) => product.product_id === item.product_id));
    }
  };

  const handlePrintTagsTreeExpand = ({ id }, { expanded }) => {
    if (expanded) {
      printTagsTreeExpandedKeys.value = [...printTagsTreeExpandedKeys.value, id];
    } else {
      if (id === 0) {
        printTagsTreeExpandedKeys.value = [];
        return;
      }
      printTagsTreeExpandedKeys.value = printTagsTreeExpandedKeys.value.filter((item) => item !== id);
    }
  };

  const getPrintTagPid = (id) => {
    let pid;

    const data = printTagsTree.value?.[0]?.children;
    if (!data) return pid;

    for (const treeItem of data) {
      const _pid = treeItem.children?.find((v) => v.id === id)?.pid;
      if (typeof _pid === 'number') {
        pid = _pid;
        break;
      }
    }

    return pid;
  };

  const handlePrintTagsTableSelectAll = (val) => {
    if (printTagsTreeCurrentKey.value === 0) {
      selectedProductsTmp.value = val;
      toggleRowSelection(false);
      return;
    }

    const pid = getPrintTagPid(printTagsTreeCurrentKey.value);
    const filterCondition = pid ? (product) => product.label_id !== printTagsTreeCurrentKey.value : (product) => product.label_id !== printTagsTreeCurrentKey.value;

    const selectedProductsWithoutCurrent = selectedProductsTmp.value.filter(filterCondition);
    selectedProductsTmp.value = [...new Set([...selectedProductsWithoutCurrent, ...val].map((item) => item.product_id))].map((id) =>
      [...selectedProductsWithoutCurrent, ...val].find((item) => item.product_id === id)
    );

    toggleRowSelection(false);
  };
</script>

<style lang="scss" scoped>
  .product-selector-content {
    display: flex;
    justify-content: flex-start;
    align-items: stretch;
    gap: 8px;

    .product-selector-tree {
      width: 240px;
      flex-shrink: 0;
      overflow-x: hidden;
    }

    .product-selector-divider {
      margin: 0 4px;
      width: 2px;
      flex-shrink: 0;
      background-color: #f0f2f5;
    }

    .product-selector-main {
      flex-grow: 1;
      display: flex;
      overflow: hidden;
      flex-direction: column;
      gap: 8px;

      .product-selector-form {
        flex-shrink: 0;
      }

      .product-selector-table {
        flex-grow: 1;
      }
    }
  }
  // 库存为0置灰样式
  :deep(.el-table__row.is-out-of-stock) {
    color: #c0c4cc;
    pointer-events: none;
  }
</style>

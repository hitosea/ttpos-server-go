<template>
  <div class="batch-discount">
    <div class="common-return">
      <el-icon class="return-icon" @click="dialogFormVisible"> <ArrowLeftBold /> </el-icon>{{ $t('批量修改整单折扣商品') }}
    </div>

    <div class="product-selector-content">
      <div class="product-selector-tree">
        <el-auto-resizer>
          <template #default="{ height, width }">
            <el-tree-v2
              ref="categoriesTreeRef"
              :height="height"
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
        </el-auto-resizer>
      </div>
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
                @select="handleSelect"
                @select-all="handleSelectAll"
              >
                <!-- 添加 selectable 属性控制是否可选 -->
                <el-table-column type="selection" width="40" :selectable="selectable" />
                <el-table-column prop="product_name_text" :label="$t('商品名称')" />
              </el-table>
            </template>
          </el-auto-resizer>
        </div>
      </div>
    </div>

    <div class="common-button-wrapper">
      <div>
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="handleClick" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, reactive, computed, nextTick, getCurrentInstance } from 'vue';
  import IndexApi from '@/api/index.js';
  import ProductApi from '@/api/product.js';
  defineOptions({
    name: 'ProductSelector',
  });

  const { proxy } = getCurrentInstance();

  const props = defineProps({
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
    maxCount: {
      type: Number,
      default: Infinity,
    },
  });

  const loading = ref(false);

  const form = reactive({
    product_name: '',
  });

  const formRef = ref(null);

  const selectedProductIds = ref([]);

  const categories = ref([]);

  // 分类商品数量
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

  // 分类树
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

  // 分类树引用
  const categoriesTreeRef = ref(null);
  // 分类树当前选中
  const categoriesTreeCurrentKey = ref(0);
  // 分类树展开
  const categoriesTreeExpandedKeys = ref([0]);

  // 标签
  const printTags = ref([]);
  // 标签树当前选中
  const printTagsTreeCurrentKey = ref(0);

  // 商品列表
  const products = ref([]);
  const productsTableData = computed(() => {
    return products.value
      .filter((item) => categoriesTreeCurrentKey.value === 0 || categoriesTreeCurrentKey.value === item.category_id || categoriesTreeCurrentKey.value === item.parent_category_id)
      .filter((item) => !searchValue.value || item.product_name_text.includes(searchValue.value));
  });

  // 已选商品临时列表
  const selectedProductsTmp = ref([]);

  // 分类商品已选数量
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

  // 分类商品是否全选
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

  // 分页
  const pagination = ref({
    page: 1,
    pageSize: 10000,
    total: 0,
    totalPage: 0,
  });

  // 获取数据
  const getData = async (isFirst = false) => {
    loading.value = true;
    try {
      const product_name = form?.product_name;
      const category_ids = categoriesTreeCurrentKey.value === 0 ? '' : categoriesTreeCurrentKey.value;
      const label_ids = printTagsTreeCurrentKey.value === 0 ? '' : printTagsTreeCurrentKey.value;
      const res = await IndexApi.getProductList(
        {
          page: pagination.value.page,
          list_rows: pagination.value.pageSize,
          mode: 'all',
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

      // 等待设置已选商品完成后再执行选择操作
      await setSelectedProducts(res.data.list.data);
      toggleRowSelection(isFirst);
    } catch (error) {
      //
    } finally {
      loading.value = false;
    }
  };

  getData(true);

  // 商品表格引用
  const productsTableRef = ref(null);

  // 选择控制函数
  const selectable = (row) => {
    // 如果当前行已选中，始终可操作（允许取消）
    if (selectedProductsTmp.value.some((item) => item.product_id === row.product_id)) {
      return true;
    }
    // 未选中时检查是否达到最大数量
    return selectedProductsTmp.value.length < props.maxCount;
  };

  // 选择商品
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

  // 全选商品
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
    const currentPageUnselected = productsTableData.value.filter((item) => !selectedProductsTmp.value.some((p) => p.product_id === item.product_id));

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

  /**
   * 设置已选商品ID列表
   * @param {Array} productList 商品列表数据
   */
  const setSelectedProducts = async (productList) => {
    // 筛选出open_overall_discount为1的商品ID
    selectedProductIds.value = productList.filter((item) => item.open_overall_discount === 1).map((item) => item.product_id);

    // 确保数据处理完毕
    await nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
  };

  const toggleRowSelection = async (isFirst = false) => {
    if (isFirst) {
      selectedProductsTmp.value = productsTableData.value.filter((item) => selectedProductIds.value.includes(item.product_id));
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

  // 分类商品勾选
  const handleCategoryCheck = (id, checked) => {
    const _products = products.value.filter((item) => item.category_id === id || item.parent_category_id === id);

    if (id === categoriesTreeCurrentKey.value) {
      _products.forEach((item) => {
        productsTableRef.value.toggleRowSelection(item, checked, true);
      });
    }

    if (checked) {
      _products.forEach((item) => {
        if (selectedProductsTmp.value.some((product) => product.product_id === item.product_id)) return;
        selectedProductsTmp.value.push(item);
      });
    } else {
      selectedProductsTmp.value = selectedProductsTmp.value.filter((item) => !_products.some((product) => product.product_id === item.product_id));
    }
  };

  // 分类树展开
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

  // 关闭弹窗
  const dialogFormVisible = () => {
    proxy.$router.go(-1);
  };

  // 确定按钮
  const handleClick = () => {
    // 筛选出未被勾选的商品
    const unselectedProducts = products.value.filter((item) => !selectedProductsTmp.value.some((p) => p.product_id === item.product_id));
    // 把未被筛选的商品存入 product_ids
    const product_ids = unselectedProducts.map((item) => item.product_id);
    // 调用接口
    ProductApi.batchUpdateOverallDiscount({ product_ids }, true)
      .then((res) => {
        proxy.$ElMessage({
          message: $t('操作成功'),
          type: 'success',
        });
        dialogFormVisible();
      })
      .catch((err) => {
        proxy.$ElMessage({
          message: err.message,
          type: 'error',
        });
      });
  };
</script>

<style lang="scss" scoped>
  .product-selector-content {
    display: flex;
    justify-content: flex-start;
    align-items: stretch;
    gap: 8px;
    flex: 1;

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

  .batch-discount {
    display: flex;
    flex-direction: column;
    padding: 16px;
    background-color: #fff;
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    bottom: 0;
    overflow: hidden;
  }
  .dialog-content {
    flex: 1 1 auto;
    overflow: auto;
  }
  .common-return {
    font-size: 20px;
    margin-bottom: 16px;
    padding-bottom: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 700;

    border-bottom: solid 1px var(--el-border-color);
    .return-icon {
      cursor: pointer;
    }
  }
  .common-button-wrapper {
    flex: 0 0 auto;
    flex-shrink: 0;
    justify-content: center;
  }
</style>

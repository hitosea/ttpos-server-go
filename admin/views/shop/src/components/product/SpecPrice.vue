<template>
  <el-dialog class="product-selector" @close="handleClose" v-model="dialogVisible" :close-on-click-modal="false" :close-on-press-escape="false" :title="title" append-to-body>
    <div class="product-selector-content">
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
        <el-form class="product-selector-table" ref="tableFormRef" :model="tableForm">
          <el-auto-resizer>
            <template #default="{ height, width }">
              <el-table
                ref="productsTableRef"
                :height="height"
                :data="tableForm.products.filter((product) => !searchValue || product.product.name.includes(searchValue))"
                size="small"
                border
                :style="{ width: width }"
                v-loading="loading"
              >
                <el-table-column prop="product.name" :label="$t('商品名称')" />
                <el-table-column prop="category.name" :label="$t('分类')" width="200" />
                <el-table-column prop="spec.price" :label="$t('规格价格')" width="100">
                  <template #default="scope">
                    <el-form-item :prop="`tableForm.products[${scope.$index}].spec.price`" size="small" class="price-input">
                      <numInput v-model="scope.row.spec.price" :min="0" :max="100000000" :precision="2" />
                    </el-form-item>
                  </template>
                </el-table-column>
              </el-table>
            </template>
          </el-auto-resizer>
        </el-form>
      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
  import { ref, reactive, watch, getCurrentInstance } from 'vue';
  import ProductApi from '@/api/product.js';

  const { proxy } = getCurrentInstance();

  const emit = defineEmits(['close']);

  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: window.$t('规格价格'),
    },
    specId: {
      type: Number,
      required: true,
      validator(value) {
        return Number.isInteger(value);
      },
    },
  });

  const dialogVisible = ref(props.open);

  const loading = ref(false);

  const products = ref([]);

  const form = reactive({
    product_name: '',
  });

  const formRef = ref(null);

  const tableForm = reactive({
    spec_id: props.specId,
    products: [],
  });

  const tableFormRef = ref(null);

  watch(products, () => {
    updateTableFormProducts();
  });

  const updateTableFormProducts = () => {
    tableForm.products = (products.value ?? []).map((product) => {
      return {
        product: {
          name: product.product_name_text,
          id: product.product_id,
        },
        category: {
          name: product.category_name_text,
          id: product.category_id,
        },
        spec: {
          price: product.product_sku_price ?? '0.00',
          id: product.product_sku_id,
        },
      };
    });
  };

  const getData = () => {
    loading.value = true;
    ProductApi.getRelatedProductBySpec({
      // page: pagination.value.page,
      // list_rows: pagination.value.pageSize,
      // material_type: 10,
      // product_name,
      spec_id: props.specId,
    })
      .then((res) => {
        products.value = res.data;
      })
      .finally(() => {
        loading.value = false;
      });
  };

  getData();

  const onSubmit = () => {
    tableFormRef.value
      .validate()
      .then((valid) => {
        if (!valid) return;
        ProductApi.batchProductSpecPrice(
          {
            spec_id: tableForm.spec_id,
            products: tableForm.products.map((product) => ({
              product_id: product.product.id,
              product_price: product.spec.price,
              product_sku_id: product.spec.id,
            })),
          },
          false
        )
          .then((res) => {
            proxy.$ElMessage({
              message: $t('保存成功'),
              type: 'success',
            });
            handleClose();
          })
          .catch();
      })
      .catch((err) => {
        console.error(err);
      });
  };

  const reset = () => {
    form.product_name = '';
  };

  const handleClose = () => {
    emit('close');
    reset();
  };

  const productsTableRef = ref(null);

  const searchValue = ref('');

  let searchTimer;
  const onDebounceSearch = () => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
      onSearch();
    }, 200);
  };

  const onSearch = () => {
    searchValue.value = form.product_name;
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

    .product-selector-form {
      flex-grow: 1;
    }

    .product-selector-table {
      height: 480px;
    }
  }

  .price-input {
    margin-bottom: 0 !important;
  }
</style>

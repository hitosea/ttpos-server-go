<template>
  <div class="product-batch" v-loading="loading">
    <div class="common-return">
      <el-icon class="return-icon" @click="returnBack"> <ArrowLeftBold /> </el-icon>{{ title }}
    </div>
    <div class="product-body">
      <div class="common-form">
        {{ $t('选择商品') }}
      </div>
      <el-form size="small" ref="formRef" class="product-form" :model="form" label-position="top">
        <el-form-item
          :label="$t('商品')"
          for="no_click"
          prop="product_ids"
          :rules="[
            {
              validator: () => {
                return form.product_ids.length > 0 ? true : false;
              },
              required: true,
              message: $t('请选择商品'),
            },
          ]"
        >
          <div class="product-body">
            <div>
              <el-button size="small" type="primary" @click="addProduct" :loading="loading">
                {{ $t('选择商品') }}（{{ $t('已选商品') }}{{ product_list.length }}{{ $t('个') }}）
              </el-button>
            </div>
            <div class="product-list" v-if="product_list.length > 0">
              <template v-for="(item, index) in product_list" :key="index">
                <div class="product-item">
                  {{ item.product_name_text }}
                  <i class="el-icon el-tag__close" @click="removeProduct(index)">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">
                      <path
                        fill="currentColor"
                        d="M764.288 214.592 512 466.88 259.712 214.592a31.936 31.936 0 0 0-45.12 45.12L466.752 512 214.528 764.224a31.936 31.936 0 1 0 45.12 45.184L512 557.184l252.288 252.288a31.936 31.936 0 0 0 45.12-45.12L557.12 512.064l252.288-252.352a31.936 31.936 0 1 0-45.12-45.184z"
                      ></path>
                    </svg>
                  </i>
                </div>
              </template>
            </div>
          </div>
        </el-form-item>
        <!-- 上传图片 -->
        <upImages
          v-if="type == 1 && product_list.length > 0"
          @close="dialogFormVisible"
          @loading="
            (e) => {
              loading = e;
            }
          "
          ref="upImagesRef"
          :product_list="product_list || []"
        ></upImages>
        <!-- 修改分类 -->
        <typeChange
          v-if="type == 2 && product_list.length > 0"
          @close="dialogFormVisible"
          @loading="
            (e) => {
              loading = e;
            }
          "
          ref="typeChangeRef"
          :product_list="product_list || []"
        ></typeChange>
        <!-- 修改税类 -->
        <taxChange
          v-if="type == 3 && product_list.length > 0"
          @close="dialogFormVisible"
          @loading="
            (e) => {
              loading = e;
            }
          "
          ref="taxChangeRef"
          :product_list="product_list || []"
        ></taxChange>
      </el-form>
    </div>

    <div class="common-button-wrapper" v-if="product_list.length > 0">
      <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
      <el-button type="primary" @click="handleClick" :loading="loading">{{ $t('确定') }}</el-button>
    </div>

    <!-- 商品选择器 -->
    <ProductSelector
      v-if="openProductSelector"
      :open="openProductSelector"
      @close="handleProductSelectorClose"
      selectorType="all"
      type="all"
      :selectedProductIds="model?.product_ids?.map((item) => item.product_id) ?? []"
      :hasPackage="true"
    >
    </ProductSelector>
  </div>
</template>

<script setup>
  import { ref, reactive, provide, watch, onMounted } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import ProductSelector from '@/components/product/Selector.vue';
  import upImages from './components/upImages.vue';
  import typeChange from './components/typeChange.vue';
  import taxChange from './components/taxChange.vue';

  // 获取路由实例
  const router = useRouter();
  const route = useRoute();

  // 响应式数据
  const loading = ref(false);
  const openProductSelector = ref(false);
  const model = ref({});
  const form = reactive({
    product_ids: [],
    category_id: '',
    productTaxes: [
      {
        product_tax_type: '1',
        tax_category_id: '',
      },
      {
        product_tax_type: '2',
        tax_category_id: '',
      },
    ],
  });
  const product_list = ref([]);
  const type = ref(1);
  const title = ref('');

  // 模板引用
  const formRef = ref(null);
  const upImagesRef = ref(null);
  const typeChangeRef = ref(null);
  const taxChangeRef = ref(null);

  // 提供form给子组件
  provide('form', form);

  // 监听product_list变化
  watch(
    product_list,
    (val) => {
      form.product_ids = val.map((item) => {
        return item.product_id;
      });

      if (formRef.value) {
        formRef.value.validateField('product_ids');
      }
    },
    { deep: true, immediate: true }
  );

  // 组件挂载时初始化
  onMounted(() => {
    type.value = route.query.type;
    switch (type.value) {
      case '1':
        title.value = $t('修改图片');
        break;
      case '2':
        title.value = $t('修改分类');
        break;
      case '3':
        title.value = $t('修改税类');
        break;
      case '5':
        title.value = $t('商品批量导入');
        break;
    }
  });

  // 方法定义
  const dialogFormVisible = () => {
    router.go(-1);
  };

  const handleClick = () => {
    formRef.value.validate((valid) => {
      if (valid) {
        if (type.value == 1) {
          upImagesRef.value.repeatList();
        }
        if (type.value == 2) {
          typeChangeRef.value.submit();
        }
        if (type.value == 3) {
          taxChangeRef.value.submit();
        }
      }
    });
  };

  const addProduct = () => {
    openProductSelector.value = true;
    model.value.product_ids = [];
    if (product_list.value.length > 0) {
      product_list.value.map((item) => {
        model.value.product_ids.push({
          product_id: item.product_id,
        });
      });
    }
  };

  const handleProductSelectorClose = (list, categories) => {
    if (Array.isArray(list)) {
      product_list.value = list;
    }
    if (Array.isArray(categories)) {
      product_list.value.map((item) => {
        item.path_name_text = '';
        categories.map((item2) => {
          if (item.category_id == item2.category_id) {
            item.path_name_text = item2.path_name_text;
          }
          if (item2.child.length > 0) {
            item2.child.map((item3) => {
              if (item.category_id == item3.category_id) {
                item.path_name_text = item3.path_name_text;
              }
            });
          }
        });
      });
    }

    openProductSelector.value = false;
  };

  const returnBack = () => {
    router.go(-1);
  };

  const removeProduct = (index) => {
    product_list.value.splice(index, 1);
  };
</script>

<style lang="scss" scoped>
  .product-body {
    display: flex;
    flex-direction: column;
    width: 100%;
    overflow: auto;
    .product-list {
      display: flex;
      box-shadow: 0 0 0 1px var(--el-input-border-color, var(--el-border-color)) inset;
      border-radius: 4px;
      padding: 6px 11px;
      margin-top: 16px;
      gap: 8px;
      flex-wrap: wrap;
      .product-item {
        color: var(--el-tag-text-color);
        display: inline-flex;
        justify-content: center;
        align-items: center;
        vertical-align: middle;
        font-size: var(--el-tag-font-size);
        line-height: 1;
        border-width: 1px;
        border-style: solid;
        box-sizing: border-box;
        white-space: nowrap;
        padding: 0 7px;
        height: 20px;
        font-size: 12px;
        border-radius: 4px;
        background-color: var(--el-color-info-light-9);
        border: solid 1px var(--el-color-info-light-8);
      }
      .el-tag__close {
        font-size: 12px;
        margin-left: 4px;
        cursor: pointer;
      }
    }
  }
  .product-batch {
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

  .product-body {
    flex: 1 1 auto;
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
  }
</style>

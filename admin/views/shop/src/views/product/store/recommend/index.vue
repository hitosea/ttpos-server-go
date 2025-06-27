<template>
  <div class="product-list" v-loading="loading">
    <div data-v-164f2ab7="" class="common-form">{{ $t('推荐商品管理') }}</div>
    <p class="gray9">{{ $t('设置店铺推荐商品，提升商品曝光度和销量') }}</p>
    <el-card class="box-card" shadow="none">
      <el-form label-position="top" ref="formRef" :model="form" size="small">
        <el-form-item :label="$t('是否开启推荐')" prop="status" :rules="[{ required: true, message: $t('请选择是否开启推荐') }]">
          <el-radio-group v-model="form.status">
            <el-radio :label="1">{{ $t('开启') }}</el-radio>
            <el-radio :label="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.status === 1">
          <el-form-item :label="$t('推荐标题')" prop="title" :rules="[{ required: true, message: $t('请输入推荐标题') }]">
            <el-input v-model="form.title" :maxlength="30" :placeholder="$t('请输入推荐标题')" type="textarea" :rows="2" show-word-limit></el-input>
          </el-form-item>
          <el-form-item :label="$t('选择商品') + `(${form.product_packages.length}/15)`" prop="product_packages" :rules="[{ required: true, validator: validateProductPackages }]">
            <el-button type="primary" @click="openProductSelectorDialog">{{ $t('选择商品') }}</el-button>
          </el-form-item>
          <el-table :data="form.product_packages" size="small">
            <el-table-column prop="name" :label="$t('商品名称')"></el-table-column>
            <el-table-column prop="sort" :label="$t('排序')">
              <template #default="scope">
                <el-form-item
                  :prop="`product_packages.${scope.$index}.sort`"
                  :rules="[{ required: true, validator: (rule, value, callback) => validateSort(rule, value, callback, scope.$index) }]"
                >
                  <el-input-number class="mt16" v-model="scope.row.sort" :controls="false" :min="0" :precision="0" :placeholder="$t('请输入排序')"></el-input-number>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="action" :label="$t('操作')">
              <template #default="scope">
                <el-button type="danger" link @click="removeProduct(scope.$index)">{{ $t('移除') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </el-form>
    </el-card>
    <!--提交-->
    <div class="common-button-wrapper">
      <el-button @click="getRecommend" :loading="loading">{{ $t('重置') }}</el-button>
      <el-button type="primary" @click="submitForm" :loading="loading">{{ $t('保存') }}</el-button>
    </div>
    <!-- 商品选择器 -->
    <ProductSelector
      :maxCount="15"
      v-if="openProductSelector"
      :open="openProductSelector"
      @close="handleProductSelectorClose"
      selectorType="all"
      :selectedProductIds="selectedProductIds"
    >
    </ProductSelector>
  </div>
</template>

<script setup>
  import { ref, onMounted, getCurrentInstance } from 'vue';
  import ProductApi from '@/api/product';
  import ProductSelector from '@/components/product/Selector.vue';
  const { proxy } = getCurrentInstance();
  const openProductSelector = ref(false);

  const formRef = ref(null);
  const form = ref({
    status: 1,
    title: '',
    product_packages: [],
  });

  const selectedProductIds = ref([]);

  const loading = ref(false);

  const validateProductPackages = (_rule, value, callback) => {
    if (value.length < 3) {
      callback(new Error(proxy.$t('至少选择3个商品')));
    } else {
      callback();
    }
  };

  const validateSort = (_rule, value, callback, index) => {
    if (value === null) {
      callback(new Error(proxy.$t('请输入排序')));
    }
    // 排序不能重复
    else if (form.value.product_packages.some((item) => item.sort === value && item.uuid !== form.value.product_packages[index].uuid)) {
      callback(new Error(proxy.$t('排序不能重复')));
    } else {
      callback();
    }
  };

  const submitForm = () => {
    formRef.value.validate((valid) => {
      if (valid) {
        loading.value = true;
        ProductApi.setRecommend(form.value)
          .then((res) => {
            console.log(res);
          })
          .catch((error) => {
            console.log(error);
          })
          .finally(() => {
            loading.value = false;
          });
      }
    });
  };

  const removeProduct = (index) => {
    form.value.product_packages.splice(index, 1);
    formRef.value.validateField('product_packages');
  };

  const getRecommend = () => {
    loading.value = true;
    ProductApi.getRecommend()
      .then((res) => {
        form.value = res.data;
      })
      .catch((error) => {
        console.log(error);
      })
      .finally(() => {
        loading.value = false;
      });
  };

  const openProductSelectorDialog = () => {
    selectedProductIds.value = form.value.product_packages.map((item) => Number(item.uuid));
    openProductSelector.value = true;
  };

  const handleProductSelectorClose = (list) => {
    openProductSelector.value = false;
    if (!list) {
      return;
    }
    // 以 form.value.product_packages 为基准，如果 list 中有 form.value.product_packages 中没有的，就添加到 form.value.product_packages 中
    list.forEach((item) => {
      if (!form.value.product_packages.some((item2) => Number(item2.uuid) === Number(item.uuid))) {
        form.value.product_packages.push({
          uuid: Number(item.uuid),
          name: item.product_name_text,
          sort: null,
        });
      }
    });

    // 以 form.value.product_packages 为基准，如果 form.value.product_packages 中有 list 中没有的，就从 form.value.product_packages 中删除
    form.value.product_packages.forEach((item) => {
      if (!list.some((item2) => Number(item2.uuid) === Number(item.uuid))) {
        form.value.product_packages.splice(form.value.product_packages.indexOf(item), 1);
      }
    });
    formRef.value.validateField('product_packages');
  };

  onMounted(() => {
    getRecommend();
  });
</script>

<style scoped lang="scss">
  .box-card {
    margin-top: 24px;
  }
  .common-button-wrapper {
    border-top: none;
  }
</style>

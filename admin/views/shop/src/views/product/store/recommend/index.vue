<template>
  <div class="product-list">
    <div data-v-164f2ab7="" class="common-form">{{ $t('推荐商品管理') }}</div>
    <p class="gray9">{{ $t('设置店铺推荐商品，提升商品曝光度和销量') }}</p>
    <el-card class="box-card" shadow="none">
      <el-form label-position="top" ref="formRef" :model="form" size="small">
        <el-form-item :label="$t('是否开启推荐')" prop="radio" :rules="[{ required: true, message: $t('请选择是否开启推荐') }]">
          <el-radio-group v-model="form.status">
            <el-radio :label="1">{{ $t('开启') }}</el-radio>
            <el-radio :label="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('推荐标题')" prop="title" :rules="[{ required: true, message: $t('请输入推荐标题') }]">
          <el-input v-model="form.title" :maxlength="30" :placeholder="$t('请输入推荐标题')" type="textarea" :rows="2" show-word-limit></el-input>
        </el-form-item>
        <el-form-item :label="$t('选择商品') + `(0/15)`" prop="product_ids" :rules="[{ required: true, message: $t('请选择商品') }]">
          <el-button type="primary" @click="openProductSelectorDialog">选择商品</el-button>
        </el-form-item>
        <el-table :data="form.product_packages" size="small">
          <el-table-column prop="name" label="商品名称"></el-table-column>
          <el-table-column prop="sort" label="排序">
            <template #default="scope">
              <el-form-item :prop="`product_packages.${scope.$index}.sort`" :rules="[{ required: true, message: $t('请输入排序') }]">
                <el-input-number class="mt16" v-model="scope.row.sort" :controls="false" :min="0" :precision="0" :placeholder="$t('请输入排序')"></el-input-number>
              </el-form-item>
            </template>
          </el-table-column>
          <el-table-column prop="action" label="操作">
            <template #default="scope">
              <el-button type="danger" link @click="removeProduct(scope.$index)">{{ $t('移除') }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-form>
    </el-card>

    <!-- 商品选择器 -->
    <ProductSelector
      :maxCount="5"
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
  import { ref, onMounted } from 'vue';
  import ProductApi from '@/api/product';
  import ProductSelector from '@/components/product/Selector.vue';

  const openProductSelector = ref(false);

  const formRef = ref(null);
  const form = ref({
    status: 1,
    title: '',
    product_packages: [],
  });

  const selectedProductIds = ref([]);

  const submitForm = (formName) => {
    formRef.value.validate((valid) => {
      if (valid) {
        alert('submit!');
      } else {
        console.log('error submit!!');
        return false;
      }
    });
  };

  const removeProduct = (index) => {
    form.value.product_packages.splice(index, 1);
  };

  const getRecommend = () => {
    ProductApi.getRecommend()
      .then((res) => {
        form.value = res.data;
      })
      .catch((error) => {
        console.log(error);
      });
  };

  const openProductSelectorDialog = () => {
    selectedProductIds.value = form.value.product_packages.map((item) => item.uuid);
    openProductSelector.value = true;
  };

  const handleProductSelectorClose = (list) => {
    if (Array.isArray(list)) {
      form.value.product_packages = list.map((item) => {
        return {
          uuid: item.product_id,
          product_name: item.product_name,
          sort: null,
        };
      });
    }
    openProductSelector.value = false;
  };

  onMounted(() => {
    getRecommend();
  });
</script>

<style scoped lang="scss">
  .box-card {
    margin-top: 24px;
  }
</style>

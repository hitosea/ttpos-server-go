<template>
  <div class="product-add">
    <!--form表单-->
    <el-form size="small" ref="formRef" class="product-form" :model="form" label-position="top" label-width="180px">
      <div class="product-form-wrapper">
        <div class="product-form-left">
          <ImagePreview :image="form.image_base64"></ImagePreview>
        </div>
        <!--分割线-->
        <div class="product-form-line"></div>
        <div class="product-form-flex" ref="formContainer">
          <!--基础信息-->
          <Basic></Basic>
          <!--高级设置-->
          <Set></Set>
          <!--提交-->
        </div>
      </div>

      <div class="common-button-wrapper">
        <el-button size="small" @click="cancelFunc">{{ $t('取消') }}</el-button>
        <el-button size="small" type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup>
  import { ref, reactive, provide, onMounted, getCurrentInstance } from 'vue';
  import { useRouter } from 'vue-router';
  import { ElMessage } from 'element-plus';
  import MarketingApi from '@/api/marketing.js';
  import Basic from './part/Basic.vue';
  import Set from './part/set.vue';
  import { useUserStore } from '@/store/index';
  import ImagePreview from './part/imagePreview.vue';
  // 获取全局属性
  const { proxy } = getCurrentInstance();
  const { $t } = proxy;
  const router = useRouter();

  // 获取store
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;

  // 响应式数据
  const formRef = ref();
  const formContainer = ref();
  const loading = ref(false);

  // 表单数据
  const form = reactive({
    name: '',
    description: '',
    start_time: '',
    end_time: '',
    reward_condition_amount: 0,
    reward_condition_num: 0,
    reward_type: 0,
    prize_list: [],
    image_base64: '',
  });

  const qrcode = ref('');

  // 向子组件提供数据
  provide('form', form);
  provide('image', form.image);

  // 方法
  const getBaseData = () => {
    MarketingApi.activityAddGet({}, true)
      .then((res) => {
        loading.value = false;
        qrcode.value = res.data.qrcode;
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  const onSubmit = () => {
    const params = form.model;
    if (params.is_discount == 0) params.discount = 0;

    formRef.value.validate((valid) => {
      if (valid) {
        loading.value = true;
        CardApi.addcard(
          {
            params: JSON.stringify(params),
          },
          true
        )
          .then((data) => {
            loading.value = false;
            ElMessage({
              message: $t('添加成功'),
              type: 'success',
            });
            router.push('/' + app_id + '/card/card/index');
          })
          .catch((error) => {
            loading.value = false;
          });
      }
    });
  };

  const cancelFunc = () => {
    router.back(-1);
  };

  // 生命周期
  onMounted(() => {
    /*获取基础数据*/
    getBaseData();
  });
</script>

<style lang="scss" scoped>
  .basic-setting-content {
  }

  .product-add {
    height: calc(100% - 14px);
    overflow: hidden;
  }

  .product-form {
    height: 100%;
    overflow: hidden;
    display: flex;
    flex-direction: column;

    .product-form-wrapper {
      flex: 1 1 auto;
      display: flex;
      overflow: hidden;
    }

    .product-form-left {
      flex-shrink: 0;
      width: 312px;
      height: 100%;
      overflow-y: auto;
    }

    .product-form-line {
      width: 1px;
      height: 100%;
      background-color: #ebeef5;
      margin: 0 20px 0 36px;
    }

    .product-form-flex {
      flex: 1 1 auto;
      overflow-y: auto;
    }

    .common-button-wrapper {
      flex: 0 0 auto;
      flex-shrink: 0;
    }
  }
</style>

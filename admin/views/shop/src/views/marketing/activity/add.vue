<template>
  <div class="product-add">
    <!--form表单-->
    <el-form size="small" ref="formRef" class="product-form" :model="form" label-position="top" label-width="180px">
      <div class="product-form-wrapper">
        <div class="product-form-left">
          <ImagePreview ref="ImagePreviewRef" :qrcode="qrcode" :imgName="imgName" :imgDescription="imgDescription" @checkForm="checkFormAll"></ImagePreview>
        </div>
        <!--分割线-->
        <div class="product-form-line"></div>
        <div class="product-form-flex" ref="formContainer">
          <!--基础信息-->
          <Basic ref="BasicRef" @imgName="imgNameChange" @imgDescription="imgDescriptionChange" @checkForm="checkForm"></Basic>
          <!--高级设置-->
          <Set ref="SetRef"></Set>
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
  const BasicRef = ref();
  const SetRef = ref();
  const formContainer = ref();
  const loading = ref(false);
  const ImagePreviewRef = ref();
  // 表单数据
  const form = reactive({
    name: '',
    description: '',
    start_time: '',
    end_time: '',
    reward_condition_amount: null,
    is_open_reward_limit: [],
    reward_limit: null,
    reward_condition_num: 0,
    reward_type: 0,
    prize_list: [],
    image_base64: '',
  });

  const qrcode = ref('');
  const imgName = ref('');
  const imgDescription = ref('');

  // 向子组件提供数据
  provide('form', form);
  provide('image', form.image);

  // 方法
  const getBaseData = () => {
    MarketingApi.activityAddGet({}, true)
      .then((res) => {
        loading.value = false;
        qrcode.value = res.data.qr_code;
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  const imgNameChange = (data) => {
    imgName.value = data;
  };

  const imgDescriptionChange = (data) => {
    imgDescription.value = data;
  };

  const checkForm = (e) => {
    formRef.value?.validateField(e);
  };

  const checkFormAll = async () => {
    // 验证表单
    const validUniqueName = await BasicRef.value.$refs.activityNameFormRef.validate();
    const validUniqueDescription = await BasicRef.value.$refs.activityDescriptionFormRef.validate();
    formRef.value.validate((valid) => {
      if (valid && validUniqueName && validUniqueDescription) {
        ImagePreviewRef.value.downloadImage();
      }
    });
  };

  const onSubmit = async () => {
    const params = JSON.parse(JSON.stringify(form));
    const _name = BasicRef.value.$refs.activityNameFormRef.data;
    params.name = JSON.stringify(_name);
    const _description = BasicRef.value.$refs.activityDescriptionFormRef.data;
    params.description = JSON.stringify(_description);

    params.is_open_reward_limit = params.is_open_reward_limit.length > 0 ? 1 : 0;

    // 验证表单
    const validUniqueName = await BasicRef.value.$refs.activityNameFormRef.validate();
    const validUniqueDescription = await BasicRef.value.$refs.activityDescriptionFormRef.validate();

    formRef.value.validate((valid) => {
      if (valid && validUniqueName && validUniqueDescription) {
        loading.value = true;
        MarketingApi.activityAdd(params, true)
          .then((data) => {
            loading.value = false;
            ElMessage({
              message: $t('添加成功'),
              type: 'success',
            });
            router.push('/' + app_id + '/marketing/activity/index');
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

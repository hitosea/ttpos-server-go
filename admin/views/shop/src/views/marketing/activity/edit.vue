<template>
  <div class="product-add" v-loading="loading">
    <!--form表单-->
    <el-form size="small" ref="formRef" class="product-form" :model="form" label-position="top" label-width="180px" v-if="!loading">
      <div class="product-form-wrapper">
        <div class="product-form-left">
          <ImagePreview ref="ImagePreviewRef" :qrcode="qrcode" :imgName="imgName" :imgDescription="imgDescription" @checkForm="checkFormAll"></ImagePreview>
        </div>
        <!--分割线-->
        <div class="product-form-line"></div>
        <div class="product-form-flex" ref="formContainer">
          <!--基础信息-->
          <Basic
            ref="BasicRef"
            @imgName="imgNameChange"
            @imgDescription="imgDescriptionChange"
            @checkForm="checkForm"
            :dateTime="dateTime"
            :couponList="couponList"
            :status="status"
          ></Basic>
          <!--高级设置-->
          <Set ref="SetRef" :status="status"></Set>
          <!--提交-->
        </div>
      </div>

      <!--提交-->
      <div class="common-button-wrapper">
        <el-button size="small" @click="cancelFunc">{{ $t('取消') }}</el-button>
        <el-button size="small" type="primary" @click="onSubmit" :disabled="save_loading">{{ $t('确定') }}</el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup>
  import { ref, reactive, provide, onMounted, nextTick } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { ElMessage } from 'element-plus';
  import MarketingApi from '@/api/marketing.js';
  import Basic from './part/Basic.vue';
  import Set from './part/set.vue';
  import ImagePreview from './part/imagePreview.vue';

  // 获取路由实例
  const route = useRoute();
  const router = useRouter();

  // 响应式数据

  const uuid = ref(0);
  const loading = ref(true);
  const save_loading = ref(false);
  const dateTime = ref([]);
  const couponList = ref([]);
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

  // 引用
  const formRef = ref(null);
  const BasicRef = ref(null);
  const SetRef = ref(null);
  const formContainer = ref(null);
  const qrcode = ref('');
  const status = ref(0);
  const ImagePreviewRef = ref(null);
  const imgName = ref('');
  const imgDescription = ref('');
  // 提供form数据给子组件
  provide('form', form);

  // 获取基础数据
  const getData = () => {
    loading.value = true;
    MarketingApi.activityEditGet(
      {
        uuid: uuid.value,
      },
      true
    )
      .then((res) => {
        loading.value = false;
        form.name = JSON.parse(res.data.detail.name || '{}');
        form.description = JSON.parse(res.data.detail.description || '{}');
        form.start_time = res.data.detail.start_time;
        form.end_time = res.data.detail.end_time;
        dateTime.value = [res.data.detail.start_time, res.data.detail.end_time];
        form.reward_condition_amount = res.data.detail.reward_condition_amount;
        form.is_open_reward_limit = res.data.detail.is_open_reward_limit == 1 ? [1] : [];
        form.reward_limit = res.data.detail.reward_limit;
        form.reward_condition_num = res.data.detail.reward_condition_num;
        form.reward_type = res.data.detail.reward_type;
        form.image_base64 = res.data.detail.image_base64;
        qrcode.value = res.data.qr_code;
        if (res.data.detail.prizes.length > 0) {
          form.prize_list = [];
          res.data.detail.prizes.forEach((item) => {
            form.prize_list.push({
              prize_type: 1,
              prize_uuid: item.prize_uuid,
            });
          });
          couponList.value = res.data.detail.prizes;
        }
        status.value = res.data.detail.status;
      })
      .catch((error) => {
        loading.value = false;
        console.error('获取数据失败:', error);
      });
  };

  // 提交表单
  const onSubmit = async () => {
    const params = JSON.parse(JSON.stringify(form));
    params.uuid = uuid.value;
    const _name = BasicRef.value.$refs.activityNameFormRef.data;
    params.name = JSON.stringify(_name);
    const _description = BasicRef.value.$refs.activityDescriptionFormRef.data;
    params.description = JSON.stringify(_description);
    params.is_open_reward_limit = params.is_open_reward_limit.length > 0 ? 1 : 0;
    params.prize_list = [];
    (form.prize_list || []).map((e) => {
      params.prize_list.push({
        prize_type: 1,
        prize_uuid: e.prize_uuid,
      });
    });
    // 验证表单
    const validUniqueName = await BasicRef.value.$refs.activityNameFormRef.validate();
    const validUniqueDescription = await BasicRef.value.$refs.activityDescriptionFormRef.validate();
    formRef.value.validate((valid) => {
      if (valid && validUniqueName && validUniqueDescription) {
        save_loading.value = true;
        MarketingApi.activityEdit(params, true)
          .then((data) => {
            save_loading.value = false;
            ElMessage({
              message: window.$t('保存成功'),
              type: 'success',
            });
            cancelFunc();
          })
          .catch((error) => {
            save_loading.value = false;
            console.error('保存失败:', error);
          });
      }
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

  const checkFormAll = async (type, callback) => {
    // 验证表单
    const validUniqueName = await BasicRef.value.$refs.activityNameFormRef.validate();
    const validUniqueDescription = await BasicRef.value.$refs.activityDescriptionFormRef.validate();
    formRef.value.validate((valid) => {
      if (valid && validUniqueName && validUniqueDescription) {
        if (type == 'download') {
          ImagePreviewRef.value.downloadImage();
        } else if (type == 'preview') {
          callback(true);
        }
      } else {
        ElMessage({
          message: $t('请完善输入信息'),
          type: 'error',
        });
      }
    });
  };

  // 取消操作
  const cancelFunc = () => {
    router.back(-1);
  };

  // 组件挂载时执行
  onMounted(() => {
    // 获取路由参数
    uuid.value = route.query.uuid || 0;
    getData();
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

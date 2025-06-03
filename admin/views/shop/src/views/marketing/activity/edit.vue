<template>
  <div class="product-add" v-loading="loading">
    <!--form表单-->
    <el-form size="small" ref="formRef" class="product-form" :model="form" label-position="top" label-width="180px" v-if="!loading">
      <div class="product-form-wrapper">
        <div class="product-form-left">
          <ImagePreview :qrcode="qrcode" :imgName="imgName" :imgDescription="imgDescription"></ImagePreview>
        </div>
        <!--分割线-->
        <div class="product-form-line"></div>
        <div class="product-form-flex" ref="formContainer">
          <!--基础信息-->
          <Basic ref="BasicRef" @imgName="imgNameChange" @imgDescription="imgDescriptionChange" :dateTime="dateTime"></Basic>
          <!--高级设置-->
          <Set ref="SetRef"></Set>
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
  import { ref, reactive, provide, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { ElMessage } from 'element-plus';
  import MarketingApi from '@/api/marketing.js';
  import Basic from './part/Basic.vue';
  import Set from './part/set.vue';
  import ImagePreview from './part/imagePreview.vue';
  import { nextTick } from 'vue';

  // 获取路由实例
  const route = useRoute();
  const router = useRouter();

  // 响应式数据

  const uuid = ref(0);
  const loading = ref(true);
  const save_loading = ref(false);
  const dateTime = ref([]);
  // 表单数据
  const form = reactive({
    name: '',
    description: '',
    start_time: '',
    end_time: '',
    reward_condition_amount: 0,
    is_open_reward_limit: [],
    reward_limit: 0,
    reward_condition_num: 0,
    reward_type: 0,
    prize_list: [
      {
        prize_type: 1,
        prize_uuid: 123123123,
      },
    ],
    image_base64: '',
  });

  // 引用
  const formRef = ref(null);
  const BasicRef = ref(null);
  const SetRef = ref(null);
  const formContainer = ref(null);
  const qrcode = ref('');

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
        form.prize_list = res.data.detail.prize_list;
        form.image_base64 = res.data.detail.image_base64;
        qrcode.value = res.data.qr_code;
      })
      .catch((error) => {
        loading.value = false;
        console.error('获取数据失败:', error);
      });
  };

  // 提交表单
  const onSubmit = () => {
    formRef.value.validate((valid) => {
      if (valid) {
        let params = { ...form.model };
        params.card_id = card_id.value;
        params.start_time = form.start_time;
        params.end_time = form.end_time;
        params.reward_type = form.reward_type;

        if (params.is_discount == 0) {
          params.discount = 0;
        }

        save_loading.value = true;
        MarketingApi.editActivity(
          {
            card_id: card_id.value,
            params: JSON.stringify(params),
          },
          true
        )
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

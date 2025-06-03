<template>
  <div class="basic-setting-content pl16 pr16">
    <!--基本信息-->
    <div class="common-form">{{ $t('基本信息') }}</div>
    <el-form-item for="no_click" :rules="[{ required: true, message: $t('请输入活动名称') }]" prop="model.card_name">
      <UniqueNameForm ref="activityNameFormRef" :labelPrefix="$t('活动名称')" width="460px" :maxlength="50" />
    </el-form-item>
    <el-form-item for="no_click" :rules="[{ required: true, message: $t('请输入活动文案') }]" prop="model.card_name">
      <UniqueNameForm ref="activityDescriptionFormRef" :labelPrefix="$t('活动文案')" width="460px" :maxlength="100" />
    </el-form-item>
    <el-form-item for="no_click" :label="$t('活动时间')" :rules="[{ required: true, message: $t('请选择活动时间') }]" prop="model.card_name">
      <el-date-picker class="max-w460" v-model="activityTime" type="daterange" range-separator="~" start-placeholder="开始日期" end-placeholder="结束日期" />
    </el-form-item>
    <el-form-item for="no_click" :label="$t('活动奖品')" :rules="[{ required: true, message: $t('请选择活动奖品') }]" prop="model.card_name">
      <el-radio-group v-model="rewardType">
        <el-radio :label="0">{{ $t('优惠券（当前仅支持选择优惠券）') }}</el-radio>
      </el-radio-group>
    </el-form-item>
  </div>
</template>

<script setup>
  import { ref, inject, watch } from 'vue';
  import { ElMessage } from 'element-plus';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';

  // 注入form数据
  const form = inject('form');

  // 响应式数据
  const isupload = ref(false);
  const open_add = ref(false);
  const type = ref('');
  const activityTime = ref([]);
  const rewardType = ref(0);

  // 引用
  const activityNameFormRef = ref(null);
  const activityDescriptionFormRef = ref(null);

  // 方法定义
  const chooseCardType = (e) => {
    form.model.card_style = e;
  };

  // 添加优惠券
  const addCoupon = () => {
    if (form.model.open_coupons.length >= 15) {
      ElMessage.error('您已经选择了十五张优惠券，若要更换请删除其他优惠券！');
      return;
    }
    open_add.value = true;
  };

  // 关闭优惠券
  const closeProductDialogFunc = (e) => {
    open_add.value = e.openDialog;
    if (e.type == 'success') {
      let params = {
        coupon_id: e.params.coupon_id,
        name: e.params.name,
        number: 1,
        color: e.params.color,
        discount: e.params.discount,
        reduce_price: e.params.reduce_price,
        coupon_type: e.params.coupon_type,
        min_price: e.params.min_price,
      };
      form.model.open_coupons.push(params);
    }
  };

  // 删除优惠券
  const delcoupon = (item) => {
    let n = form.model.open_coupons.indexOf(item);
    form.model.open_coupons.splice(n, 1);
  };

  // 上传
  const openUpload = (e) => {
    type.value = e;
    isupload.value = true;
  };

  // 获取图片
  const returnImgsFunc = (e) => {
    if (e != null && e.length > 0) {
      form.model.default_style = e[0].file_path;
    }
    isupload.value = false;
  };

  watch(activityTime, (newVal) => {
    form.start_time = newVal[0];
    form.end_time = newVal[1];
  });
</script>

<style lang="scss">
  .edit_container {
    font-family: 'Avenir', Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    text-align: center;
    line-height: 20px;
    color: #2c3e50;
  }

  .ql-editor {
    height: 400px;
  }

  .draggable-list {
    display: flex;
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .draggable-list .wrapper > span {
    display: flex;
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .draggable-list .item {
    position: relative;
    width: 110px;
    height: 110px;
    margin-top: 10px;
    margin-right: 10px;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid #dddddd;
  }

  .draggable-list .delete-btn {
    position: absolute;
    top: 0;
    right: 0;
    width: 16px;
    height: 16px;
    background: red;
    line-height: 16px;
    font-size: 16px;
    color: #ffffff;
    display: none;
  }

  .draggable-list .item:hover .delete-btn {
    display: block;
  }

  .draggable-list .item img {
    position: absolute;
    top: 50%;
    left: 50%;
    -webkit-transform: translate(-50%, -50%);
    transform: translate(-50%, -50%);
    max-height: 100%;
    max-width: 100%;
  }

  .draggable-list .img-select {
    display: flex;
    justify-content: center;
    align-items: center;
    border: 1px dashed #dddddd;
    font-size: 30px;
  }

  .draggable-list .img-select i {
    color: #409eff;
  }

  .card-el-row {
    margin-bottom: 20px;
    margin-right: 20px;
  }

  .maxwidth-530 {
    max-width: 530px;
  }

  .card {
    border-radius: 4px;
  }

  .active.card {
    border: 2px solid #4aa3f7;
  }
</style>

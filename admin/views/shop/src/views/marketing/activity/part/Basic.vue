<template>
  <div class="basic-setting-content pl16 pr16">
    <!--基本信息-->
    <div class="common-form">{{ $t('基本信息') }}</div>
    <el-form-item for="no_click" :rules="[{ required: true, message: $t('请输入活动名称') }]">
      <UniqueNameForm
        ref="activityNameFormRef"
        :labelPrefix="$t('活动名称')"
        width="460px"
        :maxlength="50"
        :overrideLanguages="form.name"
        :isUnique="false"
        @nowLangeData="imgName"
      />
    </el-form-item>
    <el-divider border-style="dashed" />
    <el-form-item for="no_click" :rules="[{ required: true, message: $t('请输入活动文案') }]">
      <UniqueNameForm
        ref="activityDescriptionFormRef"
        :labelPrefix="$t('活动文案')"
        width="460px"
        :maxlength="100"
        :overrideLanguages="form.description"
        :isUnique="false"
        @nowLangeData="imgDescription"
      />
    </el-form-item>
    <el-form-item
      for="no_click"
      :label="$t('活动时间')"
      :rules="[
        {
          required: true,
          validator: () => {
            return form.start_time && form.end_time ? true : false;
          },
          message: $t('请选择活动时间'),
        },
      ]"
      prop="start_time"
    >
      <el-date-picker
        class="max-w460"
        v-model="activityTime"
        type="datetimerange"
        value-format="YYYY-MM-DD HH:mm:ss"
        range-separator="~"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
      />
    </el-form-item>
    <el-form-item class="flex-box" for="no_click" :label="$t('活动奖品')" :rules="[{ required: true, message: $t('请选择活动奖品') }]">
      <el-radio-group v-model="rewardType">
        <el-radio :label="0">{{ $t('优惠券（当前仅支持选择优惠券）') }}</el-radio>
      </el-radio-group>
      <div>
        <span class="select-coupon-btn" @click="selectCoupon">{{ $t('选择优惠券') }}</span>
      </div>
    </el-form-item>
  </div>
</template>

<script setup>
  import { ref, inject, watch } from 'vue';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';

  // 注入form数据
  const form = inject('form');

  // 响应式数据
  const activityTime = ref([]);
  const rewardType = ref(0);

  const props = defineProps({
    dateTime: {
      type: Array,
      default: () => [],
    },
  });
  // 引用
  const activityNameFormRef = ref(null);
  const activityDescriptionFormRef = ref(null);

  const emit = defineEmits(['imgName', 'imgDescription']);

  watch(
    props.dateTime,
    (newVal) => {
      activityTime.value = [newVal[0], newVal[1]];
    },
    {
      immediate: true,
      deep: true,
    }
  );

  watch(activityTime, (newVal) => {
    form.start_time = newVal[0];
    form.end_time = newVal[1];
  });

  const imgName = (data) => {
    emit('imgName', data);
  };

  const imgDescription = (data) => {
    emit('imgDescription', data);
  };

  const selectCoupon = () => {
    console.log('selectCoupon');
  };
</script>

<style lang="scss" scoped>
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
  :deep(.el-radio.el-radio--small) {
    height: 30px;
    width: 100%;
    .el-radio__input.is-checked + .el-radio__label {
      color: #100a05;
    }
  }

  .flex-box {
    :deep(.el-form-item__content) {
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      justify-content: flex-start;
    }
  }

  .select-coupon-btn {
    color: #ffbe00;
    font-size: 14px;
    font-style: normal;
    font-weight: 400;
    cursor: pointer;
  }
</style>

<template>
  <div class="basic-setting-content pl16 pr16">
    <!--基本信息-->
    <div class="common-form">{{ $t('基本信息') }}</div>
    <el-form-item for="no_click" :label="$t('活动类型')" prop="type" :rules="[{ required: true, message: $t('请选择活动类型') }]">
      <el-select class="max-w460" v-model="form.type" :placeholder="$t('请选择活动类型')">
        <el-option :key="0" :label="$t('邀请消费有礼')" :value="0"></el-option>
      </el-select>
    </el-form-item>

    <el-form-item for="no_click" :rules="[{ required: true, message: $t('请输入活动名称') }]">
      <UniqueNameForm
        ref="activityNameFormRef"
        :labelPrefix="$t('活动名称')"
        :placeholder="$t('请输入活动名称')"
        width="460px"
        :maxlength="50"
        :overrideLanguages="form.name"
        :isUnique="false"
        :disabled="status == 1"
        @nowLangeData="imgName"
      />
    </el-form-item>
    <el-divider border-style="dashed" />
    <el-form-item for="no_click" :rules="[{ required: true, message: $t('请输入活动文案') }]">
      <UniqueNameForm
        ref="activityDescriptionFormRef"
        :labelPrefix="$t('活动文案')"
        :disabled="status == 1"
        :placeholder="$t('请输入活动文案')"
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
        v-if="status == 0 || status == null"
        class="max-w460"
        v-model="activityTime"
        type="datetimerange"
        value-format="YYYY-MM-DD HH:mm:ss"
        format="YYYY-MM-DD HH:mm:ss"
        range-separator="~"
        :start-placeholder="$t('开始日期')"
        :end-placeholder="$t('结束日期')"
        :disabledDate="disabledDate"
        :disabled-hours="(role, date) => (role === 'start' && isTodayTemp ? dHours() : [])"
        :disabled-minutes="(role, date) => (role === 'start' && isTodayTemp ? dMinutes() : [])"
        :disabled-seconds="(role, date) => (role === 'start' && isTodayTemp ? dSeconds() : [])"
        @change="handleChange"
        @calendar-change="
          (value) => {
            if (value && value.length === 2) {
              isTodayTemp = isToday(new Date(value[0]));
            }
          }
        "
      />
      <el-date-picker
        v-if="status == 1"
        v-model="form.end_time"
        type="datetime"
        value-format="YYYY-MM-DD HH:mm:ss"
        format="YYYY-MM-DD HH:mm:ss"
        :placeholder="$t('结束日期')"
        class="max-w460 w100"
        :disabledDate="disabledDate"
        :disabled-hours="disabledHours"
        :disabled-minutes="disabledMinutes"
        :disabled-seconds="disabledSeconds"
        @change="handleChangeEnd"
      />
    </el-form-item>
    <el-form-item for="no_click" :label="$t('活动奖品')" :rules="[{ required: true, message: $t('请选择活动奖品') }]">
      <el-radio-group v-model="form.reward_type" :disabled="status == 1">
        <el-radio :value="0">{{ $t('优惠券') }}</el-radio>
        <el-radio :value="1">{{ $t('积分') }}</el-radio>
      </el-radio-group>
    </el-form-item>
    <el-form-item
      v-if="form.reward_type == 0"
      for="no_click"
      prop="prize_list"
      :rules="[
        {
          required: true,
          validator: () => {
            return form.prize_list.length > 0 ? true : false;
          },
          message: $t('请选择活动奖品'),
        },
      ]"
    >
      <div>
        <el-button type="primary" size="small" @click="selectCoupon" :disabled="status == 1">{{ $t('选择优惠券') }}</el-button>
        <template v-if="couponList.length > 0">
          <el-tag v-for="tag in couponList" disable-transitions size="large" :key="tag.uuid" :closable="status != 1" @close="handleClose(tag)">
            {{ `${tag.name}` }}
          </el-tag>
        </template>
      </div>
    </el-form-item>
    <el-form-item v-if="form.reward_type == 1" for="no_click" :label="$t('每次赠送积分')">
      <numInput :min="0.01" :max="100000000" :precision="2" :disabled="status == 1" v-model="form.reward_value" :placeholder="$t('请输入赠送积分数量')"></numInput>
      <div class="gray9">{{ $t('注：满足设置条件规则后，每次所赠送的积分数量') }}</div>
    </el-form-item>
    <el-form-item for="no_click" :label="$t('发送短信通知')" prop="is_send_sms" :rules="[{ required: true, message: $t('请选择是否发送短信通知') }]">
      <el-radio-group v-model="form.is_send_sms">
        <el-radio :value="1">{{ $t('是') }}</el-radio>
        <el-radio :value="0">{{ $t('否') }}</el-radio>
      </el-radio-group>
      <div class="gray9">{{ $t('获得活动奖品后，是否接收短信通知') }}</div>
    </el-form-item>
  </div>
  <SelectCouponDialog :open="openSelectCoupon" v-if="openSelectCoupon" @close="closeSelectCoupon" />
</template>

<script setup>
  import { ref, inject, watch } from 'vue';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';
  import SelectCouponDialog from './dialog.vue';

  // 注入form数据
  const form = inject('form');

  // 响应式数据
  const activityTime = ref(null);
  const openSelectCoupon = ref(false);
  const props = defineProps({
    dateTime: {
      type: Array,
      default: () => [],
    },
    couponList: {
      type: Array,
      default: () => [],
    },
    status: {
      type: Number,
      default: null,
    },
    uuid: {
      type: [String, Number],
      default: null,
    },
  });
  // 引用
  const activityNameFormRef = ref(null);
  const activityDescriptionFormRef = ref(null);
  const couponList = ref([]);
  // 临时选择时间
  const isTodayTemp = ref(false);

  const emit = defineEmits(['imgName', 'imgDescription', 'checkForm']);

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

  watch(
    props.couponList,
    (newVal) => {
      couponList.value = [];
      newVal.map((e) => {
        couponList.value.push({
          name: e.coupon_name,
          uuid: e.prize_uuid,
        });
      });
    },
    {
      immediate: true,
      deep: true,
    }
  );

  watch(activityTime, (newVal) => {
    if (newVal) {
      form.start_time = newVal[0];
      form.end_time = newVal[1];
    } else {
      form.start_time = '';
      form.end_time = '';
    }
  });

  const isToday = (date) => {
    const today = new Date();
    return date.getFullYear() === today.getFullYear() && date.getMonth() === today.getMonth() && date.getDate() === today.getDate();
  };

  const imgName = (data) => {
    emit('imgName', data);
  };

  const imgDescription = (data) => {
    emit('imgDescription', data);
  };

  const selectCoupon = () => {
    if (props.status == 1) {
      return;
    }
    openSelectCoupon.value = true;
  };

  const closeSelectCoupon = (e) => {
    openSelectCoupon.value = false;
    if (e) {
      form.prize_list = [];
      form.prize_list.push({
        prize_type: 1,
        prize_uuid: e.uuid,
      });
      couponList.value = [];
      couponList.value.push(e);
    }
    emit('checkForm', 'prize_list');
  };

  const handleClose = () => {
    if (props.status == 1) {
      return;
    }
    couponList.value = [];
    form.prize_list = [];
    emit('checkForm', 'prize_list');
  };

  // 禁用今天之前的日期
  const disabledDate = (time) => {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    return time.getTime() < today.getTime();
  };

  // 禁用小时
  const dHours = () => {
    const now = new Date();
    const hours = [];
    for (let i = 0; i < now.getHours(); i++) {
      hours.push(i);
    }
    return hours;
  };

  // 禁用分钟
  const dMinutes = () => {
    const now = new Date();
    const currentHour = new Date(activityTime.value[0]).getHours();
    // 如果是当前小时，才禁用当前分钟之前的选项
    if (currentHour === now.getHours()) {
      const minutes = [];
      for (let i = 0; i < now.getMinutes(); i++) {
        minutes.push(i);
      }
      return minutes;
    }
    return [];
  };

  // 禁用秒
  const dSeconds = () => {
    const now = new Date();
    const currentDate = new Date(activityTime.value[0]);
    // 如果是当前小时和分钟，才禁用当前秒之前的选项
    if (currentDate.getHours() === now.getHours() && currentDate.getMinutes() === now.getMinutes()) {
      const seconds = [];
      for (let i = 0; i < now.getSeconds(); i++) {
        seconds.push(i);
      }
      return seconds;
    }
    return [];
  };

  // 禁用小时
  const disabledHours = () => {
    const now = new Date();
    const selectedDate = new Date(form.end_time);
    // 如果是今天，才禁用当前小时之前的时间
    if (isToday(selectedDate)) {
      return Array.from({ length: now.getHours() }, (_, i) => i);
    }
    return [];
  };

  // 禁用分钟
  const disabledMinutes = (hour) => {
    const now = new Date();
    const selectedDate = new Date(form.end_time);
    // 如果是今天且是当前小时，才禁用当前分钟之前的时间
    if (isToday(selectedDate) && hour === now.getHours()) {
      return Array.from({ length: now.getMinutes() }, (_, i) => i);
    }
    return [];
  };

  // 禁用秒
  const disabledSeconds = (hour, minute) => {
    const now = new Date();
    const selectedDate = new Date(form.end_time);
    // 如果是今天且是当前小时和分钟，才禁用当前秒之前的时间
    if (isToday(selectedDate) && hour === now.getHours() && minute === now.getMinutes()) {
      return Array.from({ length: now.getSeconds() }, (_, i) => i);
    }
    return [];
  };

  const handleChange = (value) => {
    // 如果开始时间小于当前时间，则开始时间设置为当前时间
    if (new Date(value[0]) < new Date()) {
      form.start_time = setStartTime();
      activityTime.value[0] = setStartTime();
    }
  };

  const setStartTime = () => {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = String(now.getDate()).padStart(2, '0');
    const hours = String(now.getHours()).padStart(2, '0');
    const minutes = String(now.getMinutes()).padStart(2, '0');
    const seconds = String(now.getSeconds()).padStart(2, '0');
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
  };

  const handleChangeEnd = (value) => {
    if (new Date(value) < new Date()) {
      form.end_time = setStartTime();
    }
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
  :deep(.w100) {
    width: 460px;
  }
</style>

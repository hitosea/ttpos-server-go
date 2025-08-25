<template>
  <el-dialog :title="title" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" ref="formRef" :model="form" label-position="top">
      <div class="common-form">
        {{ $t('基础信息') }}
      </div>
      <!--店员修改-->
      <el-form-item
        for="no_click"
        :label="$t('名称')"
        prop="name"
        :rules="[
          { required: true, message: $t('请输入优惠券名称') },
          {
            validator: (rule, value, callback) => {
              if (value.trim() === '') {
                callback(new Error($t('请输入优惠券名称')));
              }
              callback();
            },
            trigger: 'blur',
          },
        ]"
      >
        <el-input class="percent-w100" v-model="form.name" :maxlength="50" :placeholder="$t('请输入优惠券名称')"></el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('排序')" prop="sort">
        <numInput width="m-full" :min="1" :max="99" :precision="0" :placeholder="$t('接近0，排序等级越高')" v-model="form.sort"></numInput>
      </el-form-item>
      <div class="common-form mt24">
        {{ $t('优惠券设置') }}
      </div>
      <el-form-item for="no_click" :label="$t('优惠券类型')" prop="type" :rules="[{ required: true, message: $t('请选择优惠券类型') }]">
        <el-radio-group v-model="form.type">
          <el-radio label="deduction">{{ $t('折扣券') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('抵扣类型')" prop="deduction_type" :rules="[{ required: true, message: $t('请选择抵扣类型') }]">
        <el-radio-group v-model="form.deduction_type">
          <el-radio label="taxed">{{ $t('税后折扣') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('金额')" prop="amount" :rules="[{ required: true, message: $t('请输入优惠券金额') }]">
        <numInput width="m-full" :min="1" :max="999999" :precision="2" v-model="form.amount" :placeholder="$t('请输入优惠券金额')"></numInput>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('数量')" prop="count" :rules="[{ required: true, message: $t('请输入优惠券数量') }]">
        <numInput width="m-full" :min="1" :max="999999" :precision="0" :placeholder="$t('请输入优惠券数量')" v-model="form.count"></numInput>
      </el-form-item>
      <el-form-item
        for="no_click"
        :label="$t('适用时间')"
        prop="day_start_time"
        :rules="[
          {
            required: true,
            validator: () => {
              return (form.day_start_time && form.day_end_time) || use_time_type == 1 ? true : false;
            },
            message: $t('请选择适用时间'),
          },
        ]"
      >
        <el-radio-group v-model="use_time_type" class="time-box-radio" @change="handleUseTimeTypeChange">
          <el-radio :label="1">{{ $t('全天') }}</el-radio>
          <el-radio :label="2">
            {{ $t('特定时间段') }}
            <el-time-picker
              v-if="use_time_type == 2"
              class="time-picker"
              v-model="day_time"
              format="HH:mm"
              value-format="HH:mm"
              is-range
              range-separator="~"
              :start-placeholder="$t('开始时间')"
              :end-placeholder="$t('结束时间')"
              @change="handleDayTimeChange"
            />
          </el-radio>
        </el-radio-group>
      </el-form-item>
      <div class="common-form mt24">
        {{ $t('优惠券条件') }}
      </div>
      <el-form-item
        for="no_click"
        :label="$t('获取条件')"
        prop="requirement"
        :rules="[
          {
            required: true,
            validator: () => {
              return (form.requirement == 'none' && reg_date.length > 0) || (form.requirement == 'marketing' && form.valid_days > 0) ? true : false;
            },
            message: $t('请选择获取条件'),
          },
        ]"
      >
        <el-radio-group v-model="form.requirement" class="flex-box-radio" :disabled="props.editData.uuid">
          <el-radio label="none">
            {{ $t('所有人可用') }}
            <el-date-picker
              v-if="form.requirement == 'none'"
              v-model="reg_date"
              type="daterange"
              value-format="YYYY-MM-DD"
              range-separator="~"
              :start-placeholder="$t('开始日期')"
              :end-placeholder="$t('结束日期')"
              :disabledDate="props.editData.uuid ? disabledDateEdit : disabledDate"
              clearable
              @change="handleRegDateChange"
            ></el-date-picker>
          </el-radio>
          <el-radio label="marketing">
            {{ $t('通过营销活动获取') }}
            <div class="flex-box" v-if="form.requirement == 'marketing'">
              <div class="span-text-left"> {{ $t('活动奖励后') }}</div>
              <el-input-number :controls="false" :min="0" :max="999" :precision="0" v-model="form.valid_days" :placeholder="$t('请输入')"></el-input-number>
              <div class="span-text">{{ $t('个自然日内有效') }}</div>
            </div>
          </el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button size="small" @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button size="small" type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
  import { ref, reactive, watch, getCurrentInstance } from 'vue';
  import { ElMessage } from 'element-plus';
  import { useUserStore } from '@/store';
  import MarketingApi from '@/api/marketing.js';
  // 获取全局属性
  const { proxy } = getCurrentInstance();
  const { $t } = proxy;

  // 获取store
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;

  // Props
  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '',
    },
    editData: {
      type: Object,
      default: null,
    },
  });

  // Emits
  const emit = defineEmits(['closeDialog']);

  // 响应式数据
  const dialogVisible = ref(false);
  const formRef = ref();
  const loading = ref(false);

  const form = reactive({
    name: '',
    sort: 1,
    type: 'deduction',
    deduction_type: 'taxed',
    amount: null,
    count: null,
    requirement: 'none',
    day_start_time: '00:00',
    day_end_time: '23:59',
    valid_start_time: '',
    valid_end_time: '',
    valid_days: null,
  });

  const use_time_type = ref(1);
  const reg_date = ref([]);
  const day_time = ref([]);

  // 监听props变化
  watch(
    () => props.open,
    (newVal) => {
      dialogVisible.value = newVal;
      if (newVal) {
        if (props.editData) {
          // 复制编辑数据到表单
          Object.keys(form).forEach((key) => {
            form[key] = props.editData[key];
          });

          // 处理时间显示
          if (props.editData.day_start_time === '00:00' && props.editData.day_end_time === '23:59') {
            use_time_type.value = 1;
            day_time.value = [];
          } else {
            use_time_type.value = 2;
            // 确保时间格式正确（添加空格以匹配format格式）
            day_time.value = [props.editData.day_start_time ? ` ${props.editData.day_start_time}` : null, props.editData.day_end_time ? ` ${props.editData.day_end_time}` : null];
          }

          // 处理日期范围
          if (props.editData.valid_start_time && props.editData.valid_end_time) {
            reg_date.value = [props.editData.valid_start_time, props.editData.valid_end_time];
          }
        }
      }
    },
    { immediate: true }
  );

  // 方法
  const onSubmit = () => {
    if (props.editData) {
      const params = {
        ...form,
      };
      params.uuid = props.editData.uuid;
      formRef.value.validate((valid) => {
        if (valid) {
          loading.value = true;
          MarketingApi.couponEdit(params, true)
            .then((data) => {
              loading.value = false;
              ElMessage({
                message: $t('保存成功'),
                type: 'success',
              });
              dialogFormVisible(1);
            })
            .catch((error) => {
              loading.value = false;
            });
        }
      });
    } else {
      const params = { ...form };
      formRef.value.validate((valid) => {
        if (valid) {
          loading.value = true;
          MarketingApi.couponAddGet(params, true)
            .then((data) => {
              loading.value = false;
              ElMessage({
                message: $t('添加成功'),
                type: 'success',
              });
              dialogFormVisible(1);
            })
            .catch((error) => {
              loading.value = false;
            });
        }
      });
    }
  };

  const dialogFormVisible = (e) => {
    dialogVisible.value = false;
    emit('closeDialog', e);
  };

  const handleUseTimeTypeChange = (val) => {
    if (val === 1) {
      // 选择全天时，设置默认时间
      form.day_start_time = '00:00';
      form.day_end_time = '23:59';
      day_time.value = [];
    } else {
      // 选择特定时间段时，如果已有时间则使用已有时间，否则清空
      if (form.day_start_time && form.day_end_time && !(form.day_start_time === '00:00' && form.day_end_time === '23:59')) {
        day_time.value = [form.day_start_time, form.day_end_time];
      } else {
        form.day_start_time = '';
        form.day_end_time = '';
        day_time.value = [];
      }
    }
  };

  const handleRegDateChange = (e) => {
    form.valid_start_time = e[0];
    form.valid_end_time = e[1];
  };

  const handleDayTimeChange = (val) => {
    if (val && val.length === 2) {
      // 去掉可能存在的前导空格
      form.day_start_time = val[0].trim();
      form.day_end_time = val[1].trim();
    } else {
      form.day_start_time = '';
      form.day_end_time = '';
    }
  };

  const handleCountChange = (val) => {
    if (val < 0) {
      form.count = 0;
      return;
    }
    if (val > 100000000) {
      form.count = 100000000;
      return;
    }
    form.count = val;
  };

  // 禁用今天之前的日期
  const disabledDate = (time) => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return time.getTime() < today.getTime();
  };

  const disabledDateEdit = (time) => {
    // props.editData.valid_start_time 前一天 与 现在时间比较，谁比较靠前，按那个日期禁用
    const valid_start_time = new Date(props.editData.valid_start_time);
    valid_start_time.setDate(valid_start_time.getDate() - 1);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return time.getTime() < valid_start_time.getTime() || !valid_start_time.getTime() ? time.getTime() < today.getTime() : time.getTime() < valid_start_time.getTime();
  };
</script>

<style scoped lang="scss">
  .w-full {
    width: 100%;
  }

  .mt24 {
    margin-top: 24px;
  }

  .common-form {
    font-weight: 600;
    margin-bottom: 16px;
  }

  :deep(.flex-box-radio.el-radio-group) {
    flex-wrap: nowrap;
    font-display: auto;
    flex-direction: column;
    width: 100%;
    justify-content: flex-start;
    align-items: flex-start;
    gap: 10px;
    .el-radio.el-radio--small {
      height: auto;
      width: 100%;
      justify-content: flex-start;
      align-items: flex-start;
      .el-radio__label {
        width: 100%;
        gap: 10px;
        display: flex;
        flex-direction: column;
      }
      .el-radio__input {
        margin-top: 10px !important;
      }
    }
  }
  .time-box-radio {
    width: 100%;
    .el-radio.el-radio--small {
      height: auto;
    }
  }
  .flex-box {
    display: flex;
    width: 100%;
    gap: 10px;
    color: #333;
  }
  :deep(.time-picker) {
    margin-left: 10px !important;
  }
</style>

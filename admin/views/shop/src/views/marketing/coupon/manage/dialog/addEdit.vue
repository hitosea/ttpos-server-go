<template>
  <el-dialog :title="title" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" ref="formRef" :model="form" label-position="top">
      <div class="common-form">
        {{ $t('基础信息') }}
      </div>
      <!--店员修改-->
      <el-form-item for="no_click" :label="$t('名称')" prop="nick_name" :rules="[{ required: true, message: $t('请输入昵称') }]">
        <el-input class="percent-w100" v-model="form.nick_name" :maxlength="50" :placeholder="$t('请输入昵称')"></el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('排序')" prop="sort">
        <el-input-number :controls="false" :min="0" :max="999" :placeholder="$t('接近0，排序等级越高')" v-model.number="form.sort"></el-input-number>
      </el-form-item>
      <div class="common-form mt24">
        {{ $t('优惠券设置') }}
      </div>
      <el-form-item for="no_click" :label="$t('优惠券类型')" prop="coupon_type" :rules="[{ required: true, message: $t('请选择优惠券类型') }]">
        <el-radio-group v-model="form.coupon_type">
          <el-radio :label="1">{{ $t('折扣券') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('抵扣类型')" prop="deduct_type" :rules="[{ required: true, message: $t('请选择抵扣类型') }]">
        <el-radio-group v-model="form.deduct_type">
          <el-radio :label="1">{{ $t('税后折扣') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('金额')" prop="service_charge" :rules="[{ required: true, message: $t('请选择金额') }]">
        <numInput width="m-full" :min="0" :precision="2" v-model:valueData="form.service_charge" :value="form.service_charge" :placeholder="$t('请输入')"></numInput>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('数量')" prop="quantity" :rules="[{ required: true, message: $t('请选择数量') }]">
        <el-input class="percent-w100" v-model="form.quantity" :placeholder="$t('请输入数量')"></el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('适用时间')" prop="use_time_type" :rules="[{ required: true, message: $t('请选择适用时间') }]">
        <el-radio-group v-model="form.use_time_type">
          <el-radio :label="1">{{ $t('全天') }}</el-radio>
          <el-radio :label="2">{{ $t('特定时间段') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <div class="common-form mt24">
        {{ $t('优惠券条件') }}
      </div>
      <el-form-item for="no_click" :label="$t('适用人群')" prop="user_type" :rules="[{ required: true, message: $t('请选择适用人群') }]">
        <el-radio-group v-model="form.user_type">
          <el-radio :label="1">{{ $t('所有人可用') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('有效时间')" prop="reg_date" :rules="[{ required: true, message: $t('请选择有效时间') }]">
        <el-date-picker
          v-model="form.reg_date"
          type="daterange"
          value-format="YYYY-MM-DD"
          range-separator="~"
          :start-placeholder="$t('开始日期')"
          :end-placeholder="$t('结束日期')"
          clearable
        ></el-date-picker>
      </el-form-item>

      <div class="common-form mt24">
        {{ $t('获取途径') }}
      </div>
      <el-form-item for="no_click" :label="$t('获取途径')" prop="get_type" :rules="[{ required: true, message: $t('请选择获取途径') }]">
        <el-radio-group v-model="form.get_type">
          <el-radio :label="1">{{ $t('通过营销活动获取') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item class="flex-box" for="no_click" :label="$t('有效日期')" prop="activity_after" :rules="[{ required: true, message: $t('请选择有效日期') }]">
        <div class="span-text-left"> {{ $t('活动奖励后') }}</div>
        <el-input-number :controls="false" :min="0" :max="999" :precision="0" v-model="form.activity_after" :placeholder="$t('请输入')"></el-input-number>
        <div class="span-text">{{ $t('个自然日内有效') }}</div>
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
  import UserApi from '@/api/user.js';
  import { useUserStore } from '@/store';

  // 获取全局属性
  const { proxy } = getCurrentInstance();
  const { $t } = proxy;

  // 获取store
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;

  // Props
  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
    editform: {
      type: Object,
      default: () => ({}),
    },
    title: {
      type: String,
      default: '',
    },
    gradeList: {
      type: Array,
      default: () => [],
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
    nick_name: '',
    sort: 0,
    coupon_type: 1,
    deduct_type: 1,
    service_charge: 0,
    quantity: '',
    use_time_type: 1,
    user_type: 1,
    reg_date: '',
    get_type: 1,
    activity_after: 0,
    gender: 2,
    mobile: '',
    grade_id: 1,
    password: '',
    birthday: '',
  });

  const gradeSelectList = ref([]);

  // 监听props变化
  watch(
    () => props.open,
    (newVal) => {
      dialogVisible.value = newVal;
      if (newVal) {
        if (props.editData) {
          Object.assign(form, JSON.parse(JSON.stringify(props.editData)));
          form.nick_name = props.editData.nickName;
        }

        if (props.gradeList && props.gradeList.length > 0) {
          gradeSelectList.value = props.gradeList.map((item) => ({
            grade_id: item.grade_id,
            name: item.name,
          }));
        }
      }
    },
    { immediate: true }
  );

  // 方法
  const onSubmit = () => {
    if (props.editData) {
      const params = {
        user_id: form.user_id,
        nick_name: form.nick_name,
        gender: form.gender,
        grade_id: form.grade_id,
        mobile: form.mobile,
        password: form.password,
        birthday: form.birthday,
      };

      formRef.value.validate((valid) => {
        if (valid) {
          loading.value = true;
          UserApi.edituser(params, true)
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
          UserApi.adduser(params, true)
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

  .flex-box {
    :deep(.el-form-item__content) {
      flex-wrap: nowrap;
      flex-direction: row;
    }
    .span-text-left {
      margin-right: 10px;
      flex-shrink: 0;
    }
    .span-text {
      margin-left: 10px;
      flex-shrink: 0;
    }
  }
</style>

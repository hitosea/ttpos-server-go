<template>
  <el-dialog width="960" :title="$t('同步支付方式')" :modelValue="props.show" :close-on-click-modal="false" :close-on-press-escape="false" align-center @close="handleClose()">
    <div class="max-h-[75vh] overflow-auto pr-4" v-if="props.show">
      <el-form :model="formData" :rules="formRules" ref="formRef" label-position="top" label-width="auto">
        <!-- 请选择 -->
        <el-form-item :label="$t('请选择')" prop="erpnext_payment">
          <el-select v-model="formData.erpnext_payment" :placeholder="$t('支付方式')" clearable filterable @change="handleErpnextPaymentChange">
            <el-option v-for="item in erpPayMethodList" :key="item.name" :value="item.name" :label="item.name" :disabled="!item.is_addable" />
          </el-select>
        </el-form-item>
        <template v-if="formData.erpnext_payment">
          <!-- 名称 -->
          <el-form-item :label="$t('名称')" prop="name">
            <el-input v-model="formData.name" :placeholder="$t('请输入名称')" clearable maxlength="100" />
          </el-form-item>

          <!-- 支付方式 -->
          <el-form-item :label="$t('支付方式')" prop="erpnext_payment">
            <el-select v-model="formData.erpnext_payment" :placeholder="$t('回显erp中的支付方式名称')" disabled>
              <el-option v-for="item in erpPayMethodList" :key="item.name" :value="item.name" :label="item.name" />
            </el-select>
          </el-form-item>

          <!-- 手续费 -->
          <el-form-item :label="$t('手续费')" prop="fee">
            <el-input-number class="!w-full" v-model="formData.fee" :placeholder="$t('默认0')" :min="0" :max="100" clearable :controls="false">
              <template #suffix>%</template>
            </el-input-number>
          </el-form-item>

          <!-- 排序 -->
          <el-form-item :label="$t('排序')" prop="sort">
            <el-input-number class="!w-full" v-model="formData.sort" :placeholder="$t('接近0，排序等级越高')" :min="0" :max="999" :precision="0" clearable :controls="false" />
          </el-form-item>

          <!-- 结账显示 -->
          <el-form-item :label="$t('结账显示')" prop="checkout_show">
            <el-checkbox-group v-model="formData.checkout_show">
              <el-checkbox value="cashier">{{ $t('收银机') }}</el-checkbox>
              <el-checkbox value="assistant">{{ $t('点餐助手') }}</el-checkbox>
            </el-checkbox-group>
          </el-form-item>

          <!-- 会员充值 -->
          <el-form-item :label="$t('会员充值')" prop="member_recharge_show" v-if="formData.erpnext_payment != 'Balance'">
            <el-checkbox-group v-model="formData.member_recharge_show">
              <el-checkbox value="cashier">{{ $t('收银机') }}</el-checkbox>
            </el-checkbox-group>
          </el-form-item>

          <!-- 状态 -->
          <el-form-item :label="$t('状态')" prop="status">
            <el-switch v-model="formData.status" :active-value="1" :inactive-value="0" />
          </el-form-item>
        </template>
      </el-form>
    </div>

    <template #footer>
      <el-button @click="handleClose">{{ $t('取消') }}</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="isLoading">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref, defineProps, defineEmits, watch } from 'vue';
  import { ElMessage, FormInstance } from 'element-plus';
  import { getErpnextPaymentMethodList, erpnextPaymentMethodListType, erpnextAddPaymentMethod, erpnextAddPaymentMethodParams } from '@/api/authorization';
  import { $t } from '@/i18n';
  // 定义组件属性
  const props = defineProps({
    show: {
      type: Boolean,
      default: false,
    },
    editRow: {
      type: Object,
      default: () => ({}),
    },
  });

  // 定义事件
  const emit = defineEmits(['update:show', 'refresh']);

  // 表单引用
  const formRef = ref<FormInstance>();

  // 加载状态
  const isLoading = ref(false);

  // 名称错误状态（用于演示错误提示）
  const nameError = ref(false);

  // ERP支付方式列表
  const erpPayMethodList = ref<erpnextPaymentMethodListType[]>([]);

  // 表单数据
  const formData = ref<erpnextAddPaymentMethodParams>({
    company_uuid: 0, // ERP支付方式
    name: '', // 名称
    erpnext_payment: '', // 支付方式
    fee: undefined, // 手续费
    sort: 0, // 排序
    checkout_show: ['cashier', 'assistant'], // 结账显示（复选框）
    member_recharge_show: ['cashier'], // 会员充值（复选框）
    status: 1, // 状态
  });

  // 表单验证规则
  const formRules = ref({
    company_uuid: [{ required: true, message: $t('请选择ERP支付方式') }],
    name: [
      { required: true, message: $t('请输入名称') },
      { max: 100, message: $t('名称不能超过100个字符') },
    ],
    erpnext_payment: [{ required: true, message: $t('请选择支付方式') }],
    fee: [
      {
        required: true,
        message: $t('请输入手续费'),
      },
      {
        validator: (_rule: any, value: string, callback: any) => {
          if (value !== '' && value !== null && value !== undefined) {
            const num = parseFloat(value);
            if (isNaN(num) || num < 0 || num > 100) {
              callback(new Error($t('手续费必须在0-100之间')));
            } else {
              callback();
            }
          } else {
            callback();
          }
        },
        trigger: 'blur',
      },
    ],
    sort: [
      {
        required: true,
        message: $t('请输入排序'),
      },
      {
        validator: (_rule: any, value: string, callback: any) => {
          if (value !== '' && value !== null && value !== undefined) {
            const num = parseInt(value);
            if (isNaN(num) || num < 0 || num > 999) {
              callback(new Error($t('排序值必须在0-999之间')));
            } else {
              callback();
            }
          } else {
            callback();
          }
        },
        trigger: 'blur',
      },
    ],
  });

  // 关闭弹窗
  const handleClose = () => {
    emit('update:show', false);
    handleReset();
  };

  // 重置表单
  const handleReset = () => {
    formData.value = {
      company_uuid: 0, // ERP支付方式
      name: '', // 名称
      erpnext_payment: '', // 支付方式
      fee: undefined, // 手续费
      sort: 0, // 排序
      checkout_show: ['cashier'], // 结账显示（复选框）
      member_recharge_show: ['cashier'], // 会员充值（复选框）
      status: 1, // 状态
    };
    nameError.value = false;
    formRef.value?.clearValidate();
  };

  // 提交表单
  const handleSubmit = async () => {
    formRef.value?.validate(async (valid) => {
      if (valid) {
        try {
          isLoading.value = true;
          const res = await erpnextAddPaymentMethod(formData.value);
          ElMessage.success(res.msg);
          handleClose();
          emit('refresh');
        } catch (error) {
          console.log(error);
        } finally {
          isLoading.value = false;
        }
      }
    });
  };

  const getErpnextPaymentMethodListData = async () => {
    try {
      const res = await getErpnextPaymentMethodList({ company_uuid: props.editRow.uuid });
      erpPayMethodList.value = res.data.list;
    } catch (error) {
      console.log(error);
    }
  };

  const handleErpnextPaymentChange = (value: string) => {
    if (value == 'Balance') {
      formData.value.member_recharge_show = [];
    } else {
      formData.value.member_recharge_show = ['cashier'];
    }
  };

  // 监听弹窗显示状态
  watch(
    () => props.show,
    (newVal: boolean) => {
      if (newVal) {
        getErpnextPaymentMethodListData();
        // 如果有编辑数据，填充表单
        if (props.editRow && Object.keys(props.editRow).length > 0) {
          formData.value.company_uuid = props.editRow.uuid;
        }
        // 清除验证状态
        formRef.value?.clearValidate();
      }
    },
    { immediate: true, deep: true },
  );
</script>

<style lang="scss" scoped>
  .image-uploader {
    :deep(.el-upload) {
      border: 1px dashed var(--el-border-color);
      border-radius: 6px;
      cursor: pointer;
      position: relative;
      overflow: hidden;
      transition: var(--el-transition-duration-fast);
      width: 64px;
      height: 64px;
      display: flex;
      align-items: center;
      justify-content: center;

      &:hover {
        border-color: var(--el-color-primary);
      }
    }
  }

  .image-uploader-icon {
    font-size: 24px;
    color: #8c939d;
  }

  .uploaded-image {
    width: 64px;
    height: 64px;
    object-fit: cover;
  }

  .qrcode-uploader {
    :deep(.el-upload) {
      border: 1px dashed var(--el-border-color);
      border-radius: 6px;
      cursor: pointer;
      position: relative;
      overflow: hidden;
      transition: var(--el-transition-duration-fast);
      width: 64px;
      height: 64px;
      display: flex;
      align-items: center;
      justify-content: center;

      &:hover {
        border-color: var(--el-color-primary);
      }
    }
  }

  .qrcode-uploader-icon {
    font-size: 28px;
    color: #8c939d;
  }

  .uploaded-qrcode {
    width: 64px;
    height: 64px;
    object-fit: cover;
  }
</style>

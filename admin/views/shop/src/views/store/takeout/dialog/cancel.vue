<template>
  <!--
      
      时间：2019-10-25
      描述：会员-用户列表-会员充值
  -->
  <div>
    <el-dialog :title="$t('取消订单')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
      <el-form size="small" ref="formRef" :model="form" label-position="top">
        <el-form-item for="no_click" :label="$t('订单号')" :label-width="formLabelWidth" prop="order_no" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.order_no" disabled :placeholder="$t('请输入订单号')" :readonly="true"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('备注')" :label-width="formLabelWidth" prop="cancel_reason" :rules="[{ required: true, message: $t('请输入备注') }]">
          <el-input type="textarea" v-model="form.cancel_reason" :placeholder="$t('请输入备注')"></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ElMessage } from 'element-plus';
  import OrderApi from '@/api/order.js';

  // 使用国际化
  const { t: $t } = useI18n();

  // 定义props
  const props = defineProps({
    open_cancel: {
      type: Boolean,
      default: false,
    },
    member_sale_order_uuid: {
      type: [String, Number],
      default: '',
    },
    order_no: {
      type: [String, Number],
      default: '',
    },
  });

  // 定义emits
  const emit = defineEmits(['closeDialog']);

  // 响应式数据
  const loading = ref(false);
  const formLabelWidth = ref('120px');
  const dialogVisible = ref(false);

  const form = reactive({
    member_sale_order_uuid: '',
    order_no: '',
    cancel_reason: '',
  });

  // 表单引用
  const formRef = ref(null);

  // 生命周期
  onMounted(() => {
    dialogVisible.value = props.open_cancel;
    form.member_sale_order_uuid = props.member_sale_order_uuid;
    form.order_no = props.order_no;
  });

  // 方法定义
  const submit = () => {
    const formData = {
      member_sale_order_uuid: form.member_sale_order_uuid,
      cancel_reason: form.cancel_reason,
    };

    formRef.value.validate((valid) => {
      if (valid) {
        loading.value = true;
        OrderApi.postTakeoutOrderCancel(formData, true)
          .then((data) => {
            loading.value = false;
            ElMessage({
              message: $t('操作成功'),
              type: 'success',
            });
            dialogFormVisible(true);
          })
          .catch((error) => {
            loading.value = false;
          });
      }
    });
  };

  const dialogFormVisible = (e) => {
    if (e) {
      emit('closeDialog', {
        type: 'success',
        openDialog: false,
      });
    } else {
      emit('closeDialog', {
        type: 'error',
        openDialog: false,
      });
    }
  };
</script>

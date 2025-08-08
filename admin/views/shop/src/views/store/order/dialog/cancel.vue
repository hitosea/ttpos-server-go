<template>
    <div>
        <el-dialog :title="$t('取消订单')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
            <el-form size="small" ref="formRef" :model="form" label-position="top">
                <el-form-item for="no_click" :label="$t('订单号')" :label-width="formLabelWidth" prop="order_no" :rules="[{ required: true, message: ' ' }]">
                    <el-input v-model="form.order_no" disabled :placeholder="$t('请输入订单号')" :readonly="true"></el-input>
                </el-form-item>
                <el-form-item for="no_click" :label="$t('备注')" :label-width="formLabelWidth" prop="cancel_remark" :rules="[{ required: true, message: $t('请输入备注') }]">
                    <el-input type="textarea" v-model="form.cancel_remark" :placeholder="$t('请输入备注')"></el-input>
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
import { ref, reactive, watch, getCurrentInstance, onMounted } from 'vue';
import OrderApi from '@/api/order.js';

const props = defineProps({
  open_edit: { type: Boolean, default: false },
  order_no: { type: [String, Number], default: '' },
  order_id: { type: [String, Number], default: '' },
});
const emit = defineEmits(['closeDialog']);

const { proxy } = getCurrentInstance();

const loading = ref(false);
const formLabelWidth = ref('120px');
const dialogVisible = ref(false);
const formRef = ref();

const form = reactive({
  order_id: '',
  cancel_remark: '',
  order_no: '',
});

watch(
  () => [props.open_edit, props.order_no, props.order_id],
  ([open, no, id]) => {
    dialogVisible.value = open;
    form.order_no = String(no ?? '');
    form.order_id = String(id ?? '');
  },
  { immediate: true }
);

onMounted(() => {
  if (form.order_no === '' && props.order_no) form.order_no = String(props.order_no);
  if (form.order_id === '' && props.order_id) form.order_id = String(props.order_id);
});

// 提交处理
const submit = async () => {
  const valid = await formRef.value.validate();
  if (!valid) return;

  loading.value = true;
  try {
    const data = await OrderApi.storeConfirm({ ...form }, true);
    proxy.$ElMessage({
      message: data.msg,
      type: 'success',
    });
    dialogFormVisible(true);
  } catch (e) {
    // 忽略错误，loading 在 finally 关闭
  } finally {
    loading.value = false;
  }
};

// 关闭弹窗
const dialogFormVisible = (e = false) => {
  emit('closeDialog', {
    type: e ? 'success' : 'error',
    openDialog: false,
  });
};
</script>

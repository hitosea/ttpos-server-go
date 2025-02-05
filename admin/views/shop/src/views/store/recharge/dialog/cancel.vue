<template>
  <div>
    <el-dialog :title="$t('取消订单')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
      <el-form size="small" ref="formRef" :model="form" label-position="top">
        <el-form-item for="no_click" :label="$t('订单号')" prop="order_no" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.order_no" disabled :placeholder="$t('请输入订单号')" :readonly="true"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('备注')" prop="cancel_remark" :rules="[{ required: true, message: $t('请输入备注') }]">
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
  import { ref, reactive } from 'vue';
  import OrderApi from '@/api/order.js';
  import { message } from '@/utils/message.js';

  const emits = defineEmits(['closeDialog']);

  const props = defineProps({
    open_edit: Boolean,
    order_no: String,
    id: String,
  });

  const loading = ref(false);
  const dialogVisible = ref(props.open_edit);
  const form = reactive({
    id: props.id,
    cancel_remark: '',
    order_no: props.order_no,
  });

  const formRef = ref(null);

  const submit = async () => {
    formRef.value.validate(async (valid) => {
      if (valid) {
        loading.value = true;
        try {
          const params = {
            id: form.id,
            cancel_remark: form.cancel_remark,
          };
          const res = await OrderApi.getRechargeOrderCancel(params, true);
          loading.value = false;
          message({
            message: res.msg,
            type: 'success',
          });
          dialogFormVisible(true);
        } catch (error) {
          loading.value = false;
        }
      }
    });
  };

  const dialogFormVisible = (e) => {
    emits('closeDialog', {
      type: e ? 'success' : 'error',
      openDialog: false,
    });
  };
</script>

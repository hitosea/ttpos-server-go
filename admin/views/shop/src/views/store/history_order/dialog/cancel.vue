<template>
  <div>
    <el-dialog :title="$t('取消订单')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
      <el-form size="small" ref="form" :model="form" label-position="top">
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
  import { ref, reactive, onMounted, getCurrentInstance, watch } from 'vue';
  import { ElMessage } from 'element-plus';
  import OrderOldApi from '@/api/orderOld.js';


  // 获取Vue实例
  const { proxy } = getCurrentInstance();

  // 定义props
  const props = defineProps({
    open_edit: {
      type: Boolean,
      default: false,
    },
    order_no: {
      type: [String, Number],
      default: '',
    },
    order_id: {
      type: [String, Number],
      default: '',
    },
  });

  // 定义emits
  const emit = defineEmits(['closeDialog']);

  // 响应式变量
  const loading = ref(false);
  const formLabelWidth = ref('120px'); // 左边长度
  const dialogVisible = ref(false); // 是否显示
  
  // 表单数据
  const form = reactive({
    order_id: '',
    cancel_remark: '',
    order_no: '',
  });

  // 监听props变化
  watch(
    () => props.open_edit,
    (newVal) => {
      dialogVisible.value = newVal;
    }
  );

  // 生命周期 - 组件创建时
  onMounted(() => {
    dialogVisible.value = props.open_edit;
    form.order_no = props.order_no;
    form.order_id = props.order_id;
  });

  // 方法定义
  
  // 处理提交
  const submit = async () => {
    try {
      const valid = await proxy.$refs.form.validate();
      if (valid) {
        loading.value = true;
        const data = await OrderOldApi.storeConfirm(form, true);
        loading.value = false;
        ElMessage({
          message: data.msg,
          type: 'success',
        });
        dialogFormVisible(true);
      }
    } catch (error) {
      loading.value = false;
    }
  };

  // 关闭弹窗
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

<template>
  <el-dialog title="订单价格修改" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false" width="30%">
    <el-form size="small" :model="order" ref="order">
      <el-form-item for="no_click" label="订单金额" :label-width="formLabelWidth" prop="update_price" :rules="[{ required: true, message: ' ' }]">
        <el-input type="number" v-model="order.update_price" autocomplete="off"></el-input>
        <p>最终付款价 = 订单金额 + 运费金额</p>
      </el-form-item>
      <el-form-item for="no_click" label="运费金额" :label-width="formLabelWidth" prop="update_express_price" :rules="[{ required: true, message: ' ' }]">
        <el-input type="number" v-model="order.update_express_price" autocomplete="off"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">取 消</el-button>
        <el-button type="primary" @click="submitFunc()" :loading="loading">确 定</el-button>
      </div>
    </template>
  </el-dialog>
</template>
<script setup>
  import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
  import { useRoute } from 'vue-router';
  import { ElMessage } from 'element-plus';
  import OrderApi from '@/api/order.js';

  // 获取Vue实例和路由
  const { proxy } = getCurrentInstance();
  const route = useRoute();

  // 定义props
  const props = defineProps({
    open_edit: {
      type: Boolean,
      default: false,
    },
  });

  // 定义emits
  const emit = defineEmits(['close']);

  // 响应式变量
  const order_id = ref(0);
  const loading = ref(false);
  const formLabelWidth = ref('100px'); // 左边长度
  const dialogVisible = ref(true); // 是否显示

  // 表单数据
  const order = reactive({
    update_price: 0,
    update_express_price: 0.0,
  });

  // 生命周期 - 组件创建时
  onMounted(() => {
    order_id.value = route.query.order_id;
    getData();
  });

  // 方法定义

  // 获取数据
  const getData = async () => {
    try {
      const data = await OrderApi.orderdetail(
        {
          order_id: order_id.value,
        },
        true
      );
      loading.value = false;
      order.update_price = data.data.detail.pay_price;
    } catch (error) {
      loading.value = false;
    }
  };

  // 确认事件
  const submitFunc = async (e) => {
    try {
      const valid = await proxy.$refs.order.validate();
      if (valid) {
        loading.value = true;
        await OrderApi.updatePrice(
          {
            order_id: order_id.value,
            order: order,
          },
          true
        );
        loading.value = false;
        ElMessage({
          message: '保存成功',
          type: 'success',
        });
        dialogFormVisible();
      }
    } catch (error) {
      loading.value = false;
    }
  };

  // 关闭弹窗
  const dialogFormVisible = () => {
    emit('close', { openDialog: false });
  };
</script>

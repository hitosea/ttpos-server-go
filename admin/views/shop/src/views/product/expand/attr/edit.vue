<!-- 描述：商品-属性库-编辑属性 -->
<template>
  <el-dialog :title="$t('编辑属性')" v-model="dialogVisible" @close="handleClose" :close-on-click-modal="false" :close-on-press-escape="false" append-to-body>
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <UniqueNameForm
        ref="uniqueNameFormRef"
        :labelPrefix="$t('属性名称')"
        apiSource="attribute"
        :apiId="form.attribute_id"
        :overrideLanguages="form.attribute_name"
        :parent_id="editform.parent_id || 0"
      />
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">{{ $t('取消') }}</el-button>
        <el-tooltip effect="dark" placement="top" :content="$t('如长时间无响应，请刷新后重试。')">
          <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
        </el-tooltip>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
import ProductApi from '@/api/product.js';
import UniqueNameForm from '@/components/product/UniqueNameForm.vue';

// 获取组件实例
const { proxy } = getCurrentInstance();

// 定义props
const props = defineProps({
  open_edit: {
    type: Boolean,
    default: false,
  },
  editform: {
    type: Object,
    default: () => ({}),
  },
});

// 定义emits
const emit = defineEmits(['closeDialog']);

// 响应式数据
const formRef = ref(null);
const uniqueNameFormRef = ref(null);
const dialogVisible = ref(false);
const loading = ref(false);

const form = reactive({
  attribute_id: undefined,
  attribute_name: {},
});

// 初始化数据
onMounted(() => {
  dialogVisible.value = props.open_edit;

  form.attribute_id = props.editform.attribute_id;
  try {
    const _names = typeof props.editform.attribute_name === 'string' ? JSON.parse(props.editform.attribute_name) : props.editform.attribute_name ?? {};
    form.attribute_name = _names;
  } catch (error) {
    console.error('parse name failed', error);
  }
});

// 提交方法
const submit = async () => {
  loading.value = true;
  try {
    const validForm = await formRef.value.validate();
    if (!validForm) return;

    const validUniqueName = await uniqueNameFormRef.value.validate();
    if (!validUniqueName) return;

    const _name = uniqueNameFormRef.value.data;
    const params = JSON.parse(JSON.stringify(form));
    params.attribute_name = JSON.stringify(_name);

    const res = await ProductApi.editAttribute(params, true);
    proxy.$ElMessage({
      message: $t('保存成功'),
      type: 'success',
    });

    handleClose(true, res.data);
  } catch (error) {
    console.error(error);
  } finally {
    loading.value = false;
  }
};

// 关闭弹窗
const handleClose = (isSuccess = false, data) => {
  emit('closeDialog', {
    type: isSuccess ? 'success' : 'error',
    openDialog: false,
    data: data,
  });
};
</script>

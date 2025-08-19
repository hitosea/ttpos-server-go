<template>
  <el-dialog :title="$t('添加特色分类')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <UniqueNameForm ref="uniqueNameFormRef" :labelPrefix="$t('分类名称')" apiSource="category" />

      <el-form-item
        for="no_click"
        :label="$t('分类排序')"
        prop="sort"
        :rules="[
          {
            required: true,
            message: $t('分类排序不能为空'),
          },
          {
            type: 'number',
            message: $t('分类排序必须为数字'),
          },
        ]"
      >
        <el-input-number
          :controls="false"
          :precision="0"
          :placeholder="$t('接近0，排序等级越高')"
          :min="0"
          :max="999"
          v-model.number="form.sort"
          autocomplete="off"
        ></el-input-number>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-tooltip effect="dark" placement="top" :content="$t('如长时间无响应，请刷新后重试。')">
          <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
        </el-tooltip>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
// 引入Vue3组合式API
import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
// 引入Element Plus消息
import { ElMessage } from 'element-plus';
// 引入API
import ProductApi from '@/api/product.js';
// 引入组件
import UniqueNameForm from '@/components/product/UniqueNameForm.vue';
// 引入store
import { languageStore } from '@/store/model/language.js';


// 定义props
const props = defineProps({
  open_add: {
    type: Boolean,
    default: false
  },
  addform: {
    type: Object,
    default: () => ({})
  }
});

// 定义emits
const emit = defineEmits(['closeDialog']);

// 获取语言数据
const languageData = JSON.stringify(languageStore().getLanguageKeyForm());

// 分类数据
const category = ref([]);
// 表单数据
const form = reactive({
  parent_id: 0,
  category_id: 0,
  name: JSON.parse(languageData),
  sort: null,
  is_special: 1,
});
// 是否显示弹窗
const dialogVisible = ref(false);
// 是否正在加载
const loading = ref(false);
// 表单引用
const formRef = ref();
// 唯一名称表单引用
const uniqueNameFormRef = ref();

// 提交表单
async function submit() {
  loading.value = true;
  try {
    // 验证唯一名称表单
    const validUniqueName = await uniqueNameFormRef.value.validate();
    // 验证主表单
    const validForm = await formRef.value.validate();

    if (!validUniqueName || !validForm) return;

    // 复制表单数据
    const params = JSON.parse(JSON.stringify(form));

    // 获取唯一名称数据
    const _name = uniqueNameFormRef.value.data;
    params.name = JSON.stringify(_name);
    
    // 调用API
    await ProductApi.storeCatAdd(params, true);
    
    // 显示成功消息
    ElMessage({
      message: $t('保存成功'),
      type: 'success',
    });
    
    // 关闭弹窗
    dialogFormVisible(true);
  } catch (error) {
    // 错误处理
  } finally {
    loading.value = false;
  }
}

// 关闭弹窗
function dialogFormVisible(e) {
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
}

// 组件挂载时设置弹窗状态
onMounted(() => {
  dialogVisible.value = props.open_add;
});
</script>

<style scoped>
.img {
  margin-top: 10px;
}
</style>

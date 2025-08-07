<!-- 描述：商品-单位库-添加单位 -->
<template>
  <el-dialog :title="$t('添加单位')" v-model="dialogVisible" @close="handleClose" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <UniqueNameForm ref="uniqueNameFormRef" :labelPrefix="$t('单位名称')" apiSource="unit" />
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
import { ref, reactive, watch, getCurrentInstance } from 'vue'
import { ElMessage } from 'element-plus'
import ProductApi from '@/api/product.js'
import UniqueNameForm from '@/components/product/UniqueNameForm.vue'

// 获取当前实例
const { proxy } = getCurrentInstance()

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
})

// 定义emits
const emit = defineEmits(['closeDialog'])

// 响应式数据
const form = reactive({
  unit_name: {},
})
const dialogVisible = ref(false)
const loading = ref(false)

// 模板引用
const formRef = ref(null)
const uniqueNameFormRef = ref(null)

// 监听props变化
watch(() => props.open_add, (newVal) => {
  dialogVisible.value = newVal
}, { immediate: true })

// 方法定义
const submit = async () => {
  loading.value = true
  
  try {
    // 验证表单
    const validForm = await formRef.value.validate()
    if (!validForm) return

    // 验证唯一名称
    const validUniqueName = await uniqueNameFormRef.value.validate()
    if (!validUniqueName) return

    // 获取名称数据
    const _name = uniqueNameFormRef.value.data
    const params = JSON.parse(JSON.stringify(form))
    params.unit_name = JSON.stringify(_name)

    // 提交数据
    const res = await ProductApi.addUnit(params, true)
    
    ElMessage({
      message: $t('保存成功'),
      type: 'success',
    })

    handleClose(true, res.data)
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 关闭弹窗
const handleClose = (isSuccess = false, data) => {
  emit('closeDialog', {
    type: isSuccess ? 'success' : 'error',
    openDialog: false,
    data: data,
  })
}
</script>

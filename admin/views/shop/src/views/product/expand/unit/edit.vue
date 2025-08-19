<!-- 描述：商品-单位库-编辑单位 -->
<template>
  <el-dialog :title="$t('编辑单位')" v-model="dialogVisible" @close="handleClose" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <UniqueNameForm ref="uniqueNameFormRef" :labelPrefix="$t('单位名称')" apiSource="unit" :apiId="form.unit_id" :overrideLanguages="form.unit_name" />
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
  open_edit: {
    type: Boolean,
    default: false
  },
  editform: {
    type: Object,
    default: () => ({})
  }
})

// 定义emits
const emit = defineEmits(['closeDialog'])

// 响应式数据
const form = reactive({
  unit_id: undefined,
  unit_name: {},
})
const dialogVisible = ref(false)
const loading = ref(false)

// 模板引用
const formRef = ref(null)
const uniqueNameFormRef = ref(null)

// 监听props变化并初始化数据
watch(() => props.open_edit, (newVal) => {
  dialogVisible.value = newVal
  
  if (newVal && props.editform) {
    // 初始化表单数据
    form.unit_id = props.editform.unit_id
    
    try {
      const _names = typeof props.editform.unit_name === 'string' 
        ? JSON.parse(props.editform.unit_name) 
        : props.editform.unit_name ?? {}
      form.unit_name = _names
    } catch (error) {
      console.error('parse name failed', error)
    }
  }
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
    const res = await ProductApi.editUnit(params, true)
    
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

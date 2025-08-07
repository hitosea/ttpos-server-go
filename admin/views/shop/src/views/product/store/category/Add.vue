<template>
  <el-dialog :title="$t('添加普通分类')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false" append-to-body>
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <el-form-item for="no_click" :label="$t('分类级别')" prop="parent">
        <el-radio-group v-model="parent" @change="radioChange">
          <el-radio :value="1">{{ $t('一级分类') }}</el-radio>
          <el-radio :value="0">{{ $t('二级分类') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item for="no_click" v-if="parent == 0" :label="$t('上级分类')" prop="parent_id" :rules="[{ required: true, message: $t('请选择上级分类') }]">
        <el-select v-model="form.parent_id" :placeholder="$t('请选择上级分类')">
          <template v-for="cat in category" :key="cat.category_id">
            <el-option :value="cat.category_id" :label="cat.name_text"></el-option>
          </template>
        </el-select>
      </el-form-item>

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
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import ProductApi from '@/api/product.js'
import { languageStore } from '@/store/model/language.js'
import UniqueNameForm from '@/components/product/UniqueNameForm.vue'

// 获取语言数据
const languageData = JSON.stringify(languageStore().getLanguageKeyForm())


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
const category = ref([])
const parent = ref(1)
const form = reactive({
  parent_id: 0,
  category_id: 0,
  name: JSON.parse(languageData),
  sort: null,
})
const dialogVisible = ref(false)
const loading = ref(false)

// 表单引用
const formRef = ref(null)
const uniqueNameFormRef = ref(null)

// 监听open_add变化
watch(() => props.open_add, (newVal) => {
  dialogVisible.value = newVal
}, { immediate: true })

// 获取父级分类
const getParentCategory = async () => {
  loading.value = true
  try {
    const res = await ProductApi.storeCatParentList({}, true)
    category.value = res.data.list
  } catch (err) {
    ElMessage.error(err?.message || $t('获取父级分类失败'))
  } finally {
    loading.value = false
  }
}

// 提交表单
const submit = async () => {
  loading.value = true
  try {
    const validUniqueName = await uniqueNameFormRef.value.validate()
    const validForm = await formRef.value.validate()

    if (!validUniqueName || !validForm) return
    
    const params = JSON.parse(JSON.stringify(form))
    if (parent.value === 1) {
      params.parent_id = ''
    }

    const _name = uniqueNameFormRef.value.data
    params.name = JSON.stringify(_name)
    
    const res = await ProductApi.storeCatAdd(params, true)
    ElMessage({
      message: $t('保存成功'),
      type: 'success',
    })
    dialogFormVisible(true, res.data)
  } catch (error) {
    // 错误处理
  } finally {
    loading.value = false
  }
}

// 关闭弹窗
const dialogFormVisible = (e, data) => {
  if (e) {
    emit('closeDialog', {
      type: 'success',
      data: data,
      openDialog: false,
    })
  } else {
    emit('closeDialog', {
      type: 'error',
      openDialog: false,
    })
  }
}

// 单选按钮变化
const radioChange = () => {
  form.parent_id = ''
}

// 组件挂载时获取父级分类
onMounted(() => {
  getParentCategory()
})
</script>

<style scoped>
  .img {
    margin-top: 10px;
  }
</style>

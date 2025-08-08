<template>
  <template v-for="(item, index) in form.productTaxes">
    <el-form-item
      for="no_click"
      :label="returnType(item.product_tax_type)"
      :prop="`form.productTaxes[${index}].tax_category_id`"
      :rules="[
        {
          required: true,
          validator: () => {
            return item.tax_category_id ? true : false;
          },
          message: returnMessage(item.product_tax_type),
        },
      ]"
    >
      <el-select v-model="item.tax_category_id" clearable class="max-w460" size="default" :placeholder="$t('请选择')">
        <template v-for="cat in taxList" :key="cat.id">
          <el-option :value="cat.id" :label="cat.name"></el-option>
        </template>
      </el-select>
    </el-form-item>
  </template>
</template>

<script setup>
import { ref, inject, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import PorductApi from '@/api/product.js'

// 注入form
const form = inject('form', {})

// 定义emits
const emit = defineEmits(['loading', 'close'])

// 响应式数据
const taxList = ref([])
const loading = ref(false)

// 获取基础数据
const getTaxData = async () => {
  try {
    const res = await PorductApi.getTaxList({}, true)
    taxList.value = res.data.list
    
    // 验证现有选择的税费类别是否有效
    await Promise.resolve().then(() => {
      let idArr = []
      taxList.value.map((item) => {
        idArr.push(item.id)
      })
      form.productTaxes.map((item) => {
        if (!idArr.includes(item.tax_category_id)) {
          item.tax_category_id = ''
        }
      })
    })
  } catch (error) {
    // 错误处理
  }
}

// 返回税费类型名称
const returnType = (type) => {
  let result = ''
  if (type == '1') {
    result = $t('堂食税类：')
  } else {
    result = $t('外带税类：')
  }
  return result
}

// 返回验证消息
const returnMessage = (type) => {
  let result = ''
  if (type == '1') {
    result = $t('请选择堂食税类')
  } else {
    result = $t('请选择外带税类')
  }
  return result
}

// 提交
const submit = async () => {
  loading.value = true
  emit('loading', true)
  try {
    const data = {
      productTaxes: form.productTaxes,
      product_ids: form.product_ids,
    }
    await PorductApi.batchUpdateTax(data, true)
    loading.value = false
    emit('loading', false)
    ElMessage({
      type: 'success',
      message: $t('操作成功'),
    })
    emit('close')
  } catch (error) {
    loading.value = false
    emit('loading', false)
  }
}

onMounted(() => {
  // 获取税费数据
  getTaxData()
})

// 暴露方法给父组件
defineExpose({
  submit,
  getTaxData,
  returnType,
  returnMessage
})
</script>

<style lang=""></style>

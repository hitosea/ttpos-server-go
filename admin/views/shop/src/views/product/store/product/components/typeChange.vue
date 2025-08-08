<template>
  <el-form-item
    :label="$t('所属分类')"
    for="no_click"
    prop="category_id"
    :rules="[
      {
        validator: () => {
          return form.category_id ? true : false;
        },
        required: true,
        message: $t('请选择分类'),
      },
    ]"
  >
    <el-cascader class="max-w460 mr8" :options="options" v-model="form.category_id" clearable style="width: 100%" :placeholder="$t('请选择分类')"></el-cascader>
    <el-button type="primary" size="small" :loading="loading" @click="add">{{ $t('添加分类') }} </el-button>
  </el-form-item>

  <!--添加-->
  <Add v-if="open_add" :open_add="open_add" @closeDialog="closeDialogFunc($event, 'add')"> </Add>
</template>

<script setup>
import { ref, inject, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import PorductApi from '@/api/product.js'
import Add from '../../category/Add.vue'

// 注入form
const form = inject('form', {})

// 定义emits
const emit = defineEmits(['loading', 'close'])

// 响应式数据
const loading = ref(false)
const open_add = ref(false)
const options = ref([])

// 获取列表
const getData = async () => {
  loading.value = true
  try {
    const data = await PorductApi.storeCatList(
      {
        page: 1,
        list_rows: 1000,
      },
      true
    )
    loading.value = false
    options.value = []
    
    // 第一遍：推入一级分类
    await Promise.resolve().then(() => {
      data.data.list.data.map((item) => {
        if (item.category_id != '0') {
          options.value.push({
            value: item.category_id,
            label: item.name_text,
            children: [],
          })
        }
      })
    })
    
    // 第二遍：为每个有子级的分类添加子节点
    await Promise.resolve().then(() => {
      data.data.list.data.map((item, index) => {
        if (item.child && item.child.length > 0) {
          item.child.map((items) => {
            if (options.value[index]) {
              options.value[index].children.push({
                value: items.category_id,
                label: items.name_text,
              })
            }
          })
        }
      })
    })
  } catch (error) {
    loading.value = false
  }
}

// 提交
const submit = async () => {
  loading.value = true
  emit('loading', true)
  try {
    const data = {
      category_id: form.category_id[form.category_id.length - 1],
      product_ids: form.product_ids,
    }
    await PorductApi.batchUpdateCategory(data, true)
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

// 打开新增分类
const add = () => {
  open_add.value = true
}

// 关闭弹窗
const closeDialogFunc = async (e, f) => {
  if (f == 'add') {
    open_add.value = e.openDialog
    if (e.type == 'success') {
      await getData()
    }
  }
}

onMounted(() => {
  // 获取列表
  getData()
})

// 暴露方法给父组件
defineExpose({
  submit,
  getData,
  add,
  closeDialogFunc
})
</script>

<style lang="scss" scoped></style>

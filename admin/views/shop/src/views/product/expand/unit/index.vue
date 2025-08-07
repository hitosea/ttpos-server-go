<template>
  <!--
      描述：单位库
  -->
  <div class="product-list">
    <!--添加单位-->
    <div class="common-level-rail">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item>
          <el-input size="small" v-model="searchForm.name" :placeholder="$t('单位名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <div>
        <el-button size="small" type="primary" icon="Plus" v-auth="'/product/expand/unit/add'" @click="addClick">{{ $t('添加单位') }}</el-button>
        <el-button size="small" v-auth="'/product/expand/unit/batch_delete'" :disabled="multipleSelection.length == 0" @click="deleteBatch">{{ $t('批量删除') }}</el-button>
      </div>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading" @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="45"></el-table-column>
          <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center" :index="indexMethod"></el-table-column>
          <el-table-column prop="unit_name_text" :label="$t('单位名称')"></el-table-column>
          <!-- <el-table-column prop="sort" :label="$t('排序')"></el-table-column> -->
          <el-table-column prop="product_ids" :label="$t('关联商品数量')" width="120">
            <template #default="scope">
              {{ scope.row.product_ids?.length ?? 0 }}
            </template>
          </el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="240">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/product/expand/unit/edit'">{{ $t('编辑') }} </el-button>
              <el-button @click="relatedProductClick(scope.row)" type="primary" v-auth="'/product/expand/unit/relatedProduct'" link size="small"
                >{{ $t('关联商品/材料') }}
              </el-button>
              <el-button
                @click="deleteClick(scope.row.unit_id)"
                type="primary"
                link
                size="small"
                v-auth="'/product/expand/unit/delete'"
                :disabled="scope.row.product_ids?.length > 0"
                >{{ $t('删除') }}</el-button
              >
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
    <!--分页-->
    <div class="pagination">
      <el-pagination
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        background
        :current-page="curPage"
        :page-size="pageSize"
        layout="total, prev, pager, next, jumper"
        :total="totalDataNumber"
      ></el-pagination>
    </div>
    <!--添加-->
    <Add v-if="open_add" :open_add="open_add" :addform="model" @closeDialog="closeDialogFunc($event, 'add')"></Add>
    <!--修改-->
    <Edit v-if="open_edit" :open_edit="open_edit" :editform="model" @closeDialog="closeDialogFunc($event, 'edit')"> </Edit>

    <!-- 商品选择器 -->
    <ProductSelector
      v-if="openProductSelector"
      :open="openProductSelector"
      @close="handleProductSelectorClose"
      selectorType="all"
      type="all"
      :selectedProductIds="model?.product_ids ?? []"
      :isLoading="loading"  
    >
    </ProductSelector>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, getCurrentInstance } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProductApi from '@/api/product.js'
import Add from './add.vue'
import Edit from './edit.vue'
import ProductSelector from '@/components/product/Selector.vue'

// 获取当前实例
const { proxy } = getCurrentInstance()

// 响应式数据
const loading = ref(true)
const pageSize = ref(10)
const totalDataNumber = ref(0)
const curPage = ref(1)
const model = ref({})
const open_edit = ref(false)
const open_add = ref(false)
const tableData = ref([])
const multipleSelection = ref([])
const searchForm = reactive({
  name: '',
})
const searchLoading = ref('')
const openProductSelector = ref(false)

// 方法定义
const handleCurrentChange = (val) => {
  loading.value = true
  curPage.value = val
  getData()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  getData()
}

const handleClick = (tab, event) => {
  curPage.value = 1
  getData()
}

const onSearch = () => {
  clearTimeout(searchLoading.value)
  searchLoading.value = setTimeout(() => {
    curPage.value = 1
    getData()
  }, 200)
}

const getData = async () => {
  const params = {
    page: curPage.value,
    list_rows: pageSize.value,
    unit_name: searchForm.name,
  }
  loading.value = true
  
  try {
    const data = await ProductApi.UnitList(params, true)
    loading.value = false
    tableData.value = data.data.list.data
    totalDataNumber.value = data.data.list.total
  } catch (error) {
    loading.value = false
  }
}

const closeDialogFunc = (e, f) => {
  if (f === 'add') {
    open_add.value = e.openDialog
    if (e.type === 'success') {
      getData()
    }
  }
  if (f === 'edit') {
    open_edit.value = e.openDialog
    if (e.type === 'success') {
      getData()
    }
  }
}

const addClick = () => {
  open_add.value = true
}

const deleteClick = async (id) => {
  try {
    await ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      type: 'warning',
    })
    
    await ProductApi.deleteUnit({
      unit_id: id,
    })
    
    ElMessage({
      message: $t('删除成功'),
      type: 'success',
    })
    getData()
  } catch (error) {
    // 用户取消删除或删除失败
  }
}

const deleteBatch = async () => {
  const arr = []
  multipleSelection.value.forEach((item) => {
    arr.push(item.unit_id)
  })
  const unit_id = arr.join(',')
  
  try {
    await ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      type: 'warning',
    })
    
    await ProductApi.deleteUnit({
      unit_id: unit_id,
    })
    
    ElMessage({
      message: $t('删除成功'),
      type: 'success',
    })
    getData()
  } catch (error) {
    // 用户取消删除或删除失败
  }
}

const handleSelectionChange = (e) => {
  multipleSelection.value = e
}

const editClick = (row) => {
  model.value = row
  open_edit.value = true
}

const relatedProductClick = (row) => {
  model.value = row
  openProductSelector.value = true
}

const handleProductSelectorClose = async (list) => {
  if (Array.isArray(list)) {
    try {
      loading.value = true;
      await ProductApi.relateByUnit(
        {
          unit_id: model.value.unit_id,
          product_ids: list.map((item) => item.product_id),
        },
        false
      )
      
      ElMessage({
        message: $t('关联成功'),
        type: 'success',
      })
      getData()
      loading.value = false;
    } catch (error) {
      // 处理错误
      loading.value = false;
    }
  }
  model.value = {}
  openProductSelector.value = false
}

const indexMethod = (index) => {
  return index + 1 + (curPage.value - 1) * pageSize.value
}

// 组件挂载时获取数据
onMounted(() => {
  getData()
})
</script>

<style scoped>
  .common-level-rail {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }
</style>

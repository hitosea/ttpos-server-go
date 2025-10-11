<template>
  <div class="product">
    <!--搜索表单-->
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('分类名称')">
          <el-input size="small" v-model="searchForm.name" :placeholder="$t('请输入分类名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <el-button size="small" type="primary" @click="addClick" icon="Plus" v-auth="'/product/store/category/add'">{{ $t('添加分类') }}</el-button>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" row-key="category_id" default-expand-all :tree-props="{ children: 'child' }" style="width: 100%" v-loading="loading">
          <el-table-column prop="name_text" :label="$t('分类名称')"></el-table-column>
          <el-table-column prop="name_text" :label="$t('分类级别')">
            <template #default="scope">
              <p v-if="scope.row.parent_id == 0 && scope.row.is_button != 1">{{ $t('一级分类') }}</p>
              <p v-else-if="scope.row.is_button == 1">-</p>
              <p v-else>{{ $t('二级分类') }}</p>
            </template>
          </el-table-column>
          <el-table-column prop="product_num" :label="$t('关联商品数量')" width="120">
            <template #default="scope">
              <p v-if="scope.row.is_button == 1">-</p>
              <p v-else>{{ scope.row.product_num }}</p>
            </template>
          </el-table-column>
          <el-table-column prop="sort" :label="$t('状态')">
            <template #default="scope">
              <el-switch
                :disabled="!proxy.$filter.isAuth('/product/store/category/state')"
                v-model="scope.row.status"
                :active-value="1"
                :inactive-value="0"
                @change="statusSet($event, scope.row.category_id)"
              >
              </el-switch>
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')">
            <template #default="scope">
              <p>{{ scope.row.create_time }}</p>
            </template>
          </el-table-column>
          <el-table-column prop="sort" :label="$t('分类排序')" width="120"></el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="120">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/product/store/category/edit'">{{ $t('编辑') }}</el-button>
              <el-button
                @click="deleteClick(scope.row)"
                type="primary"
                :disabled="scope.row.product_num > 0 || scope.row.is_button == 1 || (scope.row.child || []).length > 0"
                link
                size="small"
                v-auth="'/product/store/category/delete'"
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
    <Add v-if="open_add" :open_add="open_add" :addform="categoryModel" @closeDialog="closeDialogFunc($event, 'add')"> </Add>
    <!--修改-->
    <Edit v-if="open_edit" :open_edit="open_edit" :editform="categoryModel" @closeDialog="closeDialogFunc($event, 'edit')"></Edit>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, getCurrentInstance } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import PorductApi from '@/api/product.js'
import Add from './Add.vue'
import Edit from './Edit.vue'


const { proxy } = getCurrentInstance();

// 响应式数据
const loading = ref(true)
const pageSize = ref(10)
const totalDataNumber = ref(0)
const curPage = ref(1)
const tableData = ref([])
const open_add = ref(false)
const open_edit = ref(false)
const categoryModel = reactive({
  catList: [],
  model: {},
})
const searchForm = reactive({
  name: '',
})
const searchLoading = ref(null)

// 方法定义
// 选择第几页
const handleCurrentChange = (val) => {
  loading.value = true
  curPage.value = val
  getData()
}

// 每页多少条
const handleSizeChange = (val) => {
  pageSize.value = val
  getData()
}

// 搜索查询
const onSearch = () => {
  clearTimeout(searchLoading.value)
  searchLoading.value = setTimeout(() => {
    curPage.value = 1
    getData()
  }, 200)
}

// 切换菜单
const handleClick = () => {
  curPage.value = 1
  getData()
}

// 获取列表
const getData = async () => {
  loading.value = true
  try {
    const res = await PorductApi.storeCatList(
    {
      name: searchForm.name,
      page: curPage.value,
      list_rows: pageSize.value,
    },
    true
  )
  loading.value = false
  tableData.value = res.data.list.data || res.data.data || []
  categoryModel.catList = tableData.value
  totalDataNumber.value = res.data.list.total || 0
  } catch (error) {
    loading.value = false
  }
    
}

// 打开添加
const addClick = () => {
  open_add.value = true
}

// 打开编辑
const editClick = (item) => {
  categoryModel.model = item
  open_edit.value = true
}

// 状态设置
const statusSet = async (e, id) => {
  try {
    const res = await PorductApi.storeCatSet({
      category_id: id,
      status: e,
    },true)
    ElMessage({
      message: res.msg,
      type: 'success',
    })
  } catch (error) {
    ElMessage.error(error?.message || $t('状态设置失败'))
  }
}

// 关闭弹窗
const closeDialogFunc = (e, f) => {
  if (f == 'add') {
    open_add.value = e.openDialog
    if (e.type == 'success') {
      getData()
    }
  }
  if (f == 'edit') {
    open_edit.value = e.openDialog
    if (e.type == 'success') {
      getData()
    }
  }
}

// 删除分类
const deleteClick = (row) => {
  ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
    type: 'warning',
  }).then(async () => {
    try {
      const res = await PorductApi.storeCatDel({
        category_id: row.category_id,
      },true)
      ElMessage({
        message: $t('删除成功'),
        type: 'success',
      })
      getData()
    } catch (err) {
      ElMessage.error(err?.message || $t('删除失败'))
    }
  })
}

// 组件挂载时获取数据
onMounted(() => {
  getData()
})
</script>

<style scoped>
  .common-search-wrap {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }
</style>

<template>
  <div class="product">
    <div class="common-level-rail">
      <el-button
        size="small"
        type="primary"
        @click="addClick"
        icon="Plus"
        v-auth="'/product/store/feature/add'"
      >{{ $t('添加分类') }}</el-button>
    </div>
    <div class="product-content">
      <div class="table-wrap">
        <el-table
          size="small"
          :data="tableData"
          row-key="category_id"
          default-expand-all
          :tree-props="{ children: 'child' }"
          style="width: 100%"
          v-loading="loading"
        >
          <el-table-column prop="name_text" :label="$t('分类名称')"></el-table-column>
          <el-table-column prop="product_num" :label="$t('关联商品数量')" width="120"></el-table-column>
          <el-table-column prop="sort" :label="$t('状态')">
            <template #default="scope">
              <el-switch
                :disabled="!proxy.$filter.isAuth('/product/store/feature/state')"
                v-model="scope.row.status"
                :active-value="1"
                :inactive-value="0"
                @change="val => statusSet(val, scope.row.category_id)"
              />
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')"></el-table-column>
          <el-table-column prop="sort" :label="$t('排序')"></el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="120">
            <template #default="scope">
              <el-button
                @click="() => editClick(scope.row)"
                type="primary"
                link
                size="small"
                v-auth="'/product/store/feature/edit'"
              >{{ $t('编辑') }}</el-button>
              <el-button
                @click="() => deleteClick(scope.row)"
                type="primary"
                :disabled="scope.row.product_num > 0"
                link
                size="small"
                v-auth="'/product/store/feature/delete'"
              >{{ $t('删除') }}</el-button>
            </template>
          </el-table-column>
        </el-table>
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
        />
      </div>
      <!--添加-->
      <Add v-if="openAdd" :open_add="openAdd" :addform="categoryModel" @closeDialog="e => closeDialogFunc(e, 'add')" />
      <!--修改-->
      <Edit v-if="openEdit" :open_edit="openEdit" :editform="categoryModel" @closeDialog="e => closeDialogFunc(e, 'edit')" />
    </div>
  </div>
</template>

<script setup>
// 引入Vue3组合式API
import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
// 引入Element Plus弹窗和消息
import { ElMessageBox, ElMessage } from 'element-plus';
// 引入API
import PorductApi from '@/api/product.js';
// 引入子组件
import Edit from './Edit.vue';
import Add from './Add.vue';

const { proxy } = getCurrentInstance();

// 是否正在加载
const loading = ref(false);
// 每页多少条
const pageSize = ref(10);
// 总数据条数
const totalDataNumber = ref(0);
// 当前页码
const curPage = ref(1);
// 添加弹窗开关
const openAdd = ref(false);
// 编辑弹窗开关
const openEdit = ref(false);
// 分类模型
const categoryModel = reactive({
  catList: [],
  model: {},
});
// 表格数据
const tableData = ref([]);

// 获取数据
async function getData() {
  loading.value = true;
  try {
    const res = await PorductApi.storeCatSP(
      {
        page: curPage.value,
        list_rows: pageSize.value,
      },
      true
    );
    tableData.value = res.data.list.data || res.data.data || [];
    categoryModel.catList = tableData.value;
    totalDataNumber.value = res.data.list.total || 0;
  } catch (e) {
    // 错误处理
  } finally {
    loading.value = false;
  }
}

// 分页-切换页码
function handleCurrentChange(val) {
  loading.value = true;
  curPage.value = val;
  getData();
}
// 分页-切换每页条数
function handleSizeChange(val) {
  pageSize.value = val;
  getData();
}
// 打开添加弹窗
function addClick() {
  openAdd.value = true;
}
// 打开编辑弹窗
function editClick(item) {
  categoryModel.model = item;
  openEdit.value = true;
}
// 关闭弹窗
function closeDialogFunc(e, f) {
  if (f === 'add') {
    openAdd.value = e.openDialog;
    if (e.type === 'success') {
      getData();
    }
  }
  if (f === 'edit') {
    openEdit.value = e.openDialog;
    if (e.type === 'success') {
      getData();
    }
  }
}
// 删除分类
async function deleteClick(row) {
  try {
    await ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      type: 'warning',
    });
    await PorductApi.storeCatDel({
      category_id: row.category_id,
    });
    ElMessage({
      message: $t('删除成功'),
      type: 'success',
    });
    getData();
  } catch (e) {
    // 用户取消或出错
  }
}
// 状态设置
async function statusSet(e, id) {
  try {
    const data = await PorductApi.storeCatSet({
      category_id: id,
      status: e,
    });
    ElMessage({
      message: data.msg,
      type: 'success',
    });
  } catch (err) {
    // 错误处理
  }
}
// 页面挂载时获取数据
onMounted(() => {
  getData();
});
</script>

<style scoped>
.common-level-rail {
  text-align: right;
}
</style>

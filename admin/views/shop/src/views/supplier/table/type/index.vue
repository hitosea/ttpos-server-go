<template>
  <div class="product">
    <div class="common-level-rail">
      <el-button size="small" type="primary" @click="addClick" icon="Plus" v-auth="'/supplier/table/type/add'">{{ $t('添加类型') }}</el-button>
    </div>
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" row-key="type_id" style="width: 100%" v-loading="loading">
          <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center"></el-table-column>
          <el-table-column prop="type_name" :label="$t('类型名称')"></el-table-column>
          <el-table-column prop="max_num" :label="$t('人数区间')">
            <template #default="scope"> {{ scope.row.min_num }}-{{ scope.row.max_num }}{{ $t('人') }} </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')"></el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="100">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/supplier/table/type/edit'">{{ $t('编辑') }}</el-button>
              <el-button @click="deleteClick(scope.row)" :disabled="scope.row.can_delete == 0" type="primary" link size="small" v-auth="'/supplier/table/type/delete'">{{
                $t('删除')
              }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
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
    </div>
    <!--添加-->
    <Add v-if="open_add" :open_add="open_add" :addform="categoryModel" @closeDialog="closeDialogFunc($event, 'add')"> </Add>
    <!--修改-->
    <Edit v-if="open_edit" :open_edit="open_edit" :editform="categoryModel" @closeDialog="closeDialogFunc($event, 'edit')"></Edit>
  </div>
</template>

<script>
  import StoreApi from '@/api/store.js';
  import Add from './add.vue';
  import Edit from './edit.vue';
  export default {
    name: 'SupplierTableTypeIndex',
    components: {
      Add,
      Edit,
    },
    data() {
      return {
        /*是否加载完成*/
        loading: true,
        /*列表数据*/
        tableData: [],
        /*是否打开添加弹窗*/
        open_add: false,
        /*是否打开编辑弹窗*/
        open_edit: false,
        /*当前编辑的对象*/
        categoryModel: {
          model: '',
        },
        curPage: 1,
        pageSize: 10,
        totalDataNumber: 0,
      };
    },
    created() {
      /*获取列表*/
      this.getData();
    },
    methods: {
      handleSizeChange(size) {
        this.pageSize = size;
        this.getData();
      },
      handleCurrentChange(page) {
        this.curPage = page;
        this.getData();
      },
      handleClick() {
        this.getData();
      },
      /*获取列表*/
      getData() {
        let self = this;
        self.loading = true;
        let params = {};
        params.page = self.curPage;
        params.list_rows = self.pageSize;
        StoreApi.typelist(params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.categoryModel = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
          })
          .catch(() => {
            self.loading = false;
          });
      },
      /*打开添加*/
      addClick() {
        this.open_add = true;
      },

      /*打开编辑*/
      editClick(item) {
        this.categoryModel.model = item;
        this.open_edit = true;
      },

      /*关闭弹窗*/
      closeDialogFunc(e, f) {
        if (f == 'add') {
          this.open_add = e.openDialog;
          if (e.type == 'success') {
            this.getData();
          }
        }
        if (f == 'edit') {
          this.open_edit = e.openDialog;
          if (e.type == 'success') {
            this.getData();
          }
        }
      },
      /*删除分类*/
      deleteClick(row) {
        const self = this;
        ElMessageBox.confirm(self.$t('删除后不可恢复，确认删除吗？'), self.$t('提示'), {
          type: 'warning',
        }).then(() => {
          StoreApi.deleteType({
            type_id: row.type_id,
          }).then(() => {
            self.$ElMessage({
              message: self.$t('删除成功'),
              type: 'success',
            });
            self.getData();
          });
        });
      },
    },
  };
</script>

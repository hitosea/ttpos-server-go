<template>
  <div class="product">
    <div class="common-level-rail flex">
      <el-form size="small" :inline="true" :model="form" class="demo-form-inline d-s-c">
        <el-form-item :label="$t('桌位名称')">
          <el-input v-model="form.search" autocomplete="off" :placeholder="$t('桌位名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item :label="$t('选择区域')">
          <a-select size="small" v-model:value="form.area_id" :placeholder="$t('请选择')" @change="onSearch">
            <el-option :label="$t('全部')" value=""></el-option>
            <el-option v-for="(item, index) in area_list" :key="index" :label="item.area_name" :value="item.area_id"> </el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('选择类型')">
          <a-select size="small" v-model:value="form.type_id" :placeholder="$t('请选择')" @change="onSearch">
            <el-option :label="$t('全部')" value=""></el-option>
            <el-option v-for="(item, index) in type_list" :key="index" :label="item.type_name" :value="item.type_id"> </el-option>
          </a-select>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <div class="product-operation">
        <el-dropdown>
          <el-button size="small" type="primary" icon="ArrowDown">
            {{ $t('批量') }}
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="importQrcode" v-if="this.$filter.isAuth('/supplier/table/table/importQrcode')">
                {{ $t('导入桌台') }}
              </el-dropdown-item>
              <el-dropdown-item @click="batchDownloadQrcode('download')" v-if="this.$filter.isAuth('/supplier/table/table/batchQrcode')">
                {{ $t('下载二维码') }}
              </el-dropdown-item>
              <el-dropdown-item @click="batchDownloadQrcode('delete')" v-if="this.$filter.isAuth('/supplier/table/table/delete')">
                {{ $t('删除') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button size="small" type="primary" @click="addClick" icon="Plus" v-auth="'/supplier/table/table/add'">
          {{ $t('添加桌位') }}
        </el-button>
      </div>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" style="width: 100%" v-loading="loading">
          <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center" :index="indexMethod"></el-table-column>
          <el-table-column prop="area_name" :label="$t('所属区域')"></el-table-column>
          <el-table-column prop="type_name" :label="$t('所属类型')"></el-table-column>
          <el-table-column prop="table_no" :label="$t('人数区间')">
            <template #default="scope"> {{ scope.row.min_num }}-{{ scope.row.max_num }}{{ $t('人') }}</template>
          </el-table-column>
          <el-table-column prop="table_no" :label="$t('桌位名称')"></el-table-column>
          <el-table-column prop="sort" :label="$t('排序')"></el-table-column>
          <el-table-column prop="" :label="$t('状态')">
            <template #default="scope">
              <el-switch
                :disabled="!this.$filter.isAuth('/supplier/pay/state')"
                :model-value="scope.row.switch_status"
                :active-value="1"
                :inactive-value="0"
                @click="changeStatus(scope.row)"
              ></el-switch>
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')"></el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="220">
            <template #default="scope">
              <el-button v-if="is_open_scan" @click="qrcode(scope.row)" type="primary" link size="small" v-auth="'/supplier/table/table/qrcode'"> {{ $t('二维码') }}</el-button>
              <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/supplier/table/table/edit'">{{ $t('编辑') }} </el-button>
              <el-button
                v-if="scope.row.is_bind == 1"
                :disabled="scope.row.status == 30"
                @click="untieClick(scope.row)"
                type="primary"
                link
                size="small"
                v-auth="'/supplier/table/table/untie'"
                >{{ $t('解绑') }}
              </el-button>
              <el-button @click="deleteClick(scope.row)" type="primary" :disabled="scope.row.status == 30" link size="small" v-auth="'/supplier/table/table/delete'">
                {{ $t('删除') }}</el-button
              >
            </template>
          </el-table-column>
        </el-table>
      </div>
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
    <!--添加-->
    <Add v-if="open_add" :open_add="open_add" :type="type_list" :area_list="area_list" :addform="categoryModel" @closeDialog="closeDialogFunc($event, 'add')"> </Add>
    <!--修改-->
    <Edit v-if="open_edit" :open_edit="open_edit" :editform="categoryModel" :type="type_list" :area_list="area_list" @closeDialog="closeDialogFunc($event, 'edit')"></Edit>

    <Qrcode :open="isqrcode" :code_id="code_id" :code_name="code_name" @close="closeQrcode"></Qrcode>
    <DownloadQrcode v-if="is_open_batch_download_qrcode" :open="is_open_batch_download_qrcode" :DType="DType" @close="closeDownloadQrcode"> </DownloadQrcode>
  </div>
</template>

<script>
  import StoreApi from '@/api/store.js';
  import Add from './add.vue';
  import Edit from './edit.vue';
  import Qrcode from './dialog/Qrcode.vue';
  import DownloadQrcode from './batch/DownloadQrcode.vue';
  import { useUserStore } from '@/store';
  import Aselect from '@/components/a-select/index.vue';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_scan = supplier.value?.is_open_scan || 0;
  const app_id = supplier.value?.app_id || 0;
  export default {
    name: 'SupplierTableTableIndex',
    components: {
      Add,
      Edit,
      Qrcode,
      DownloadQrcode,
      Aselect,
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
        form: {
          search: '',
          area_id: '',
          type_id: '',
        },
        type_list: [],
        area_list: [],
        source: 'wx',
        /*一页多少条*/
        pageSize: 10,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*当前是第几页*/
        curPage: 1,
        isqrcode: false,
        code_id: '',
        code_name: '',
        searchLoading: '',
        is_open_scan: is_open_scan,

        is_open_batch_download_qrcode: false,
        DType: '',
        app_id: app_id,
      };
    },
    created() {
      /*获取列表*/
      this.getData();
    },
    methods: {
      /*选择第几页*/
      handleCurrentChange(val) {
        let self = this;
        self.curPage = val;
        self.getData();
      },

      /*每页多少条*/
      handleSizeChange(val) {
        this.curPage = 1;
        this.pageSize = val;
        this.getData();
      },

      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
          this.getData();
        }, 200);
      },

      /*获取列表*/
      getData() {
        let self = this;
        self.loading = true;
        let params = self.form;
        params.page = self.curPage;
        params.list_rows = self.pageSize;
        StoreApi.tablelist(params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
            self.type_list = data.data.type_list;
            self.area_list = data.data.area_list;
            self.categoryModel = data.data.list.data;
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

      batchDownloadQrcode(e) {
        this.is_open_batch_download_qrcode = true;
        this.DType = e;
      },
      closeDownloadQrcode(e) {
        this.is_open_batch_download_qrcode = false;
        this.DType = '';
        if (e) {
          this.getData();
        }
      },
      //导入
      importQrcode() {
        this.$router.push({ path: '/' + this.app_id + '/supplier/table/table/importQrcode' });
      },

      changeStatus(row) {
        let text = '';
        let self = this;
        let params = { table_id: row.table_id, switch_status: row.switch_status == 1 ? 0 : 1 };
        text = row.switch_status == 1 ? self.$t('禁用') : self.$t('启用');
        ElMessageBox.confirm(self.$t('确定') + text + self.$t('这个桌位?'), self.$t('提示'), {
          confirmButtonText: self.$t('确定'),
          cancelButtonText: self.$t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            StoreApi.switchStatus(params, true)
              .then(() => {
                self.loading = false;
                self.getData();
              })
              .catch(() => {
                self.loading = false;
              });
          })
          .catch(() => {});
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
      /*解除绑定*/
      untieClick(row) {
        let self = this;
        ElMessageBox.confirm(self.$t('确定解除与平板设备的绑定关系吗？'), self.$t('提示'), {
          type: 'warning',
        }).then(() => {
          StoreApi.unbindTable({
            table_id: row.table_id,
          }).then(() => {
            this.$ElMessage({
              message: self.$t('操作成功'),
              type: 'success',
            });
            self.getData();
          });
        });
      },
      /*删除*/
      deleteClick(row) {
        let self = this;
        ElMessageBox.confirm(self.$t('删除后不可恢复，确认删除吗？'), self.$t('提示'), {
          type: 'warning',
        }).then(() => {
          StoreApi.deleteTable({
            table_id: row.table_id,
          }).then(() => {
            this.$ElMessage({
              message: self.$t('删除成功'),
              type: 'success',
            });
            self.getData();
          });
        });
      },
      qrcode(row) {
        let self = this;
        self.code_id = row.table_id;
        self.code_name = row.table_no;
        self.isqrcode = true;
      },
      closeQrcode() {
        let self = this;
        self.isqrcode = false;
        self.code_id = '';
        self.code_name = '';
      },
      indexMethod(index) {
        return index + 1 + (this.curPage - 1) * this.pageSize;
      },
    },
  };
</script>

<style scoped lang="scss">
  .product-operation {
    display: flex;
    gap: 12px;
  }
</style>

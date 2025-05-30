<template>
  <div class="user">
    <!--搜索表单-->
    <div class="common-seach-wrap flex">
      <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item :label="$t('优惠券名称')">
          <el-input v-model="formInline.keyword" :placeholder="$t('优惠券名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item :label="$t('所有记录')">
          <a-select v-model:value="formInline.grade_id" :placeholder="$t('所有记录')" @change="onSearch">
            <el-option :label="$t('全部')" value="0"></el-option>
            <el-option v-for="(item, index) in gradeList" :key="index" :label="item.name" :value="item.grade_id"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('优惠券类型')">
          <a-select v-model:value="formInline.grade_id" :placeholder="$t('优惠券类型')" @change="onSearch">
            <el-option :label="$t('全部')" value="0"></el-option>
            <el-option v-for="(item, index) in gradeList" :key="index" :label="item.name" :value="item.grade_id"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('时间')">
          <el-date-picker
            v-model="formInline.reg_date"
            type="daterange"
            value-format="YYYY-MM-DD"
            range-separator="~"
            :start-placeholder="$t('开始日期')"
            :end-placeholder="$t('结束日期')"
            clearable
            @change="onSearch"
          ></el-date-picker>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="name" :label="$t('优惠券名称')" width="200"></el-table-column>
          <el-table-column prop="weight" :label="$t('编号')" width="200"></el-table-column>
          <el-table-column prop="equity" :label="$t('记录类型')" width="200">
            <template #default="scope">
              <span class="red fb">{{ scope.row.equity }}%</span>
            </template>
          </el-table-column>
          <el-table-column prop="remark" :label="$t('时间')" show-overflow-tooltip>
            <template #default="scope">
              <span v-html="keepTextStyle(scope.row.remark)"></span>
            </template>
          </el-table-column>
          <el-table-column prop="remark" :label="$t('数量（张）')" show-overflow-tooltip>
            <template #default="scope">
              <span v-html="keepTextStyle(scope.row.remark)"></span>
            </template>
          </el-table-column>
          <el-table-column prop="remark" :label="$t('剩余有效张数')" show-overflow-tooltip>
            <template #default="scope">
              <span v-html="keepTextStyle(scope.row.remark)"></span>
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
        >
        </el-pagination>
      </div>
    </div>

    <!--添加-->
    <Add v-if="open_add" :open_add="open_add" @closeDialog="closeDialogFunc($event, 'add')"></Add>

    <!--编辑-->
    <Edit v-if="open_edit" :open_edit="open_edit" :form="userModel" @closeDialog="closeDialogFunc($event, 'edit')"> </Edit>
  </div>
</template>

<script>
  import UserApi from '@/api/user.js';
  import Edit from './Edit.vue';
  import Add from './Add.vue';
  import { deepClone } from '@/utils/base.js';
  export default {
    components: {
      /*编辑组件*/
      Edit,
      Add,
    },
    data() {
      return {
        /*是否加载完成*/
        loading: true,
        /*列表数据*/
        tableData: [],
        /*一页多少条*/
        pageSize: 10,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*当前是第几页*/
        curPage: 1,
        /*横向表单数据模型*/
        formInline: {
          user: '',
          region: '',
        },
        /*是否打开添加弹窗*/
        open_add: false,
        /*是否打开编辑弹窗*/
        open_edit: false,
        /*当前编辑的对象*/
        userModel: {},
      };
    },
    created() {
      /*获取列表*/
      this.getTableList();
    },
    methods: {
      /*换行*/
      keepTextStyle(val) {
        let str = val.replace(/(\\r\\n)/g, '</br>');
        return str;
      },

      /*选择第几页*/
      handleCurrentChange(val) {
        let self = this;
        self.curPage = val;
        self.loading = true;
        self.getTableList();
      },

      /*每页多少条*/
      handleSizeChange(val) {
        this.curPage = 1;
        this.pageSize = val;
        this.getTableList();
      },

      /*获取列表*/
      getTableList() {
        let self = this;
        let Params = {};
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        UserApi.gradelist(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
          })
          .catch((error) => {});
      },

      /*打开添加*/
      addClick() {
        this.open_add = true;
      },

      /*打开编辑*/
      editClick(item) {
        this.userModel = deepClone(item);
        this.open_edit = true;
      },

      /*关闭弹窗*/
      closeDialogFunc(e, f) {
        if (f == 'add') {
          this.open_add = e.openDialog;
          if (e.type == 'success') {
            this.getTableList();
          }
        }
        if (f == 'edit') {
          this.open_edit = e.openDialog;
          if (e.type == 'success') {
            this.getTableList();
          }
        }
      },

      /*删除用户*/
      deleteClick(row) {
        let self = this;
        ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            UserApi.deletegrade(
              {
                grade_id: row.grade_id,
              },
              true
            )
              .then((data) => {
                self.loading = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: data.msg,
                    type: 'success',
                  });
                  self.getTableList();
                } else {
                  ElMessage.error($t('删除失败'));
                }
              })
              .catch((error) => {
                self.loading = false;
              });
          })
          .catch(() => {});
      },
    },
  };
</script>

<template>
  <div class="user">
    <!--搜索表单-->
    <div class="common-seach-wrap">
      <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item :label="$t('删除时间')">
          <div class="block">
            <el-date-picker
              v-model="formInline.date"
              type="daterange"
              value-format="YYYY-MM-DD"
              range-separator="~"
              :start-placeholder="$t('开始日期')"
              :end-placeholder="$t('结束日期')"
              clearable
              @change="onSearch"
            >
            </el-date-picker>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="onSubmit">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="card_name" :label="$t('会员卡名称')"> </el-table-column>
          <el-table-column prop="money" :label="$t('价格')">
            <template #default="scope">
              {{ this.$formatPrice(scope.row.money) }}
            </template>
          </el-table-column>
          <el-table-column prop="receive_num" :label="$t('领取人数')"></el-table-column>
          <el-table-column prop="create_time" :label="$t('创建时间')"></el-table-column>
          <el-table-column prop="update_time" :label="$t('删除时间')"></el-table-column>
          <!-- <el-table-column fixed="right" label="操作" width="120">
            <template slot-scope="scope">
            </template>
          </el-table-column> -->
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
  </div>
</template>

<script>
  import CardApi from '@/api/card.js';
  import dayjs from '@/utils/dayjs';
  export default {
    components: {},
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
          date: '',
        },
      };
    },
    mounted() {
      // js获取当天时间 日期格式 YYYY-MM-DD
      this.formInline.date = [dayjs(), dayjs()];
      /*获取列表*/
      this.getTableList();
    },
    methods: {
      /*换行*/
      keepTextStyle(val) {
        let str = val.replace(/(\\r\\n)/g, '<br/>');
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
        let Params = self.formInline;
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        CardApi.deleterecordlist(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
          })
          .catch((error) => {});
      },
      /*搜索查询*/
      onSubmit() {
        let self = this;
        self.loading = true;
        self.curPage = 1;
        self.getTableList();
      },
    },
  };
</script>

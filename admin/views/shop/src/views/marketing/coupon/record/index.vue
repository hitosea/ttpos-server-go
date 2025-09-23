<template>
  <div class="user">
    <!--搜索表单-->
    <div class="common-search-wrap flex">
      <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item :label="$t('优惠券名称')">
          <el-input v-model="formInline.coupon_name" :placeholder="$t('优惠券名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item :label="$t('所有记录')">
          <a-select v-model:value="formInline.record_type" :placeholder="$t('所有记录')" @change="onSearch">
            <el-option :label="$t('全部')" value=""></el-option>
            <el-option :label="$t('首次添加')" value="1"></el-option>
            <el-option :label="$t('调整添加')" value="2"></el-option>
            <el-option :label="$t('调整扣减')" value="3"></el-option>
            <el-option :label="$t('反结账退还')" value="4"></el-option>
            <el-option :label="$t('奖励领取（冻结）')" value="5"></el-option>
            <el-option :label="$t('核销扣减')" value="6"></el-option>
            <el-option :label="$t('删除优惠券')" value="7"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('优惠券类型')">
          <a-select v-model:value="formInline.coupon_type" :placeholder="$t('优惠券类型')" @change="onSearch">
            <el-option :label="$t('全部')" value=""></el-option>
            <el-option :label="$t('折扣券')" value="deduction"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('时间')">
          <el-date-picker
            v-model="formInline.create_time"
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
          <el-table-column prop="coupon_name" :label="$t('优惠券名称')" width="200"></el-table-column>
          <el-table-column prop="serial_no" :label="$t('编号')" width="200"></el-table-column>
          <el-table-column prop="record_type" :label="$t('记录类型')" width="200">
            <template #default="scope">
              <span v-if="scope.row.record_type == 1">{{ $t('首次添加') }}</span>
              <span v-if="scope.row.record_type == 2">{{ $t('调整添加') }}</span>
              <span v-if="scope.row.record_type == 3">{{ $t('调整扣减') }}</span>
              <span v-if="scope.row.record_type == 4">{{ $t('反结账退还') }}</span>
              <span v-if="scope.row.record_type == 5">{{ $t('奖励领取（冻结）') }}</span>
              <span v-if="scope.row.record_type == 6">{{ $t('核销扣减') }}</span>
              <span v-if="scope.row.record_type == 7">{{ $t('删除优惠券') }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('时间')" show-overflow-tooltip> </el-table-column>
          <el-table-column prop="count" :label="$t('数量（张）')" show-overflow-tooltip> </el-table-column>
          <el-table-column prop="left_count" :label="$t('剩余有效张数')" show-overflow-tooltip> </el-table-column>
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
  import MarketingApi from '@/api/marketing.js';
  export default {
    components: {
      /*编辑组件*/
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
          coupon_name: '',
          coupon_type: '',
          record_type: '',
          create_time: [],
        },
        reg_date: [],
        /*是否打开添加弹窗*/
        open_add: false,
        /*是否打开编辑弹窗*/
        open_edit: false,
        /*当前编辑的对象*/
        userModel: {},
        searchLoading: null,
      };
    },
    created() {
      /*获取列表*/
      this.getTableList();
    },
    methods: {
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

      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
          this.getTableList();
        }, 200);
      },

      /*获取列表*/
      getTableList() {
        let self = this;
        let Params = {};
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        Params.coupon_name = self.formInline.coupon_name;
        Params.coupon_type = self.formInline.coupon_type;
        Params.record_type = self.formInline.record_type;
        Params.create_time = self.formInline.create_time;

        MarketingApi.couponRecord(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
          })
          .catch((error) => {});
      },
    },
  };
</script>

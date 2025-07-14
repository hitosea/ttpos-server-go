<template>
  <div class="user">
    <!--搜索表单-->
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item :label="$t('变更场景')">
          <a-select v-model:value="formInline.scene" :placeholder="$t('请选择')" @change="onSearch">
            <el-option :label="$t('全部')" value="0"></el-option>
            <el-option v-for="(item, index) in Scene" :key="index" :label="item.name" :value="item.value"></el-option>
          </a-select>
        </el-form-item>

        <el-form-item :label="$t('昵称/手机号/会员ID')">
          <el-input v-model="formInline.keyword" :placeholder="$t('昵称/手机号/会员ID')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item :label="$t('变动时间')">
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
          <el-table-column prop="log_id" label="ID" width="60"></el-table-column>
          <el-table-column prop="" :label="$t('昵称')">
            <template #default="scope">
              <span>{{ scope.row.user.nickName }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="user.mobile" :label="$t('手机号')" width="160"> </el-table-column>
          <el-table-column prop="user_id" :label="$t('会员ID')" width="80"> </el-table-column>

          <el-table-column prop="value" :label="$t('变动数量')">
            <template #default="scope">
              <p v-if="scope.row.value > 0"> +{{ this.$priceTwo(scope.row.value) }} </p>
              <p v-else>{{ scope.row.value }}</p>
            </template>
          </el-table-column>
          <el-table-column prop="scene.text" :label="$t('变动场景')">
            <template #default="scope">
              <span v-if="!scope.row.scene">-</span>
              <span v-if="scope.row.scene.value == 10" style="color: #409eff">{{ scope.row.scene.text }}</span>
              <span v-if="scope.row.scene.value == 20" style="color: #67c23a">{{ scope.row.scene.text }}</span>
              <span v-if="scope.row.scene.value == 30" style="color: #f56c6c">{{ scope.row.scene.text }}</span>
              <span v-if="scope.row.scene.value == 40" style="color: #e6a23c">{{ scope.row.scene.text }}</span>
              <span v-if="scope.row.scene.value == 50" style="color: #e63c81">{{ scope.row.scene.text }}</span>
              <span v-if="scope.row.scene.value == 90" style="color: #e63c81">{{ scope.row.scene.text }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="describe" :label="$t('描述/说明')" show-overflow-tooltip>
            <template #default="scope">
              {{ description(scope.row.describe) }}
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('变动时间')"></el-table-column>
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
  import Aselect from '@/components/a-select/index.vue';
  import PointsApi from '@/api/points.js';
  import dayjs from '@/utils/dayjs';
  export default {
    components: {
      Aselect,
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
          keyword: '',
          date: '',
          map: '',
          scene: '',
        },
        /*场景*/
        Scene: [],
        searchLoading: '',
      };
    },
    async mounted() {
      // js获取当天时间 日期格式 YYYY-MM-DD
      this.formInline.date = [dayjs(), dayjs()];
      await this.$nextTick();
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

      /*获取列表*/
      getTableList() {
        let self = this;
        let Params = self.formInline;
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        self.loading = true;
        PointsApi.GetUserList(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
            self.Scene = data.data.attributes.scene;
          })
          .catch((error) => {});
      },

      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
          this.getTableList();
        }, 200);
      },

      /**
       * 翻译文本中的关键词
       * @param {string} text 原始文本
       * @returns {string} 翻译后的文本
       */
      description(text) {
        let result = text;
        
        // 翻译关键词映射表（按长度从长到短排序，避免部分替换问题）
        const translateKeywords = [
          '收银机管理员充值赠送操作',
          '收银机管理员操作',
          '撤销会员卡减少积分',
          '发会员卡获取积分',
          '后台管理员扣减',
          '订单积分抵扣',
          '抵扣反结账',
          '订单反结账',
          '邀请有礼奖励',
          '后台管理员',
          '订单赠送',
          '退款扣除',
          '收银充值',
          '充值',
          '操作'
        ];
        
        // 遍历关键词进行替换
        translateKeywords.forEach(keyword => {
          if (result.includes(keyword)) {
            result = result.replace(new RegExp(keyword, 'g'), $t(keyword));
          }
        });
        
        return result;
      },
    },
  };
</script>

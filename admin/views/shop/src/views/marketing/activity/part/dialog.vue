<template>
  <div>
    <el-dialog :title="$t('选择优惠券')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
      <!--搜索表单-->
      <div class="common-search-wrap flex">
        <el-form size="small" :inline="true" :model="form" class="demo-form-inline">
          <el-form-item :label="$t('优惠券名称')">
            <el-input v-model="form.name" :placeholder="$t('请输入优惠券名称')" @input="onSearch"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
              {{ $t('查询') }}
            </el-button>
          </el-form-item>
        </el-form>
      </div>
      <el-table :data="couponList" style="width: 100%">
        <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center" :index="indexMethod"></el-table-column>
        <el-table-column prop="name" :label="$t('优惠券名称')" width="120"></el-table-column>

        <el-table-column prop="type" :label="$t('优惠券类型')">
          <template #default="scope">
            <span v-if="scope.row.type == 'deduction'">{{ $t('折扣券') }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="amount" :label="$t('金额')">
          <template #default="scope">
            <span>{{ $formatPrice(scope.row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="valid_date" :label="$t('有效日期')" width="200"> </el-table-column>
        <el-table-column prop="valid_day_time_range" :label="$t('适用时间')" width="120"> </el-table-column>
        <el-table-column prop="count" :label="$t('数量（张）')"> </el-table-column>
        <el-table-column prop="create_time" :label="$t('添加時間')"> </el-table-column>
        <el-table-column type="radio" width="45" fixed="right">
          <template #default="scope">
            <div class="radio-box">
              <el-radio v-model="selectedRow" :value="scope.row" @change="handleRadioChange(scope.row)">&nbsp;</el-radio>
            </div>
          </template>
        </el-table-column>
      </el-table>
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
      <template #footer>
        <el-button size="small" @click="dialogFormVisible">{{ $t('取消') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup>
  import { ref, reactive, onMounted } from 'vue';
  import MarketingApi from '@/api/marketing.js';

  const emit = defineEmits(['close']);
  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
  });
  const dialogVisible = ref(props.open);
  const form = reactive({
    name: '',
  });
  const couponList = ref([]);
  const curPage = ref(1);
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const selectedRow = ref(null);
  const loading = ref(false);
  const searchLoading = ref(null);

  const handleCurrentChange = (val) => {
    curPage.value = val;
    getCouponList();
  };

  const handleSizeChange = (val) => {
    curPage.value = 1;
    pageSize.value = val;
    getCouponList();
  };

  const dialogFormVisible = () => {
    emit('close');
    dialogVisible.value = false;
    couponList.value = [];
  };

  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      getCouponList();
    }, 200);
  };

  const getCouponList = () => {
    loading.value = true;
    MarketingApi.activityCouponList({
      name: form.name,
      status: 1,
      page: curPage.value,
      list_rows: pageSize.value,
    })
      .then((res) => {
        loading.value = false;
        couponList.value = res.data.data;
        totalDataNumber.value = res.data.total;
      })
      .catch((err) => {
        console.log(err);
      });
  };

  const indexMethod = (index) => {
    return index + 1 + (curPage.value - 1) * pageSize.value;
  };

  const handleRadioChange = (row) => {
    selectedRow.value = row;
    emit('close', row);
  };

  onMounted(() => {
    getCouponList();
  });
</script>
<style lang="scss" scoped></style>

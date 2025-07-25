<template>
  <el-dialog
    :title="rewardType == 0 ? $t('查看发放记录') : $t('查看发放积分记录')"
    v-model="dialogVisible"
    @close="handleClose"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
  >
    <div>
      <div class="record-dialog-wrapper">
        <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
          <el-form-item :label="$t('昵称/手机号/ID/会员卡号')">
            <el-input v-model="formInline.keyword" :placeholder="$t('昵称/手机号/ID/会员卡号')" @input="onSearch"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
              {{ $t('查询') }}
            </el-button>
          </el-form-item>
        </el-form>
      </div>
      <el-table :data="tableData" style="width: 100%">
        <el-table-column prop="nickname" :label="$t('发放对象')">
          <template #default="scope">
            <span>{{ scope.row.nickname }}</span>
            <span>({{ scope.row.member_uuid }})</span>
          </template>
        </el-table-column>
        <el-table-column v-if="rewardType == 1" prop="reward_count" :label="$t('发放次数')">
          <template #default="scope">
            {{ proxy.$priceTwo(scope.row.reward_count) || '-' }}
          </template>
        </el-table-column>
        <el-table-column v-if="rewardType == 1" prop="reward_value" :label="$t('发放总积分')">
          <template #default="scope">
            {{ proxy.$priceTwo(scope.row.reward_value) || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="last_reward_time" :label="$t('发放时间')" />
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
          :loading="loading"
        ></el-pagination>
      </div>
    </div>
  </el-dialog>
</template>
<script setup>
  import { ref, onMounted, getCurrentInstance } from 'vue';
  import MarketingApi from '@/api/marketing.js';

  const { proxy } = getCurrentInstance();

  const emit = defineEmits(['update:recordDialogVisible']);

  const props = defineProps({
    recordDialogVisible: {
      type: Boolean,
      default: false,
    },
    recordUuid: {
      type: [String, Number],
      default: '',
    },
    rewardType: {
      type: Number,
      default: 0,
    },
  });
  const dialogVisible = ref(props.recordDialogVisible);
  const formInline = ref({
    activity_uuid: '',
    keyword: '',
  });
  const tableData = ref([]);
  const curPage = ref(1);
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const loading = ref(false);
  const searchLoading = ref();

  const getTableList = async () => {
    try {
      loading.value = true;
      const res = await MarketingApi.activityRecord({
        activity_uuid: props.recordUuid,
        keyword: formInline.value.keyword,
        page: curPage.value,
        list_rows: pageSize.value,
      });
      tableData.value = res.data.list.data;
      totalDataNumber.value = res.data.list.total;
    } catch (error) {
      console.log(error);
    } finally {
      loading.value = false;
    }
  };

  /*选择第几页*/
  const handleCurrentChange = (val) => {
    if (curPage.value === val) {
      return;
    }
    console.log(val);

    curPage.value = val;
    getTableList();
  };

  /*每页多少条*/
  const handleSizeChange = (val) => {
    console.log(13213213213123);

    curPage.value = 1;
    pageSize.value = val;
    getTableList();
  };

  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      getTableList();
    }, 200);
  };

  const handleClose = () => {
    emit('update:recordDialogVisible', false);
    formInline.value.keyword = '';
    formInline.value.activity_uuid = '';
    tableData.value = [];
    totalDataNumber.value = 0;
    curPage.value = 1;
    pageSize.value = 10;
  };
  onMounted(() => {
    getTableList();
  });
</script>
<style lang="scss" scoped></style>

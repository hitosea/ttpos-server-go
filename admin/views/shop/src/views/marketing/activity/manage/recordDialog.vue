<template>
  <el-dialog :title="$t('查看发放记录')" v-model="dialogVisible" @close="handleClose" :close-on-click-modal="false" :close-on-press-escape="false">
    <div class="record-dialog-wrapper">
      <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item :label="$t('活动名称')">
          <el-input v-model="formInline.card_name" :placeholder="$t('请输入活动名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
    <el-table :data="tableData" style="width: 100%">
      <el-table-column prop="nickname" :label="$t('发放对象')" />
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
      >
      </el-pagination>
    </div>
  </el-dialog>
</template>
<script setup>
  import { ref, onMounted } from 'vue';
  import MarketingApi from '@/api/marketing.js';

  const emit = defineEmits(['update:recordDialogVisible']);

  const props = defineProps({
    recordDialogVisible: {
      type: Boolean,
      default: false,
    },
    recordUuid: {
      type: String,
      default: '',
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
    const res = await MarketingApi.activityRecord({
      activity_uuid: props.recordUuid,
      keyword: formInline.value.keyword,
      page: curPage.value,
      list_rows: pageSize.value,
    });
    try {
      tableData.value = res.data.list.data;
      totalDataNumber.value = res.data.list.total;
    } catch (error) {
      console.log(error);
    } finally {
      loading.value = false;
    }
  };

  const handleSizeChange = (size) => {
    pageSize.value = size;
    getTableList();
  };

  const handleCurrentChange = (page) => {
    curPage.value = page;
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

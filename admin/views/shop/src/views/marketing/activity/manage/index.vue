<template>
  <div class="user">
    <!--搜索表单-->
    <div class="common-search-wrap flex">
      <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item :label="$t('活动名称')">
          <el-input v-model="formInline.name" :placeholder="$t('请输入活动名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <!--添加等级-->
      <div class="common-level-rail">
        <el-button size="small" type="primary" @click="addClick" icon="Plus" v-auth="'/marketing/activity/add'">{{ $t('新增活动') }} </el-button>
      </div>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="name_text" :label="$t('活动名称')"></el-table-column>
          <el-table-column prop="type" :label="$t('活动类型')">
            <template #default="scope">
              <span v-if="scope.row.type == 0">{{ $t('邀请消费有礼') }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="reward_type" :label="$t('活动奖品')">
            <template #default="scope">
              <span v-if="scope.row.reward_type == 0 && scope.row.prizes.length > 0">{{ scope.row.prizes[0].coupon_name }}</span>
              <span v-else>
                {{ $t('积分') }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('创建时间')"></el-table-column>

          <el-table-column prop="status_text" :label="$t('状态')"> </el-table-column>

          <el-table-column :label="$t('活动时间')" width="200">
            <template #default="scope">
              <span>{{ scope.row.start_time }} ~ {{ scope.row.end_time }}</span>
            </template>
          </el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="200">
            <template #default="scope">
              <el-button @click="sendClick(scope.row)" type="primary" link size="small">{{ $t('发放记录') }} </el-button>
              <el-button @click="editClick(scope.row)" :disabled="scope.row.status == 2 || scope.row.headquarter_uuid > 0" type="primary" link size="small" v-auth="'/marketing/activity/edit'"
                >{{ $t('编辑') }}
              </el-button>
              <el-button @click="disableClick(scope.row)" :disabled="scope.row.status == 2 || scope.row.headquarter_uuid > 0" type="primary" link size="small" v-auth="'/marketing/activity/disable'"
                >{{ $t('失效') }}
              </el-button>
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
    <record-dialog
      v-if="recordDialogVisible"
      :recordDialogVisible="recordDialogVisible"
      :recordUuid="recordUuid"
      :rewardType="rewardType"
      @update:recordDialogVisible="recordDialogVisible = $event"
    />
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
  import { useRouter } from 'vue-router';
  import { ElMessageBox, ElMessage } from 'element-plus';
  import MarketingApi from '@/api/marketing.js';
  import { useUserStore } from '@/store/index';
  import recordDialog from './recordDialog.vue';
  const { proxy } = getCurrentInstance();

  // 获取路由实例
  const router = useRouter();

  // 获取用户store
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;

  // 响应式数据
  const loading = ref(true);
  const tableData = ref([]);
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const curPage = ref(1);
  const searchLoading = ref('');
  const recordDialogVisible = ref(false);
  const recordUuid = ref('');
  const rewardType = ref(0);
  // 表单数据
  const formInline = reactive({
    name: '',
    status: '',
  });

  const handleCurrentChange = (val) => {
    curPage.value = val;
    loading.value = true;
    getTableList();
  };

  const handleSizeChange = (val) => {
    curPage.value = 1;
    pageSize.value = val;
    getTableList();
  };

  const getTableList = () => {
    let params = { ...formInline };
    params.page = curPage.value;
    params.list_rows = pageSize.value;

    loading.value = true;
    MarketingApi.activityList(params, true)
      .then((data) => {
        loading.value = false;
        tableData.value = data.data.list.data;
        totalDataNumber.value = data.data.list.total;
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      getTableList();
    }, 200);
  };

  const memberAuth = () => {
    if (!proxy.$filter.isAuth('/card/user/index')) {
      ElMessage.error($t('该营销活动需商家开通会员中心功能，请联系销售代表处理'));
      return false;
    }
    return true;
  };

  const addClick = () => {
    if (!memberAuth()) {
      return;
    }
    router.push('/' + app_id + '/marketing/activity/add');
  };

  const sendClick = (row) => {
    if (!memberAuth()) {
      return;
    }
    rewardType.value = row.reward_type;
    recordUuid.value = row.uuid;
    recordDialogVisible.value = true;
  };

  const editClick = (item) => {
    if (!memberAuth()) {
      return;
    }
    router.push({
      path: '/' + app_id + '/marketing/activity/edit',
      query: {
        uuid: item.uuid,
      },
    });
  };

  const disableClick = (row) => {
    if (!memberAuth()) {
      return;
    }
    ElMessageBox.confirm(window.$t('失效后将不可恢复，活动变为已结束状态，确定将该活动失效?'), window.$t('失效营销活动'), {
      confirmButtonText: window.$t('确定'),
      cancelButtonText: window.$t('取消'),
      type: 'warning',
    })
      .then(() => {
        // 这里可以添加具体的删除API调用
        MarketingApi.activityDisable({ uuid: row.uuid }, true)
          .then((res) => {
            ElMessage.success(window.$t('失效成功'));
            getTableList();
          })
          .catch((error) => {
            loading.value = false;
          });
      })
      .catch(() => {});
  };

  // 组件挂载时执行
  onMounted(() => {
    getTableList();
  });
</script>

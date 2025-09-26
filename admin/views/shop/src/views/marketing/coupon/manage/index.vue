<template>
  <div class="user">
    <!--搜索表单-->
    <div class="common-search-wrap flex">
      <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item :label="$t('优惠券名称')">
          <el-input v-model="formInline.name" :placeholder="$t('优惠券名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item :label="$t('优惠券类型')">
          <a-select v-model:value="formInline.type" :placeholder="$t('优惠券类型')" @change="onSearch">
            <el-option :label="$t('全部')" value=""></el-option>
            <el-option :label="$t('折扣券')" value="deduction"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <div class="common-level-rail">
        <el-button type="primary" v-auth="'/marketing/coupon/add'" icon="Plus" @click="addMenber">{{ $t('添加优惠券') }}</el-button>
      </div>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="sort" :label="$t('排序')" width="80"></el-table-column>
          <el-table-column prop="name" :label="$t('优惠券名称')"></el-table-column>

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
          <el-table-column prop="valid_date" :label="$t('有效日期')"> </el-table-column>
          <el-table-column prop="valid_day_time_range" :label="$t('适用时间')"> </el-table-column>
          <el-table-column prop="count" :label="$t('数量（张）')"> </el-table-column>
          <el-table-column prop="status" :label="$t('状态')">
            <template #default="scope">
              <el-switch v-model="scope.row.status" :active-value="1" :inactive-value="0" @click="changeStatus(scope.row)" />
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加時間')"> </el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="160">
            <template #default="scope">
              <el-button v-auth="'/marketing/coupon/edit'" @click="editClick(scope.row)" type="primary" link size="small">{{ $t('编辑') }}</el-button>
              <el-button @click="deleteClick(scope.row)" type="primary" link size="small">{{ $t('删除') }}</el-button>
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
        ></el-pagination>
      </div>
    </div>

    <AddEdit v-if="open_addEdit" :title="title" :editData="editData" :open="open_addEdit" @closeDialog="closeAddMenber"> </AddEdit>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
  import MarketingApi from '@/api/marketing.js';
  import AddEdit from './dialog/addEdit.vue';
import { message } from '@/utils/message';

  // 获取全局属性
  const { proxy } = getCurrentInstance();
  const { $t, $formatPrice, $priceTwo } = proxy;

  // 响应式数据
  const loading = ref(true);
  const tableData = ref([]);
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const curPage = ref(1);

  const formInline = reactive({
    name: '',
    type: '',
  });

  const open_addEdit = ref(false);
  const title = ref('');
  const editData = ref('');
  const searchLoading = ref('');

  // 方法
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
    const params = { ...formInline };
    params.page = curPage.value;
    params.list_rows = pageSize.value;
    loading.value = true;

    MarketingApi.couponList(params, true)
      .then((data) => {
        loading.value = false;
        tableData.value = data.data.list.data;
        totalDataNumber.value = data.data.list.total;
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  const addMenber = () => {
    title.value = $t('添加优惠券');
    open_addEdit.value = true;
  };

  const closeAddMenber = (e) => {
    open_addEdit.value = false;
    editData.value = '';
    if (e == 1) {
      getTableList();
    }
  };

  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      getTableList();
    }, 200);
  };

  const editClick = (item) => {
    title.value = $t('编辑优惠券');
    editData.value = item;
    open_addEdit.value = true;
  };

  const deleteClick = (item) => {
    ElMessageBox.confirm($t('确定删除该优惠券吗？'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
    }).then(() => {
      let params = {
        uuid: item.uuid,
      };
      loading.value = true;
      MarketingApi.couponDelete(params, true)
        .then((res) => {
          getTableList();
        })
        .catch((error) => {
          console.log(error);
        })
        .finally(() => {
          loading.value = false;
        });
    });
  };

  const changeStatus = (item) => {
    let params = {
      uuid: item.uuid,
      status: item.status == 0 ? 0 : 1,
    };
    loading.value = true;
    MarketingApi.couponStatus(params, true)
      .then((res) => {
        getTableList();
        message.success(res.msg);
      })
      .catch((error) => {
        console.log(error);
      })
      .finally(() => {
        loading.value = false;
      });
  };

  // 生命周期
  onMounted(() => {
    getTableList();
  });
</script>

<template>
  <div>
    <div class="common-search-wrap">
      <!--订单进度-->
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item label="">
          <el-radio-group v-model="searchForm.time_type" class="radio-search" @change="timeTypeChange">
            <el-radio-button label="1">{{ $t('今天') }}</el-radio-button>
            <el-radio-button label="2">{{ $t('昨天') }}</el-radio-button>
            <el-radio-button label="3">{{ $t('本周') }}</el-radio-button>
            <el-radio-button label="0">{{ $t('全部') }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('起始时间')">
          <div class="block">
            <el-date-picker
              size="small"
              v-model="searchForm.date"
              type="datetimerange"
              format="YYYY-MM-DD HH:mm"
              value-format="YYYY-MM-DD HH:mm"
              time-format="HH:mm"
              range-separator="~"
              :start-placeholder="$t('开始日期')"
              :end-placeholder="$t('结束日期')"
              clearable
              @change="dateChange"
            ></el-date-picker>
          </div>
        </el-form-item>
        <el-form-item :label="$t('选择区域')">
          <a-select size="small" v-model:value="searchForm.area_id" :placeholder="$t('请选择')" @change="onSearch">
            <el-option :label="$t('全部')" value=""></el-option>
            <el-option v-for="(item, index) in area_list" :key="index" :label="item.area_name" :value="item.area_id"> </el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('商品分类')">
          <a-cascader
            :options="categoryList"
            :props="{ checkStrictly: true, expandTrigger: 'hover' }"
            v-model:value="searchForm.category_id"
            :placeholder="$t('请选择分类')"
            @change="onSearch"
          >
            <template #default="{ data }">
              <span class="span-w" @click="handleValue(data)">{{ data.label }}</span>
            </template>
          </a-cascader>
        </el-form-item>
        <el-form-item :label="$t('商品名称')">
          <el-input size="small" :placeholder="$t('请输入商品名称')" v-model="searchForm.product_name" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" @click="onExport">{{ $t('导出') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="table-wrap">
      <el-table
        size="small"
        :data="tableData"
        border
        style="width: 100%"
        v-loading="loading"
        :default-sort="{ prop: 'product_num', order: 'descending' }"
        @sort-change="sortChange"
      >
        <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" :index="indexMethod" align="center"></el-table-column>
        <el-table-column prop="product_name_text" :label="$t('商品名称')" minWidth="180"></el-table-column>
        <el-table-column prop="category_name_text" :label="$t('商品分类')" minWidth="180"></el-table-column>
        <el-table-column prop="product_num" :label="$t('销售数量')" minWidth="120" sortable :sort-method="customSortMethod">
          <template #default="scope">
            {{ Number(scope.row.product_num) }}
          </template>
        </el-table-column>
        <el-table-column prop="sales_price" :label="$t('原价销售额')" minWidth="120" sortable :sort-method="salesSortMethod">
          <template #header="scope">
            {{ $t('原价销售额') }}
            <el-tooltip effect="dark" placement="top" :content="$t('商品的原价销售额包括商品售价，包含加料、税费')">
              <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
            </el-tooltip>
          </template>
          <template #default="scope">
            {{ proxy.$formatPrice(scope.row.sales_price) }}
          </template>
        </el-table-column>
        <el-table-column prop="free_product_num" :label="$t('赠菜')" minWidth="100"></el-table-column>
        <el-table-column prop="received_price" :label="$t('实际销售额')" minWidth="120">
          <template #default="scope">
            {{ proxy.$formatPrice(scope.row.received_price) }}
          </template>
        </el-table-column>
        <el-table-column prop="business_price" :label="$t('营业收入')" minWidth="120">
          <template #default="scope">
            {{ proxy.$formatPrice(scope.row.business_price) }}
          </template>
        </el-table-column>
      </el-table>
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
  </div>
</template>
<script setup>
  import datePick from '@/components/datePick/datePick.vue';
  import StatisticsApi from '@/api/statistics.js';
  import { ref, onMounted, getCurrentInstance } from 'vue';
  import { languageStore } from '@/store/model/language.js';
  import qs from 'qs';
  import { useUserStore } from '@/store';
  import { message } from '@/utils/message.js';

  const { proxy } = getCurrentInstance();
  const token = ref(useUserStore().token);
  const languageTag = ref(languageStore().language);
  const datePickRef = ref(null);
  const loading = ref(null);
  const searchLoading = ref(false);
  const curPage = ref(1);
  const pageSize = ref(20);
  const totalDataNumber = ref(0);
  const searchForm = ref({
    date: ['', ''],
    time_type: '1',
    area_id: '',
    category_id: '',
    product_name: '',
    sort_field: 'product_num',
    sort_type: 'desc',
  });

  const tableData = ref([]);
  const area_list = ref([]);
  const categoryList = ref([]);
  const cascader = ref(null);

  const timeTypeChange = () => {
    searchForm.value.date = ['', ''];
    onSearch();
  };

  const dateChange = () => {
    searchForm.value.time_type = '-1';
    onSearch();
  };

  const handleValue = (data) => {
    searchForm.value.category_id = [];
    searchForm.value.category_id = data.value;
    cascader.value.togglePopperVisible();
    onSearch();
  };

  const onSearch = () => {
    clearTimeout(searchLoading.value);

    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      getData();
    }, 200);
  };

  /*选择第几页*/
  const handleCurrentChange = (val) => {
    curPage.value = val;
    getData();
  };

  /*每页多少条*/
  const handleSizeChange = (val) => {
    curPage.value = 1;
    pageSize.value = val;
    tgetData();
  };

  const getData = () => {
    let self = this;
    loading.value = true;
    let params = searchForm.value;
    if (typeof params.category_id == 'object' && params.category_id) {
      params.category_id = Number(params.category_id[params.category_id.length - 1]);
    }
    if (params.date && params.date[0] && params.date[1]) {
      params.date[0] = `${params.date[0]}:00`
      params.date[1] = `${params.date[1]}:59`
    }
    params.list_rows = pageSize.value;
    params.page = curPage.value;
    StatisticsApi.getProductSell(params, true)
      .then((data) => {
        tableData.value = data.data.list.data;
        totalDataNumber.value = data.data.list.total;
        area_list.value = data.data.area_list;
        categoryList.value = [];
        data.data.category_list.map((item, index) => {
          categoryList.value.push({
            value: item.category_id,
            label: item.name_text,
            children: [],
          });
          item.child.map((items, indexs) => {
            categoryList.value[index].children.push({
              value: items.category_id,
              label: items.name_text,
            });
          });
        });
        loading.value = false;
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  const indexMethod = (index) => {
    return (curPage.value - 1) * pageSize.value + index + 1;
  };

  const onExport = () => {
    if (searchForm.value.date[0] && searchForm.value.date[1]) {
      let baseUrl = window.location.protocol + '//' + window.location.host;
      let form = searchForm.value;
      form.token = token.value;
      window.location.href = baseUrl + '/index.php/shop/store.product/export?' + qs.stringify(form) + '&language=' + languageTag.value;
    } else {
      message({
        message: $t('请选择时间'),
        type: 'warning',
      });
    }
  };

  const sortChange = (column) => {
    searchForm.value.sort_field = column.prop;
    if (column.order == 'descending') {
      searchForm.value.sort_type = 'desc';
    } else if (column.order == 'ascending') {
      searchForm.value.sort_type = 'asc';
    } else {
      searchForm.value.sort_field = '';
      searchForm.value.sort_type = '';
    }
    onSearch();
  };

  const customSortMethod = (a, b) => {
    return Number(a.product_num) - Number(b.product_num);
  };
  const salesSortMethod = (a, b) => {
    return Number(a.sales_price) - Number(b.sales_price);
  };

  onMounted(() => {
    const today = new Date();
    const startOfDay = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')} 00:00`;
    const endOfDay = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')} 23:59`;
    searchForm.value.date = [startOfDay, endOfDay];
    getData();
  });
</script>
<style lang="scss" scoped>
  .radio-search {
    :deep(.el-radio-button) {
      margin-right: -3px;
      margin-bottom: 0;

      .el-radio-button__inner {
        padding: 8px 11px !important;
      }
    }
  }
  .span-w {
    display: block;
    width: 100%;
  }
  .data-box-icon {
    color: var(--el-color-tips);
    padding-top: 2px;
  }
  .flex {
    display: flex;
    gap: 4px;
    align-items: center;
  }
</style>

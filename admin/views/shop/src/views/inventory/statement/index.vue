<template>
  <div class="product-content" v-loading="loading">
    <div class="">
      <!--订单进度-->
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('选择月份')" class="w-230">
          <el-date-picker style="width: 100%" v-model="searchForm.date" type="month" value-format="YYYY-MM" :placeholder="$t('选择月份')" @change="onSearch"></el-date-picker>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="common-form">{{ $t('库存信息') }}</div>
    <div class="operation-data">
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('月初原有库存') }}</h3>
          <SvgIcon class="data-box-icon" name="icon1"></SvgIcon>
        </div>
        <h4>{{ data.info.month_start_stock || 0 }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('月入库数量') }}</h3>
          <SvgIcon class="data-box-icon" name="icon1"></SvgIcon>
        </div>
        <h4>{{ data.info.month_entry_stock || 0 }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('月出库数量') }}</h3>
          <SvgIcon class="data-box-icon" name="icon1"></SvgIcon>
        </div>
        <h4>{{ data.info.month_exit_stock || 0 }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('月末剩余库存') }}</h3>
          <SvgIcon class="data-box-icon" name="icon1"></SvgIcon>
        </div>
        <h4>{{ data.info.month_end_stock || 0 }}</h4>
      </div>
    </div>

    <div class="common-form">{{ $t('损耗信息') }}</div>

    <div class="operation-data">
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('损耗数量') }}</h3>
          <SvgIcon class="data-box-icon" name="icon4"></SvgIcon>
        </div>
        <h4>{{ data.info.month_damaged_num || 0 }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('损耗比例') }}</h3>
          <SvgIcon class="data-box-icon" name="icon4"></SvgIcon>
        </div>
        <h4>{{ data.info.month_damaged_percent || 0 }}%</h4>
      </div>
    </div>

    <!--内容-->
    <div class="product-bottom">
      <div class="flex-1">
        <div class="right-box d-s-s d-c">
          <div class="list ww100">
            <div class="common-form">
              {{ $t('损耗TOP10') }}
              <div class="common-form-r">
                <span @click="handleClick(0)" :class="acvite == 0 ? 'active' : ''"> {{ $t('从高到低') }}</span>
                <span @click="handleClick(1)" :class="acvite == 1 ? 'active' : ''"> {{ $t('从低到高') }}</span>
              </div>
            </div>
            <el-table v-if="data.damaged_list.length > 0" :data="data.damaged_list" style="width: 100%" size="small">
              <el-table-column prop="product_name" :label="$t('商品名称')">
                <template #default="scope">
                  <div class="product-name">
                    <span :class="scope.$index < 3 ? 'key-box' : 'key-box2'">{{ scope.$index + 1 }}</span>
                    <span class="">{{ scope.row.product_name }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="damage_count" :label="$t('损耗')"> </el-table-column>
              <el-table-column prop="damage_ratio" :label="$t('比例')">
                <template #default="scope"> {{ this.$formatPrice(scope.row.damage_ratio) }}% </template>
              </el-table-column>
            </el-table>
            <div v-else class="tc pt30">{{ $t('暂无上榜记录') }}</div>
          </div>
        </div>
      </div>

      <div class="flex-1">
        <div class="right-box d-s-s d-c">
          <div class="list ww100">
            <div class="common-form">{{ $t('滞销TOP10') }}</div>
            <el-table v-if="data.unsalable_list.length > 0" :data="data.unsalable_list" style="width: 100%" size="small">
              <el-table-column prop="product_name_text" :label="$t('商品名称')">
                <template #default="scope">
                  <div class="product-name">
                    <span :class="scope.$index < 3 ? 'key-box' : 'key-box2'">{{ scope.$index + 1 }}</span>
                    <span class="">{{ scope.row.product_name_text }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="total_num" :label="$t('销量')"> </el-table-column>
            </el-table>
            <div v-else class="tc pt30">{{ $t('暂无上榜记录') }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
  import InventoryApi from '@/api/inventory.js';
  const date = new Date();
  const year = date.getFullYear();
  const month = date.getMonth() + 1; // Add 1 because getMonth() returns 0-11
  const formattedMonth = `${year}-${month.toString().padStart(2, '0')}`;

  export default {
    data() {
      return {
        loading: false,
        searchForm: {
          date: formattedMonth,
        },
        salesNumRank: [],
        data: {
          info: {
            month_damaged_num: 0,
            month_damaged_percent: 0,
            month_end_stock: 0,
            month_entry_stock: 0,
            month_exit_stock: 0,
            month_start_stock: 0,
          },
          damaged_list: [],
          unsalable_list: [],
        },
        acvite: 0,
        searchLoading: '',
      };
    },
    mounted() {
      this.getData();
    },
    methods: {
      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.getData();
        }, 200);
      },

      getData() {
        let self = this;
        let Params = {};
        Params.date = self.searchForm.date;
        self.loading = true;
        self.echartsVueShow = false;
        InventoryApi.getMonthlyStatistics(Params, true)
          .then((data) => {
            self.loading = false;
            this.data = data.data;
          })
          .catch((error) => {});
      },

      handleClick(index) {
        this.acvite = index;
        if (index == 0) {
          this.data.damaged_list.sort((a, b) => {
            // 按照 field 字段进行升序排序
            return b.damage_ratio - a.damage_ratio;
          });
        } else {
          this.data.damaged_list.sort((a, b) => {
            // 按照 field 字段进行升序排序
            return a.damage_ratio - b.damage_ratio;
          });
        }
      },
    },
  };
</script>
<style lang="scss" scoped>
  .w-230 {
    width: 230px;
  }

  .operation-data {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 16px;
  }

  .data-box {
    flex: 1;
    padding: 16px;
    background: #fff6de;
    border-radius: 4px;
  }

  .data-box-title {
    display: flex;
    justify-content: space-between;
  }

  .data-box-title h3 {
    font-size: 16px;
    font-weight: 400;
    color: var(--el-color-black);
  }

  .data-box h4 {
    font-size: 28px;
    font-weight: 700;
    margin-top: 20px;
    color: var(--el-color-black);
  }

  .data-box:hover {
    background: #ffbe00;
  }

  .data-box:hover .data-box-icon {
    color: #fff;
  }

  .data-box-icon {
    width: 24px;
    height: 24px;
    color: #ffbe00;
  }

  .product-bottom {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-top: 16px;
  }

  .product-name {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .key-box {
    flex-shrink: 0;
    display: block;
    width: 20px;
    height: 20px;
    font-size: 12px;
    line-height: 20px;
    border-radius: 50%;
    font-weight: bold;
    text-align: center;
    color: var(--el-color-black);
    background: var(--el-color-primary);
    font-weight: 700;
  }

  .key-box2 {
    flex-shrink: 0;
    display: block;
    width: 20px;
    height: 20px;
    font-size: 12px;
    line-height: 20px;
    border-radius: 50%;
    font-weight: bold;
    text-align: center;
    color: var(--el-color-primary);
    background: #fff6de;
    font-weight: 700;
  }

  .common-form {
    display: flex;
    align-items: center;

    .common-form-r {
      margin-left: auto;
      display: flex;
      align-items: center;
      gap: 16px;

      span {
        font-size: 14px;
        color: var(--el-color-black);
        cursor: pointer;
      }

      span.active {
        color: var(--el-color-primary);
        font-weight: bold;
      }
    }
  }
</style>

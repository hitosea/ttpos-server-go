<template>
  <!--内容-->
  <div class="product-content" v-loading="loading">
    <div class="common-search-wrap">
      <!--订单进度-->
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('起始时间')">
          <div class="block">
            <el-date-picker
              size="small"
              v-model="searchForm.date"
              type="daterange"
              value-format="YYYY-MM-DD"
              range-separator="~"
              :start-placeholder="$t('开始日期')"
              :end-placeholder="$t('结束日期')"
              clearable
              @change="onSearch"
            ></el-date-picker>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button v-auth="'/statistics/survey/export'" size="small" type="primary" @click="onExport">{{ $t('导出') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="common-form">{{ $t('店内概况') }}</div>

    <div class="operation-data">
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('总销售额') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('总销售额：') }}</span
              ><br />
              <span>{{ $t('商品已含税（原商品金额+实收服务费+实收服务税费+实收支付手续费）') }}</span
              ><br />
              <span>{{ $t('商品未含税（原商品金额+实收服务费+实收商品及服务税费+实收支付手续费）') }}</span
              ><br />
              <span>{{ $t('（以上计算都不包含会员充值金额）') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ this.$formatPrice(detail.receivable_price || 0) }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('实收金额') }}</h3>
          <el-tooltip
            class="item"
            effect="dark"
            :content="$t('指订单原价扣除优惠折扣、会员折扣后的金额，包含支付手续费+收银机会员充值金额（不包括会员余额消费金额）')"
            placement="bottom"
          >
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ this.$formatPrice(detail.received_price || 0) }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('商品数量') }}</h3>
          <el-tooltip class="item" effect="dark" :content="$t('所售卖的商品数量，仅计算普通商品及自助餐（不包括自助餐加钟）')" placement="bottom">
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ detail.product_num || 0 }}</h4>
      </div>
      <div class="data-box" v-if="is_open_member == '1'">
        <div class="data-box-title">
          <h3 v-if="is_open_member == '1'">{{ $t('新增会员数/会员折扣') }}</h3>
          <h3 v-else>{{ $t('新增会员数') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('新增会员数：新增会员数量') }}</span>
              <br />
              <span>{{ $t('会员折扣：会员折扣总金额（包含等级折扣及会员卡折扣）') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ (detail.user_count || 0) + (is_open_member == '1' ? '/' + (detail.user_discount_money || 0) : '') }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('营业收入') }}</h3>
          <el-tooltip class="item" effect="dark" :content="$t('营业收入=实收金额-税费+会员余额消费金额')" placement="bottom">
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ this.$formatPrice(detail.business_price || 0) }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('服务费') }}/{{ $t('支付手续费') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('服务费：退款后将减少退款单所含的服务费') }}</span>
              <br />
              <span>{{ $t('支付手续费：使用需另付手续费的支付方式，所增加的手续费，整单退款后将减少退款单所含的支付手续费') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ this.$formatPrice(detail.service_money || 0) }}/{{ this.$formatPrice(detail.pay_fee_money || 0) }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('税费') }}</h3>
          <el-tooltip class="item" effect="dark" :content="$t('税费，包含商品税费、服务费税费')" placement="bottom">
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ this.$formatPrice(detail.consumption_tax_money || 0) }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('退款金额') }}</h3>
          <el-tooltip class="item" effect="dark" :content="$t('订单退款的金额（反结账不计入在此）')" placement="bottom">
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ this.$formatPrice(detail.refund_money || 0) }}</h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('优惠折扣/优惠占比') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('优惠折扣=优惠折扣总金额（包含改价、折扣比例、抹零、结账抹零）') }}</span>
              <br />
              <span>{{ $t('优惠占比=优惠折扣/总销售额') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>
          {{ (detail.discount_money || 0) + ('/' + (detail.discount_ratio || 0)) }}
        </h4>
      </div>

      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('赠菜总额/赠菜数量') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('赠菜总额：赠菜的总金额') }}</span>
              <br />
              <span>{{ $t('赠菜数量：赠菜的数量') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>
          {{ (detail.free_product_price || 0) + ('/' + (detail.free_product_num || 0)) }}
        </h4>
      </div>
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('免单总额/免单数量') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('免单总额：免单的总金额') }}</span>
              <br />
              <span>{{ $t('免单数量：免单的数量') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>
          {{ (detail.free_order_price || 0) + ('/' + (detail.free_order_num || 0)) }}
        </h4>
      </div>

      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('充值金额') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('会员充值金额（实际充值的金额，不包含赠送金额）') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>
          {{ this.$formatPrice(detail.recharge_amount || 0) }}
        </h4>
      </div>

      <div class="data-box" v-if="delivery_status == '1'">
        <div class="data-box-title">
          <h3>{{ $t('外送销售') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('顾客通过在线外卖下单的订单总额') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>
          {{ this.$formatPrice(detail.delivery_order_amount || 0) }}
        </h4>
      </div>
      <div class="data-box" v-if="delivery_status == '1'">
        <div class="data-box-title">
          <h3>{{ $t('外送营收') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('顾客通过在线外卖下单的订单总额-外卖订单退款总额-配送费') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>
          {{ this.$formatPrice(detail.delivery_order_revenue || 0) }}
        </h4>
      </div>
      <div class="data-box" v-if="delivery_status == '1'">
        <div class="data-box-title">
          <h3>{{ $t('外送退款') }}</h3>
        </div>
        <h4>
          {{ this.$formatPrice(detail.delivery_order_refund_amount || 0) }}
        </h4>
      </div>
      <div class="data-box" v-if="delivery_status == '1'">
        <div class="data-box-title">
          <h3>{{ $t('外送配送费') }}</h3>
        </div>
        <h4>
          {{ this.$formatPrice(detail.delivery_fee || 0) }}
        </h4>
      </div>
    </div>

    <div class="area-data">
      <div class="area-data-item">
        <div class="common-form">{{ $t('区域数据') }}</div>
        <template v-if="regionData.length > 0">
          <div class="area-data-item-content">
            <div class="area-data-item-content-item" v-for="(item, index) in regionData" :key="index">
              <div class="area-data-item-content-item-title">
                <p>{{ item.area_name }}</p>
              </div>
              <div class="area-data-item-content-item-content">
                <div class="area-data-item-content-item-content-wrap">
                  <p>
                    {{ $t('总销售额') }}
                  </p>
                  <p>
                    {{ this.$formatPrice(item.sales_price || 0) }}
                  </p>
                </div>
                <div class="area-data-item-content-item-content-wrap">
                  <p>
                    {{ $t('营业收入') }}
                  </p>
                  <p>
                    {{ this.$formatPrice(item.business_price || 0) }}
                  </p>
                </div>
                <div class="area-data-item-content-item-content-wrap">
                  <p>
                    {{ $t('商品数量') }}
                  </p>
                  <p>
                    {{ this.$formatPrice(item.product_num || 0) }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </template>
        <div v-else>{{ $t('暂无数据') }}</div>
      </div>
    </div>

    <div class="order-form flex">
      <div class="order-box">
        <div class="common-form">{{ $t('订单数据') }}</div>
        <div class="order-box-l">
          <div class="box-l-title">
            <div class="title-box">
              <h3>{{ $t('合计订单数') }}</h3>
              <h4>{{ this.$formatPrice(detail.total_order_num || 0) }}</h4>
            </div>
            <div class="title-box">
              <h3>{{ $t('最小订单金额') }}</h3>
              <h4>{{ this.$formatPrice(detail.min_order_price || 0) }}</h4>
            </div>
            <div class="title-box">
              <h3>{{ $t('最大订单金额') }}</h3>
              <h4>{{ this.$formatPrice(detail.max_order_price || 0) }}</h4>
            </div>
            <div class="title-box">
              <h3>{{ $t('平均订单金额') }}</h3>
              <h4>{{ this.$formatPrice(detail.avg_order_price || 0) }}</h4>
            </div>
          </div>
          <div class="box-main">
            <h4>{{ $t('桌台方式') }}</h4>
            <div class="main-body">
              <div class="main-div">
                <h3>{{ $t('桌数') }}</h3>
                <h4>{{ this.$formatPrice(detail.table_order_num || 0) }}</h4>
              </div>
              <div class="main-div">
                <h3>{{ $t('人数') }}</h3>
                <h4>{{ this.$formatPrice(detail.table_people_num || 0) }}</h4>
              </div>
              <div class="main-div">
                <h3>{{ $t('最小/大订单金额') }}</h3>
                <h4>
                  {{ this.$formatPrice(detail.table_min_order_price || 0) + '/' + this.$formatPrice(detail.table_max_order_price || 0) }}
                </h4>
              </div>
              <div class="main-div">
                <h3>{{ $t('平均订单金额') }}</h3>
                <h4>
                  {{ this.$formatPrice(detail.table_avg_order_price || 0) }}
                </h4>
              </div>
            </div>
          </div>
          <div class="box-main">
            <h4>{{ $t('点餐方式') }}</h4>
            <div class="main-body">
              <div class="main-div">
                <h3>{{ $t('订单数') }}</h3>
                <h4>{{ this.$formatPrice(detail.cashier_order_num || 0) }}</h4>
              </div>

              <div class="main-div">
                <h3>{{ $t('最小/大订单金额') }}</h3>
                <h4>
                  {{ this.$formatPrice(detail.cashier_min_order_price || 0) + '/' + this.$formatPrice(detail.cashier_max_order_price || 0) }}
                </h4>
              </div>
              <div class="main-div">
                <h3>{{ $t('平均订单金额') }}</h3>
                <h4>
                  {{ this.$formatPrice(detail.cashier_avg_order_price || 0) }}
                </h4>
              </div>
            </div>
          </div>
          <div class="box-main" v-if="delivery_status == '1'">
            <h4>{{ $t('外送点餐') }}</h4>
            <div class="main-body">
              <div class="main-div">
                <h3>{{ $t('订单数') }}</h3>
                <h4>{{ this.$formatPrice(detail.delivery_order_num || 0) }}</h4>
              </div>

              <div class="main-div">
                <h3>{{ $t('最小/大订单金额') }}</h3>
                <h4>
                  {{ this.$formatPrice(detail.delivery_min_order_price || 0) + '/' + this.$formatPrice(detail.delivery_max_order_price || 0) }}
                </h4>
              </div>
              <div class="main-div">
                <h3>{{ $t('平均订单金额') }}</h3>
                <h4>
                  {{ this.$formatPrice(detail.delivery_avg_order_price || 0) }}
                </h4>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="order-box">
        <div class="common-form">{{ $t('支付数据') }}</div>
        <div class="pay-data">
          <el-table v-if="incomes.length > 0" :data="incomes" style="width: 100%" size="small">
            <el-table-column prop="pay_type_name" :label="$t('支付方式')"> </el-table-column>
            <el-table-column prop="order_num" :label="$t('订单数')"> </el-table-column>
            <el-table-column prop="price" :label="$t('金额')">
              <template #default="scope">
                {{ this.$formatPrice(scope.row.price) }}
              </template>
            </el-table-column>
          </el-table>
          <div v-else class="tc p30">{{ $t('暂无数据') }}</div>
        </div>
      </div>
    </div>

    <!--内容-->
    <div class="product-bottom">
      <div class="flex-1">
        <div class="right-box d-s-s d-c pr16">
          <div class="common-form">{{ $t('销量TOP10') }}</div>
          <div class="list ww100">
            <el-table v-if="salesNumRank.length > 0" :data="salesNumRank" style="width: 100%" size="small">
              <el-table-column prop="product_name_text" :label="$t('商品名称')">
                <template #default="scope">
                  <div class="product-name">
                    <span :class="scope.$index < 3 ? 'key-box' : 'key-box2'">{{ scope.$index + 1 }}</span>
                    <span class="">{{ scope.row.product_name_text }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="total_num" :label="$t('销量')"> </el-table-column>
              <el-table-column prop="total_price" :label="$t('销售额')">
                <template #default="scope">
                  {{ this.$formatPrice(scope.row.total_price) }}
                </template>
              </el-table-column>
            </el-table>
            <div v-else class="tc pt30">{{ $t('暂无上榜记录') }}</div>
          </div>
        </div>
      </div>
      <div class="flex-1">
        <div class="right-box d-s-s d-c pr16">
          <div class="common-form">{{ $t('销售额TOP10') }}</div>
          <div class="list ww100">
            <el-table v-if="salesMoneyRank.length > 0" :data="salesMoneyRank" style="width: 100%" size="small">
              <el-table-column prop="product_name_text" :label="$t('商品名称')">
                <template #default="scope">
                  <div class="product-name">
                    <span :class="scope.$index < 3 ? 'key-box' : 'key-box2'">{{ scope.$index + 1 }}</span>
                    <span class="">{{ scope.row.product_name_text }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="total_num" :label="$t('销量')"> </el-table-column>
              <el-table-column prop="total_price" :label="$t('销售额')">
                <template #default="scope">
                  {{ this.$formatPrice(scope.row.total_price) }}
                </template>
              </el-table-column>
            </el-table>
            <div v-else class="tc pt30">{{ $t('暂无上榜记录') }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import qs from 'qs';
  import StoreApi from '@/api/store.js';
  import datePick from '@/components/datePick/datePick.vue';
  import { languageStore } from '@/store/model/language.js';
  import { useUserStore } from '@/store';
  import dayjs from '@/utils/dayjs';
  const languageTag = languageStore().language;
  const { token } = useUserStore();
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_member = supplier.value?.is_open_member || 0;
  const delivery_status = supplier.value?.delivery_status || 0;
  export default {
    components: { datePick },
    data() {
      return {
        active: 0,
        /*是否加载完成*/
        loading: true,
        /*订单数据*/
        detail: {
          total_price: '',
          income_money: '',
          order_count: '',
          refund_money: '',
        },
        searchForm: {
          date: ['', ''],
        },
        activeName: 'sale',
        incomes: [],
        salesNumRank: [],
        salesMoneyRank: [],
        searchLoading: '',
        languageTag: languageTag,
        is_open_member: is_open_member,
        delivery_status: delivery_status,
        token,
        regionData: [],
      };
    },
    async mounted() {
      this.searchForm.date = [dayjs(), dayjs()];
      await this.$nextTick();
      /*获取列表*/
      this.getParams();
    },

    methods: {
      onChange(starDate, endDate) {
        this.searchForm.date[0] = starDate;
        this.searchForm.date[1] = endDate;
        this.onSearch();
      },

      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.getParams();
        }, 200);
      },

      /*获取参数*/
      getParams() {
        let self = this;
        self.loading = true;
        let params = self.searchForm;
        StoreApi.storeSurvey(params, true)
          .then((data) => {
            self.detail = data.data.detail;
            self.salesNumRank = data.data.salesNumRank;
            self.salesMoneyRank = data.data.salesMoneyRank;
            self.incomes = data.data.detail.incomes;
            self.regionData = data.data.regionData;
            self.loading = false;
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      timeLimit() {
        if (this.searchForm.date[0] && this.searchForm.date[1]) {
          //不能超过31天
          const startDate = new Date(this.searchForm.date[0]);
          const endDate = new Date(this.searchForm.date[1]);
          const diffTime = endDate.getTime() - startDate.getTime();
          const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
          if (diffDays > 30) {
            return true;
          }
        }
      },

      onExport: function () {
        if (this.searchForm.date[0] && this.searchForm.date[1]) {
          if (this.timeLimit()) {
            this.$ElMessage({
              message: this.$t('导出时间范围不能超过31天'),
              type: 'warning',
            });
            return;
          }
          let baseUrl = window.location.protocol + '//' + window.location.host;
          this.searchForm.token = this.token;
          window.location.href = baseUrl + '/index.php/shop/store.survey/export?' + qs.stringify(this.searchForm) + '&language=' + this.languageTag;
        } else {
          this.$ElMessage({
            message: $t('请选择时间'),
            type: 'warning',
          });
        }
      },
    },
  };
</script>
<style lang="scss" scoped>
  .el-row {
    margin-bottom: 20px;

    &:last-child {
      margin-bottom: 0;
    }
  }

  .el-col {
    border-radius: 4px;
  }

  .grid-content {
    padding: 20px;
    border-radius: 4px;
    min-height: 36px;
  }

  .bg-purple {
    background: #f4f4f4;
  }

  .table-wrap {
    padding: 20px;
    padding-top: 0;
  }

  .common-form-data {
    margin-left: 15px;
    font-weight: 500;
  }

  .tips {
    padding: 15px;
    margin-bottom: 20px;
  }

  .tips_tit {
    font-size: 22px;
    margin-bottom: 10px;
  }

  .tips_txt {
    line-height: 25px;
    color: #666;
    font-size: 14px;
  }

  .detail_prici {
    font-size: 20px;
    color: #000;
    font-weight: bold;
    margin-top: 10px;
    max-width: 250px;
  }

  .order-form {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 16px;
  }

  .order-box {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    flex: 1;
  }

  .order-box-l {
    display: flex;
    flex-direction: column;
    flex: 1;
    padding: 16px;
    background: #fff6de;
    border-radius: 4px;
  }

  .box-l-title {
    display: flex;
  }

  .box-l-title .title-box {
    display: flex;
    flex-direction: column;
    flex: 1;
  }

  .box-l-title .title-box h3 {
    font-size: 14px;
    font-weight: 400;
    color: var(--el-color-black);
    margin-bottom: 6px;
  }

  .box-l-title .title-box h4 {
    font-size: 20px;
    font-weight: 700;
    color: var(--el-color-black);
    margin-top: auto;
  }

  .box-main {
    display: flex;
    flex-direction: column;
    flex: 1;
    padding: 16px;
    background: #fff;
    border-radius: 4px;
    margin-top: 16px;
  }

  .box-main h4 {
    font-size: 16px;
    font-weight: 700;
    color: var(--el-color-black);
  }

  .main-body {
    display: flex;
    margin-top: 12px;
  }

  .main-div {
    display: flex;
    flex-direction: column;
    flex: 1;
  }

  .main-body h3 {
    font-size: 14px;
    font-weight: 400;
    color: var(--el-color-black);
    margin-bottom: 6px;
  }

  .main-body h4 {
    font-size: 20px;
    font-weight: 700;
    color: var(--el-color-black);
    margin-top: auto;
  }

  .common-form {
    flex-shrink: 0;
  }

  .pay-data {
    display: flex;
    flex-direction: column;
    border-radius: 4px;
    flex: 1 1 auto;
    overflow-y: auto;
    max-height: 500px;
    background: #fff6de;
  }

  .product-bottom {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 16px;
  }

  .operation-data {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 16px;
  }

  .data-box {
    display: flex;
    flex-direction: column;
    flex: 1;
    padding: 16px;
    background: #fff6de;
    border-radius: 4px;
    min-width: calc(25% - 12px);
    max-width: calc(25% - 8px);
  }

  .data-box-title {
    display: flex;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .data-box-title h3 {
    font-size: 16px;
    font-weight: 400;
    color: var(--el-color-black);
  }

  .data-box h4 {
    font-size: 20px;
    font-weight: 700;
    margin-top: auto;
    color: var(--el-color-black);
  }

  .data-box h5 {
    color: var(--el-color-tips);
    font-size: 12px;
    font-style: normal;
    font-weight: 400;
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

  .area-data {
    margin-bottom: 16px;

    .area-data-item-content {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 12px;

      .area-data-item-content-item {
        border-radius: 4px;
        background: #fff6de;
        padding: 12px;
        display: flex;

        .area-data-item-content-item-title {
          flex-shrink: 0;
          min-width: 80px;
          height: 80px;
          background: var(--el-color-primary);
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 18px;
          border-radius: 4px;
        }

        .area-data-item-content-item-content {
          flex-grow: 1;
          padding-left: 12px;
          display: flex;
          flex-direction: column;
          justify-content: space-between;

          .area-data-item-content-item-content-wrap {
            display: flex;
            align-items: baseline;
            justify-content: space-between;
            font-size: 14px;
            font-weight: bold;
          }
        }
      }
    }
  }
</style>

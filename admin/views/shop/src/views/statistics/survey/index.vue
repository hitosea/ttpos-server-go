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
              type="datetimerange"
              format="YYYY-MM-DD HH:mm"
              value-format="YYYY-MM-DD HH:mm"
              time-format="HH:mm"
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
      <dataBox
        :title="$t('总销售额')"
        :content="
          $t('总销售额：') +
          `<br />` +
          $t('商品已含税（原商品金额+实收服务费+实收服务税费+实收支付手续费）') +
          `<br />` +
          $t('商品未含税（原商品金额+实收服务费+实收商品及服务税费+实收支付手续费）') +
          `<br />` +
          $t('（以上计算都不包含会员充值金额）')
        "
        :value="this.$formatPrice(detail.receivable_price || 0)"
      ></dataBox>

      <dataBox
        :title="$t('实收金额')"
        :content="$t('指订单原价扣除优惠折扣、会员折扣后的金额，包含支付手续费+收银机会员充值金额（不包括会员余额消费金额）')"
        :value="this.$formatPrice(detail.received_price || 0)"
      ></dataBox>

      <dataBox :title="$t('商品数量')" :content="$t('所售卖的商品数量，仅计算普通商品及自助餐（不包括自助餐加钟）')" :value="detail.product_num || 0"></dataBox>

      <dataBox
        v-if="is_open_member == '1'"
        :title="$t('新增会员数/会员折扣')"
        :content="$t('新增会员数：新增会员数量') + `<br />` + $t('会员折扣：会员折扣总金额（包含等级折扣及会员卡折扣）')"
        :value="(detail.user_count || 0) + '/' + (detail.user_discount_money || 0)"
      ></dataBox>

      <dataBox :title="$t('营业收入')" :content="$t('营业收入=实收金额-税费+会员余额消费金额')" :value="this.$formatPrice(detail.business_price || 0)"></dataBox>

      <dataBox
        :title="$t('服务费') + '/' + $t('支付手续费')"
        :content="$t('服务费：退款后将减少退款单所含的服务费') + `<br />` + $t('支付手续费：使用需另付手续费的支付方式，所增加的手续费，整单退款后将减少退款单所含的支付手续费')"
        :value="this.$formatPrice(detail.service_money || 0) + '/' + this.$formatPrice(detail.pay_fee_money || 0)"
      ></dataBox>

      <dataBox :title="$t('税费')" :content="$t('税费，包含商品税费、服务费税费')" :value="this.$formatPrice(detail.consumption_tax_money || 0)"></dataBox>

      <dataBox :title="$t('退款金额')" :content="$t('订单退款的金额（反结账不计入在此）')" :value="this.$formatPrice(detail.refund_money || 0)"></dataBox>

      <dataBox
        :title="$t('优惠折扣/优惠占比')"
        :content="$t('优惠折扣=优惠折扣总金额（包含改价、折扣比例、抹零、结账抹零）') + `<br />` + $t('优惠占比=优惠折扣/总销售额')"
        :value="(detail.discount_money || 0) + '/' + (detail.discount_ratio || 0)"
      ></dataBox>

      <dataBox
        :title="$t('赠菜总额/赠菜数量')"
        :content="$t('赠菜总额：赠菜的总金额') + `<br />` + $t('赠菜数量：赠菜的数量')"
        :value="(detail.free_product_price || 0) + '/' + (detail.free_product_num || 0)"
      ></dataBox>

      <dataBox
        :title="$t('免单总额/免单数量')"
        :content="$t('免单总额：免单的总金额') + `<br />` + $t('免单数量：免单的数量')"
        :value="(detail.free_order_price || 0) + '/' + (detail.free_order_num || 0)"
      ></dataBox>

      <dataBox :title="$t('充值金额')" :content="$t('会员充值金额（实际充值的金额，不包含赠送金额）')" :value="this.$formatPrice(detail.recharge_amount || 0)"></dataBox>

      <dataBox
        v-if="delivery_status == '1'"
        :title="$t('外送销售')"
        :content="$t('顾客通过在线外卖下单的订单总额')"
        :value="this.$formatPrice(detail.delivery_order_amount || 0)"
      ></dataBox>

      <dataBox
        v-if="delivery_status == '1'"
        :title="$t('外送营收')"
        :content="$t('顾客通过在线外卖下单的订单总额-外卖订单退款总额-配送费')"
        :value="this.$formatPrice(detail.delivery_order_revenue || 0)"
      ></dataBox>

      <dataBox v-if="delivery_status == '1'" :title="$t('外送退款')" :value="this.$formatPrice(detail.delivery_order_refund_amount || 0)"></dataBox>

      <dataBox v-if="delivery_status == '1'" :title="$t('外送配送费')" :value="this.$formatPrice(detail.delivery_fee || 0)"></dataBox>
    </div>

    <!--区域数据-->
    <areaData :regionData="regionData"></areaData>

    <!--订单数据-->
    <orderForm :detail="detail" :incomes="incomes" :delivery_status="delivery_status"></orderForm>

    <!--销量-->
    <salesVolume :salesNumRank="salesNumRank" :salesMoneyRank="salesMoneyRank" />
  </div>
</template>

<script>
  import qs from 'qs';
  import StoreApi from '@/api/store.js';
  import datePick from '@/components/datePick/datePick.vue';
  import { languageStore } from '@/store/model/language.js';
  import { useUserStore } from '@/store';
  import dayjs from '@/utils/dayjs';
  import dataBox from '@/components/dataBox/dataBox.vue';
  import areaData from './part/areaData.vue';
  import salesVolume from './part/salesVolume.vue';
  import orderForm from './part/orderForm.vue';
  const languageTag = languageStore().language;
  const { token } = useUserStore();
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_member = supplier.value?.is_open_member || 0;
  const delivery_status = supplier.value?.delivery_status || 0;
  export default {
    components: { datePick, dataBox, areaData, salesVolume, orderForm },
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

  .operation-data {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 16px;
  }

  .data-box {
    min-width: calc(25% - 12px);
    max-width: calc(25% - 8px);
  }
</style>

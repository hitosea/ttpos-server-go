<template>
  <el-dialog :title="$t('收银交班详情')" v-model="dialogVisible" @close="dialogFormVisible" width="90%" :close-on-click-modal="false" :close-on-press-escape="false">
    <div class="common-form">{{ $t('基本信息') }}</div>
    <el-row :gutter="16">
      <el-col :span="8">
        <p class="text">
          {{ $t('当班编号') }}:<span>{{ detail.shift_no }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('收银员') }}:<span>{{ detail.user?.real_name }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('当班时间') }}:<span>{{ detail.shift_start_time }}</span> {{ $t('至') }}<span>{{ detail.shift_end_time }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('总销售额') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.total_business || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('服务费') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.order?.service_money || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('支付手续费') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.order?.pay_fee_money || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('税费') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.order?.consumption_tax_money || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('商品数量') }}:
          <span>
            {{ this.$formatPrice(detail.order?.product_num || 0) }}
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('优惠折扣') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.order?.discount_money || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8" v-if="is_open_member == '1'">
        <p class="text">
          {{ $t('会员折扣') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.order?.user_discount_money || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('退款') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.order?.refund_money || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('营业收入') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.total_income || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('实收金额') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.order?.received_price || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <template v-if="detail.order?.percentage_list || [].length > 0">
        <el-col :span="8" v-for="item in detail.order?.percentage_list">
          <p class="text">
            <template v-if="languageTag === 'ja'"> {{ item.tax_rate }}%対象: </template>
            <template v-else> VAT ({{ item.tax_rate }}%): </template>
            <span>
              <main-currency>
                {{ this.$formatPrice(item.total_price || 0) }}
              </main-currency>
            </span>
            <span>
              (
              <template v-if="languageTag === 'ja'"> 内消費税 </template>
              <template v-else>
                {{ $t('其中VAT') }}
              </template>
              <main-currency>
                {{ this.$formatPrice(item.consumption_tax || 0) }}
              </main-currency>
              )
            </span>
          </p>
        </el-col>
      </template>

      <template v-for="item in detail.incomes">
        <el-col :span="8">
          <p class="text">
            {{ item.pay_type_name }}:<span>
              <main-currency>
                {{ this.$formatPrice(item.price || 0) }}
              </main-currency>
            </span>
          </p>
        </el-col>
      </template>

      <el-col :span="8">
        <p class="text">
          {{ $t('充值金额') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail?.order?.recharge_amount || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('赠送金额') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail?.order?.gift_money || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('赠菜金额') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail?.order?.total_give_product_price || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('赠送积分') }}:
          <span>
            <main-currency>
              {{ detail?.order?.gift_points || 0 }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('上一班遗留备用金') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.previous_shift_cash || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('本班取出现金') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.cash_taken_out || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('本班遗留备用金') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.cash_left || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('中途存入现金') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.deposit_cash || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('中途取出现金') }}:
          <span>
            <main-currency>
              {{ this.$formatPrice(detail.withdraw_cash || 0) }}
            </main-currency>
          </span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('异常报备') }}:
          <span>
            {{ detail.exception_remark || '-' }}
          </span>
        </p>
      </el-col>
    </el-row>

    <div class="common-form">{{ $t('异常信息') }}</div>
    <el-row :gutter="16">
      <el-col :span="8">
        <p class="text">
          {{ $t('退菜次数') }}:<span>{{ abnormal?.refund_product_times ?? '-' }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('退款次数') }}:<span>{{ abnormal?.refund_times ?? '-' }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('反结账次数') }}:<span>{{ abnormal?.reverse_settle_times ?? '-' }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('赠菜次数') }}:<span>{{ abnormal?.product_free_times ?? '-' }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('免单次数') }}:<span>{{ abnormal?.free_order_times ?? '-' }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('转菜次数') }}:<span>{{ abnormal?.product_move_times ?? '-' }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('单品改价次数') }}:<span>{{ abnormal?.change_price_times ?? '-' }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('整单改价次数') }}:<span>{{ abnormal?.change_order_price_times ?? '-' }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('整单折扣次数') }}:<span>{{ abnormal?.discount_order_times ?? '-' }}</span>
        </p>
      </el-col>
      <el-col :span="8">
        <p class="text">
          {{ $t('整单抹零次数') }}:<span>{{ abnormal?.round_order_times ?? '-' }}</span>
        </p>
      </el-col>
    </el-row>

    <div class="common-form">{{ $t('合计') }}</div>
    <el-table size="small" :data="orderData" border style="width: 100%; margin-bottom: 16px" v-loading="loading">
      <el-table-column prop="total_order_num" :label="$t('所有订单数')"></el-table-column>
      <el-table-column prop="total_cancel_order_num" :label="$t('取消订单数')">
        <template #default="scope">
          {{ scope.row.total_cancel_order_num || 0 }}
        </template>
      </el-table-column>
      <el-table-column prop="total_table_num" :label="$t('桌数')"></el-table-column>
      <el-table-column prop="total_people_num" :label="$t('人数')"></el-table-column>
      <el-table-column :label="$t('最小/最大订单金额')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.min_order_price) }}
          </main-currency>
          /
          <main-currency>
            {{ this.$formatPrice(scope.row.max_order_price) }}
          </main-currency>
        </template>
      </el-table-column>
      <el-table-column prop="total_cancel_order_amount" :label="$t('取消订单金额')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.total_cancel_order_amount || 0) }}
          </main-currency>
        </template>
      </el-table-column>
      <el-table-column prop="avg_order_price" :label="$t('平均订单金额')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.avg_order_price) }}
          </main-currency>
        </template>
      </el-table-column>
    </el-table>

    <div class="common-form">{{ $t('桌台方式') }}</div>
    <el-table size="small" :data="zhuoData" border style="width: 100%; margin-bottom: 16px" v-loading="loading">
      <el-table-column prop="table_order_num" :label="$t('订单数（桌数）')"></el-table-column>
      <el-table-column prop="table_people_num" :label="$t('人数')"></el-table-column>
      <el-table-column :label="$t('最小/最大订单金额')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.table_min_order_price || 0) }}
          </main-currency>
          /
          <main-currency>
            {{ this.$formatPrice(scope.row.table_max_order_price || 0) }}
          </main-currency>
        </template>
      </el-table-column>
      <el-table-column prop="table_avg_order_price" :label="$t('平均订单金额')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.table_avg_order_price || 0) }}
          </main-currency>
        </template>
      </el-table-column>
      <el-table-column prop="table_people_avg" :label="$t('人均')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.table_people_avg || 0) }}
          </main-currency>
        </template>
      </el-table-column>
    </el-table>

    <div class="common-form">{{ $t('点餐方式') }}</div>
    <el-table size="small" :data="cashierData" border style="width: 100%; margin-bottom: 16px" v-loading="loading">
      <el-table-column prop="cashier_order_num" :label="$t('订单数')"></el-table-column>
      <el-table-column :label="$t('最小/最大订单金额')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.cashier_min_order_price || 0) }}
          </main-currency>
          /
          <main-currency>
            {{ this.$formatPrice(scope.row.cashier_max_order_price || 0) }}
          </main-currency>
        </template>
      </el-table-column>
      <el-table-column prop="cashier_avg_order_price" :label="$t('平均订单金额')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.cashier_avg_order_price || 0) }}
          </main-currency>
        </template>
      </el-table-column>
    </el-table>

    <div class="common-form">{{ $t('其他') }}</div>
    <el-table size="small" :data="detail.order?.peak_hour_list" border style="width: 100%; margin-bottom: 16px" v-loading="loading">
      <el-table-column prop="time_period" :label="$t('高峰时间')"></el-table-column>
      <el-table-column prop="num" :label="$t('订单数')"></el-table-column>

      <el-table-column prop="amount" :label="$t('订单金额')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.amount || 0) }}
          </main-currency>
        </template>
      </el-table-column>
    </el-table>

    <div class="common-form">{{ $t('销售信息') }}</div>
    <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
      <el-table-column prop="name_text" :label="$t('分类')"></el-table-column>
      <el-table-column prop="sales" :label="$t('销售数量')"></el-table-column>
      <el-table-column prop="prices" :label="$t('销售金额')">
        <template #default="scope">
          <main-currency>
            {{ this.$formatPrice(scope.row.prices || 0) }}
          </main-currency>
        </template>
      </el-table-column>
    </el-table>

    <template #footer>
      <div class="dialog-footer">
        <el-button type="primary" @click="dialogFormVisible" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>
<script>
  import StatisticsApi from '@/api/statistics.js';
  import { useUserStore } from '@/store';
  import { languageStore } from '@/store/model/language.js';

  const { currency } = useUserStore();
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_member = supplier.value?.is_open_member || 0;

  const language = languageStore();
  const languageTag = languageStore().language;

  export default {
    props: ['open', 'detailId'],

    created() {
      this.dialogVisible = this.open;
      this.id = this.detailId;
      this.getData();
    },

    data() {
      return {
        languageTag,
        dialogVisible: false,
        loading: false,
        id: {},
        detail: {},
        abnormal: {},
        tableData: [],
        orderData: [],
        zhuoData: [],
        cashierData: [],
        currency: currency,
        is_open_member: is_open_member,
      };
    },

    methods: {
      /*获取数据*/
      getData() {
        let self = this;
        let params = {};
        params.id = self.id;
        StatisticsApi.getUserShiftLogdDetail(params, true)
          .then((res) => {
            self.detail = res.data.detail;
            self.abnormal = res.data.detail.abnormal;
            self.tableData = res.data.salesInfo;
            self.orderData.push({
              total_order_num: res.data.detail.order.total_order_num,
              total_table_num: res.data.detail.order.total_table_num,
              total_people_num: res.data.detail.order.total_people_num,
              min_order_price: res.data.detail.order.min_order_price,
              max_order_price: res.data.detail.order.max_order_price,
              avg_order_price: res.data.detail.order.avg_order_price,
            });

            self.zhuoData.push({
              table_order_num: res.data.detail.order.table_order_num,
              table_people_num: res.data.detail.order.table_people_num,
              table_min_order_price: res.data.detail.order.table_min_order_price,
              table_max_order_price: res.data.detail.order.table_max_order_price,
              table_avg_order_price: res.data.detail.order.table_avg_order_price,
              table_people_avg: res.data.detail.order.table_people_avg,
            });

            self.cashierData.push({
              cashier_order_num: res.data.detail.order.cashier_order_num,
              cashier_min_order_price: res.data.detail.order.cashier_min_order_price,
              cashier_max_order_price: res.data.detail.order.cashier_max_order_price,
              cashier_avg_order_price: res.data.detail.order.cashier_avg_order_price,
            });
          })
          .catch((error) => {});
      },
      dialogFormVisible() {
        this.$emit('closeDialog');
      },
    },
  };
</script>
<style scoped>
  .text {
    color: var(--el-color-tips);
    font-size: 14px;
    margin-bottom: 16px;
  }

  .text span {
    color: var(--el-color-black);
    font-weight: 500;
    margin-left: 8px;
  }
</style>

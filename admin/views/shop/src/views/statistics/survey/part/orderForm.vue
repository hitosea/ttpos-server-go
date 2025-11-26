<template>
  <div class="order-form flex">
    <div class="order-box">
      <div class="common-form">{{ $t('订单数据') }}</div>
      <div class="order-box-l">
        <div class="box-l-title">
          <div class="title-box">
            <h3>{{ $t('合计订单数') }}</h3>
            <h4>{{ proxy.$formatPrice(detail.total_order_num || 0) }}</h4>
          </div>
          <div class="title-box">
            <h3>{{ $t('最小订单金额') }}</h3>
            <h4>{{ proxy.$formatPrice(detail.min_order_price || 0) }}</h4>
          </div>
          <div class="title-box">
            <h3>{{ $t('最大订单金额') }}</h3>
            <h4>{{ proxy.$formatPrice(detail.max_order_price || 0) }}</h4>
          </div>
          <div class="title-box">
            <h3>{{ $t('平均订单金额') }}</h3>
            <h4>{{ proxy.$formatPrice(detail.avg_order_price || 0) }}</h4>
          </div>
        </div>
        <div class="box-main">
          <h4>{{ $t('桌台方式') }}</h4>
          <div class="main-body">
            <div class="main-div">
              <h3>{{ $t('桌数') }}</h3>
              <h4>{{ proxy.$formatPrice(detail.table_order_num || 0) }}</h4>
            </div>
            <div class="main-div">
              <h3>{{ $t('人数') }}</h3>
              <h4>{{ proxy.$formatPrice(detail.table_people_num || 0) }}</h4>
            </div>
            <div class="main-div">
              <h3>{{ $t('最小/大订单金额') }}</h3>
              <h4>
                {{ proxy.$formatPrice(detail.table_min_order_price || 0) + '/' + proxy.$formatPrice(detail.table_max_order_price || 0) }}
              </h4>
            </div>
            <div class="main-div">
              <h3>{{ $t('平均订单金额') }}</h3>
              <h4>
                {{ proxy.$formatPrice(detail.table_avg_order_price || 0) }}
              </h4>
            </div>
          </div>
        </div>
        <div class="box-main">
          <h4>{{ $t('点餐方式') }}-{{ $t('店内') }}</h4>
          <div class="main-body">
            <div class="main-div">
              <h3>{{ $t('订单数') }}</h3>
              <h4>{{ proxy.$formatPrice(detail.cashier_order_num || 0) }}</h4>
            </div>

            <div class="main-div">
              <h3>{{ $t('最小/大订单金额') }}</h3>
              <h4>
                {{ proxy.$formatPrice(detail.cashier_min_order_price || 0) + '/' + proxy.$formatPrice(detail.cashier_max_order_price || 0) }}
              </h4>
            </div>
            <div class="main-div">
              <h3>{{ $t('平均订单金额') }}</h3>
              <h4>
                {{ proxy.$formatPrice(detail.cashier_avg_order_price || 0) }}
              </h4>
            </div>
          </div>
        </div>

        <div class="box-main">
          <h4>{{ $t('点餐方式') }}-{{ $t('外卖') }}</h4>
          <div class="main-body">
            <div class="main-div">
              <h3>{{ $t('订单数') }}</h3>
              <h4>{{ proxy.$formatPrice(detail.takeaway_order_num || 0) }}</h4>
            </div>

            <div class="main-div">
              <h3>{{ $t('最小/大订单金额') }}</h3>
              <h4>
                {{ proxy.$formatPrice(detail.takeaway_min_order_price || 0) + '/' + proxy.$formatPrice(detail.takeaway_max_order_price || 0) }}
              </h4>
            </div>
            <div class="main-div">
              <h3>{{ $t('平均订单金额') }}</h3>
              <h4>
                {{ proxy.$formatPrice(detail.takeaway_avg_order_price || 0) }}
              </h4>
            </div>
          </div>
        </div>

        <div class="box-main" v-if="delivery_status == 1">
          <h4>{{ $t('外送点餐') }}</h4>
          <div class="main-body">
            <div class="main-div">
              <h3>{{ $t('订单数') }}</h3>
              <h4>{{ proxy.$formatPrice(detail.delivery_order_num || 0) }}</h4>
            </div>

            <div class="main-div">
              <h3>{{ $t('最小/大订单金额') }}</h3>
              <h4>
                {{ proxy.$formatPrice(detail.delivery_min_order_price || 0) + '/' + proxy.$formatPrice(detail.delivery_max_order_price || 0) }}
              </h4>
            </div>
            <div class="main-div">
              <h3>{{ $t('平均订单金额') }}</h3>
              <h4>
                {{ proxy.$formatPrice(detail.delivery_avg_order_price || 0) }}
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
              {{ proxy.$formatPrice(scope.row.price) }}
            </template>
          </el-table-column>
        </el-table>
        <div v-else class="tc p30">{{ $t('暂无数据') }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { getCurrentInstance } from 'vue';
  const { proxy } = getCurrentInstance();
  const props = defineProps({
    detail: {
      type: Object,
      default: () => {},
    },
    incomes: {
      type: Array,
      default: () => [],
    },
    delivery_status: {
      type: [String, Number],
      default: () => '',
    },
  });
</script>

<style lang="scss" scoped>
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
</style>

<template>
  <div class="home" v-loading="loading">
    <div class="operation-wrap" style="background-color: #ffffff">
      <h3 class="operation-title mb16">{{ $t('数据总览') }}</h3>
      <div class="operation-data">
        <dataBox
          :title="$t('实收金额')"
          :content="$t('实收金额：指订单原价扣除优惠折扣、会员折扣后的金额，包含支付手续费+会员充值金额（不包括会员余额消费金额）')"
          :value="$formatPrice(top_data.total_money || 0)"
        ></dataBox>

        <dataBox
          :title="$t('优惠折扣/会员折扣')"
          :content="$t('优惠折扣=优惠折扣总金额（包含改价、折扣比例、抹零、结账抹零）') + `<br />` + $t('会员折扣：会员折扣总金额（包含等级折扣及会员卡折扣）')"
          :value="$formatPrice(top_data.discount_money || 0) + '/' + $formatPrice(top_data.user_discount_money || 0)"
        ></dataBox>

        <dataBox v-if="is_open_member == '1'" :title="$t('会员数')" :content="$t('当前会员总数')" :value="$formatPrice(top_data.user_total || 0)"></dataBox>

        <dataBox :title="$t('订单数')" :content="$t('所有的订单数，包含待付款、已取消、已完成')" :value="$formatPrice(top_data.order_total || 0)"></dataBox>

        <dataBox :title="$t('退款金额')" :content="$t('订单退款的金额（反结账不计入在此）')" :value="$formatPrice(top_data.refund_money || 0)"></dataBox>
      </div>
    </div>
    <div class="operation-center mt12">
      <div class="operation-center-l">
        <h3 class="operation-title mb16">{{ $t('今日概况') }}</h3>
        <div class="operation-today">
          <gridContent
            :title="$t('实收金额')"
            :value="$formatPrice(today_data.order_total_price?.tday || 0)"
            :yesterdayData="$t('昨日：') + $formatPrice(today_data.order_total_price?.ytd || 0)"
          ></gridContent>

          <gridContent
            :title="$t('优惠折扣/会员折扣')"
            :value="$formatPrice(today_data.discount_money?.tday || 0) + '/' + $formatPrice(today_data.user_discount_money?.tday || 0)"
            :yesterdayData="$t('昨日：') + $formatPrice(today_data.discount_money?.ytd || 0) + '/' + $formatPrice(today_data.user_discount_money?.ytd || 0)"
          ></gridContent>

          <gridContent
            v-if="is_open_member == '1'"
            :title="$t('新增会员数')"
            :tooltipContent="$t('今日新增会员数量')"
            :value="today_data.new_user_total?.tday"
            :yesterdayData="$t('昨日：') + today_data.new_user_total?.ytd"
          ></gridContent>

          <gridContent
            :title="$t('订单数')"
            :value="$formatPrice(today_data.order_total?.tday || 0)"
            :yesterdayData="$t('昨日：') + $formatPrice(today_data.order_total?.ytd || 0)"
          ></gridContent>

          <gridContent
            :title="$t('退款金额')"
            :value="$formatPrice(today_data.order_refund_money?.tday || 0)"
            :yesterdayData="$t('昨日：') + $formatPrice(today_data.order_refund_money?.ytd || 0)"
          ></gridContent>
        </div>
      </div>
      <div class="operation-center-r">
        <h3 class="operation-title mb16">{{ $t('待办事项') }}</h3>

        <centerRBox :title="$t('库存告急')" :description="$t('查看')" :value="wait_data?.stock?.product || 0" @click="lockStock"></centerRBox>

        <centerRBox :title="$t('采购单')" :description="$t('查看')" :value="wait_data?.purchase?.apply || 0" @click="lookOrder"></centerRBox>
      </div>
    </div>

    <div class="home-index mt12">
      <!--main-index-->
      <div class="flex-1">
        <!--待办事项-->
        <div class="matters-wrap" style="width: 100%">
          <Transaction></Transaction>
        </div>
      </div>
      <div class="matters-wrap flex-1">
        <Ranking v-if="!loading" :product_data="product_data"></Ranking>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue';
  import { ElMessage } from 'element-plus';
  import IndexApi from '@/api/index.js';
  import Ranking from '@/views/home/part/product/Ranking.vue';
  import Transaction from '@/views/home/part/product/Transaction.vue';
  import { useUserStore } from '@/store';
  import dataBox from '@/components/dataBox/dataBox.vue';
  import gridContent from './part/product/gridContent.vue';
  import centerRBox from './part/product/centerRBox.vue';

  // 定义emits
  const emit = defineEmits(['selectMenu']);

  // 用户信息
  const { userInfo, computedRenderMenus, computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_member = ref(supplier.value?.is_open_member || 0);
  const renderMenus = computedRenderMenus().renderMenus;

  // 响应式数据
  const loading = ref(true);
  const top_data = ref({});
  const wait_data = ref({
    order: {},
    agent: {},
    supplier: {},
    activity: {},
    audit: {},
    stock: {},
  });
  const today_data = ref({
    order_total_price: {},
    order_total: {},
    new_user_total: {},
    new_supplier_total: {},
    apply_supplier_total: {},
    order_user_total: {},
    income_money: {},
  });
  const product_data = ref({
    salesMoneyRank: [],
    salesNumRank: [],
  });
  const user_type = ref('');
  const menuList = ref(renderMenus);

  // 获取基础信息
  const getBaseInfo = async () => {
    user_type.value = userInfo.user_type;
  };

  // 获取统计数据
  const getData = async () => {
    loading.value = true;
    try {
      const data = await IndexApi.getCount(true);
      if (data.code == -102) return;    
      Object.assign(product_data.value, data.data.data.product_data);
      top_data.value = data.data.data.top_data;
      wait_data.value = data.data.data.wait_data;
      today_data.value = data.data.data.today_data;
    } catch (error) {
      ElMessage.error(error?.message || $t('获取数据失败'));
    } finally {
      loading.value = false;
    }
  };

  // 跳转库存告警
  const lockStock = () => {
    menuList.value.map((item, index) => {
      if (item.name == '商品管理') {
        emit('selectMenu', {
          type: 2,
          item: item,
          index: index,
          query: { inventory: 10 },
        });
      }
    });
  };

  // 跳转采购单
  const lookOrder = () => {
    menuList.value.map((item, index) => {
      if (item.name == '采购管理') {
        emit('selectMenu', {
          type: 2,
          item: item,
          index: index,
        });
      }
    });
  };

  // 组件挂载时初始化数据
  onMounted(() => {
    getData();
    getBaseInfo();
  });
</script>

<style lang="scss" scoped>
  .operation-wrap {
    height: auto;
    width: 100%;
    padding: 16px;
    border-radius: 8px;
    background-size: 100% 100%;
    color: #ffffff;
  }

  .operation-center {
    width: 100%;
    display: flex;
    gap: 12px;
  }

  .operation-center-l {
    flex: 7;
    border-radius: 4px;
    padding: 16px;
    background: #fff;
  }

  .operation-today {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .operation-center-r {
    flex: 3;
    border-radius: 4px;
    padding: 16px;
    background: #fff;
  }

  .operation-title {
    color: var(--el-color-black);
    font-size: 20px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .operation-data {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .home-index {
    width: 100%;
    display: -ms-flexbox;
    display: flex;
    gap: 12px;
    -webkit-box-orient: horizontal;
    -webkit-box-direction: normal;
    -ms-flex-direction: row;
    flex-direction: row;
    -webkit-box-pack: justify;
    -ms-flex-pack: justify;
    justify-content: space-between;
    min-width: 1000px;
    overflow-x: auto;
  }

  .main-index {
    flex: 1;
    margin: 20px;
  }

  .matters-wrap {
    padding-bottom: 15px;
    padding: 16px;
    border-radius: 4px;
    background: #fff;
  }

  .matters .box {
    width: 100%;
  }

  .matters-wrap .matters {
    display: -ms-flexbox;
    display: flex;
    -webkit-box-direction: column;
    -ms-flex-direction: column;
    flex-direction: column;
    justify-content: flex-start;
    align-items: flex-start;
    // height: 120px;
    margin-bottom: 30px;
  }

  .matters-wrap .matters .title {
    font-size: 16px;
    color: #333333;
    display: inline-block;
    height: 20px;
    line-height: 0;
    padding: 11px;
    text-align: center;
    margin-bottom: 20px;
  }

  .matters-wrap .matters ul {
    color: #999999;
  }

  .matters-wrap .matters ul span {
    padding-right: 6px;
    color: #3a8ee6;
  }

  .border-b {
    display: flex;
    flex-direction: column;
  }

  .border-b-l {
    flex-direction: initial;
  }

  .matters_item {
    display: flex;
    justify-content: flex-start;
    align-items: center;
  }

  .matters_item li {
    width: 72px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    margin-right: 16px;
  }

  .matters_box {
    width: 90%;
    border-top: 1px solid #d9d9d9;
    padding-top: 20px;
  }

  .matters-wrap .matters_item li .fb {
    font-size: 16px;
    color: #5d75e3;
  }
</style>

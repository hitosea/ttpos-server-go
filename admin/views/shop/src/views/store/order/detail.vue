<template>
  <div class="pb50" v-loading="loading">
    <div class="product-content">
      <!--基本信息-->
      <div class="common-form">{{ $t('基本信息') }}</div>
      <div class="table-wrap">
        <el-row>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单类型：') }}</span>
              {{ detail.bill_type == 1 ? $t('点餐订单') : $t('桌台订单') }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单号：') }}</span>
              {{ detail.order_no }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.member_uuids">
            <div class="pb16">
              <span class="gray9">{{ $t('会员：') }}</span>
              <span>{{ $t('会员ID') }}&nbsp;({{ detail?.member_uuids }})</span>
            </div>
          </el-col>

          <el-col :span="6" v-if="detail.nationality_uuid">
            <div class="pb16">
              <span class="gray9">{{ $t('国籍：') }}</span>
              <span>{{ detail?.nationality_name !== '' ? detail?.nationality_name : '-' }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单金额：') }}</span>
              <main-currency>
                {{ $formatPrice(detail.order_amount) }}
              </main-currency>
              <span v-if="currency.is_open == 1" style="padding-left: 8px">
                <sub-currency>
                  {{ $formatPrice(Number(detail.order_amount) * Number(currency.vices?.unit_rate)) }}
                </sub-currency>
              </span>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.status == 1">
            <div class="pb16">
              <span class="gray9">{{ $t('实付款金额：') }}</span>
              <main-currency>
                {{ $formatPrice(Number(detail.payment_amount)) }}
              </main-currency>
            </div>
          </el-col>
          <el-col :span="6" v-if="Number(detail.refund_money || 0) > 0">
            <div class="pb16">
              <span class="gray9">{{ $t('退款金额：') }}</span>
              <span>
                <main-currency>
                  {{ $formatPrice(detail.refund_amount) }}
                </main-currency>
              </span>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.status == 1" v-for="item in detail.pay_types">
            <div class="pb16">
              <span class="gray9">{{ $t('支付方式：') }}</span>
              <span>
                {{ item.payment_type_name }}
                <template v-if="item.source_text && item.code != 40 && item.code != 10 && item.code != -1"> ({{ item.source_text }}) </template>
                <el-tag class="ml-8" v-if="item.status == 0" type="danger" size="large">{{ $t('异常') }}</el-tag>
              </span>
            </div>
          </el-col>
          <el-col :span="6" >
            <div class="pb16">
              <span class="gray9">{{ $t('来源：') }}</span> {{ detail.order_source_name !== '' ? detail.order_source_name : $t('店内') }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('用餐方式：') }}</span> {{ detail.dining_method == 0 ? $t('店内就餐') : $t('打包带走') }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.serial_no">
            <div class="pb16">
              <span class="gray9">{{ detail.bill_type == 1 ? $t('序号：') : $t('桌号：') }}</span> {{ detail.serial_no }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('交易状态：') }}</span>
              {{ detail.status == 0 ? $t('待付款') : detail.status == 2 ? $t('已取消') : $t('已完成') }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.cashier_name">
            <div class="pb16">
              <span class="gray9">{{ $t('收银员：') }}</span>
              {{ detail.cashier_name || '-' }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.is_buffet == 1">
            <div class="pb16">
              <span class="gray9">{{ $t('自助餐：') }}</span
              >{{ detail.buffet_names }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.status == 2">
            <div class="pb16">
              <span class="gray9">{{ $t('取消原因：') }}</span>
              {{ detail.cancel_reason || '-' }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.sale_orders.length == 1 && detail.sale_orders[0]?.is_free">
            <div class="pb16">
              <span class="gray9">{{ $t('免单原因：') }}</span>
              {{ detail.sale_orders[0]?.free_reason[language] || '-' }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('时间：') }}</span>
              {{ detail.create_time || '-' }} {{ $t('至') }} {{ detail.finish_time || '-' }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.bill_type != 1">
            <div class="pb16">
              <span class="gray9">{{ $t('备注：') }}</span>
              {{ detail.remark || '-' }}
            </div>
          </el-col>
        </el-row>
      </div>

      <div class="common-form mt16"> {{ $t('商品信息') }} </div>

      <el-radio-group v-model="activeName" class="radio-search" v-if="(detail?.sale_orders || []).length > 1">
        <template v-for="(item, index) in detail?.sale_orders">
          <el-radio-button :label="index">{{ item.serial_no }}</el-radio-button>
        </template>
      </el-radio-group>
      <div class="sub-order" v-if="(detail?.sale_orders || []).length > 1">
        <p class="sub-order-item">{{ $t('订单号：') }}{{ detail?.sale_orders[activeName].order_no }}</p>
        <p class="sub-order-item">
          {{ $t('订单金额：') }}
          <main-currency>
            {{ $formatPrice(detail?.sale_orders[activeName].order_amount) }}
          </main-currency>
          <span style="padding-left: 8px">
            <sub-currency>
              {{ $formatPrice(Number(detail?.sale_orders[activeName].order_amount) * Number(currency.vices?.unit_rate)) }}
            </sub-currency>
          </span>
        </p>
        <p class="sub-order-item">
          {{ $t('实付金额：') }}
          <main-currency>
            {{ $formatPrice(Number(detail?.sale_orders[activeName].payment_amount)) }}
          </main-currency>
        </p>
        <p class="sub-order-item">
          {{ $t('会员：') }}
          <span v-if="detail?.sale_orders[activeName].member_uuid">{{ $t('会员ID') }}({{ detail?.sale_orders[activeName].member_uuid }})</span>
          <span v-else>-</span>
        </p>
      </div>
      <div class="table-wrap">
        <el-table size="small" :data="detail?.sale_orders[activeName]?.products" border style="width: 100%">
          <el-table-column prop="name_text" :label="$t('商品')" width="400">
            <template #default="scope">
              <div class="product-info">
                <div class="pic">
                  <img v-if="scope.row.is_buffet_customer" src="../../../assets/img/buffet.png" />
                  <img v-else-if="scope.row.is_delay" src="../../../assets/img/clock.png" />
                  <img v-else v-img-url="scope.row.image_url" />
                </div>
                <div class="info">
                  <div class="name">
                    <el-tag class="mr-8" type="danger" effect="light" :hit="true" size="large" v-if="scope.row.refund_reason">{{ $t('退') }}</el-tag>
                    <el-tag class="mr-8" color="#FF3300" effect="dark" :hit="true" size="large" v-if="scope.row.is_gift">{{ $t('赠') }}</el-tag>
                    <el-tag class="mr-8" color="#19b800" effect="dark" :hit="true" size="large" v-if="scope.row.is_wrap">{{ $t('打包') }}</el-tag>
                    {{ scope.row.locale_name[language] || '-' }}
                  </div>
                  <div class="gray9" v-if="scope.row.locale_attribute_name[language]">
                    <autoTips :content="scope.row.locale_attribute_name[language]">{{ scope.row.locale_attribute_name[language] }}</autoTips>
                  </div>
                  <div class="price">
                    <span>
                      <main-currency>
                        {{ $formatPrice(scope.row.price) }}
                      </main-currency>
                      <span style="padding-left: 8px">
                        <sub-currency>
                          {{ $formatPrice(Number(scope.row.price) * Number(currency.vices?.unit_rate)) }}
                        </sub-currency>
                      </span>
                    </span>
                  </div>
                  <div>
                    <p class="gray9" v-if="scope.row.is_gift > 0">{{ $t('赠送原因：') }}{{ scope.row.gift_reason || '-' }}</p>
                    <p class="gray9" v-if="scope.row.refund_reason">{{ $t('退菜原因：') }}{{ scope.row.refund_reason || '-' }}</p>
                  </div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="num" :label="$t('购买数量')">
            <template #default="scope">
              <p>{{ scope.row.num }}</p>
            </template>
          </el-table-column>
          <el-table-column prop="product_price" :label="$t('商品总价')">
            <template #default="scope">
              <div>
                <template v-if="scope.row.total_price != scope.row.sale_price">
                  <span class="text-line-through" v-if="scope.row.sale_price">
                    <main-currency>
                      {{ $formatPrice(Number(scope.row.sale_price)) }}
                    </main-currency>
                  </span>
                  <span class="text-line-through" v-if="scope.row.sale_price">
                    <sub-currency>
                      {{ $formatPrice(Number(scope.row.sale_price) * Number(currency.vices?.unit_rate)) }}
                    </sub-currency>
                  </span>
                </template>

                <span>
                  <main-currency>
                    {{ $formatPrice(Number(scope.row.total_price)) }}
                  </main-currency>
                </span>

                <span>
                  <sub-currency>
                    {{ $formatPrice(Number(scope.row.total_price) * Number(currency.vices?.unit_rate)) }}
                  </sub-currency>
                </span>
                <span class="tips" v-if="Number(scope.row.refund_amount) > 0"> （{{ $t('退款：') + currency.unit + $formatPrice(scope.row.refund_amount) }}） </span>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="common-form mt16">{{ $t('操作记录') }}</div>

      <div class="timeline-box">
        <el-timeline class="timeline-main">
          <el-timeline-item v-for="(activity, index) in activities" placement="top" :key="index" :color="activity.color" :hollow="activity.hollow" :timestamp="activity.timestamp">
            <div class="flex">
              {{ activity.content }}
              <template v-if="(activity.pay_type || []).length > 0"
                >&nbsp; (<span class="flex" v-for="(item, itemIndex) in activity.pay_type" :key="itemIndex">
                  <span>{{ item.name }}:{{ item.unit + $formatPrice(item.refund_money) }}</span>
                  <el-tag v-if="item.refund_status == '0'" class="cupon" type="danger" @click="() => handleRetry(item)">{{ $t('退款失败，重试') }}</el-tag>
                  <el-tooltip v-if="item.value == '90333'" placement="bottom" trigger="click">
                    <el-icon class="icon"><WarningFilled /></el-icon>
                    <template #content>
                      <div> {{ $t('银行') }}{{ $t('：') }}{{ bankName(item.bank_code) }}</div>
                      <div> {{ $t('账户名') }}{{ $t('：') }}{{ item.account_no }} </div>
                      <div> {{ $t('账号') }}{{ $t('：') }}{{ item.account_name }} </div>
                    </template>
                  </el-tooltip>
                  <span v-if="itemIndex < (activity.pay_type || []).length - 1">、</span> </span
                >)
              </template>
            </div>
          </el-timeline-item>
        </el-timeline>
      </div>

      <div class="table-wrap" v-if="detail.status == 2 && detail.cancel_reason != ''">
        <div class="common-form mt16">{{ $t('取消信息') }}</div>
        <div class="table-wrap">
          <el-row>
            <el-col :span="6">
              <div class="pb16">
                <span class="gray9">{{ $t('商家备注') }}:</span>
                {{ detail.cancel_reason }}
              </div>
            </el-col>
          </el-row>
        </div>
      </div>
    </div>
    <div class="common-button-wrapper">
      <el-button size="small" @click="cancelFunc">{{ $t('返回') }}</el-button>
      <el-button v-if="extra?.is_cell_refund" @click="() => refundClick(detail)" type="danger" size="small" v-auth="'/store/operate/refund'">{{ $t('退款') }}</el-button>
      <el-button v-if="extra?.is_cell_cancel" @click="() => cancelClick(detail)" type="danger" size="small" v-auth="'/store/operate/order_cancel'">{{ $t('取消') }} </el-button>
      <el-button v-if="extra?.is_cell_delete" @click="() => delClick(detail)" type="danger" size="small" v-auth="'/store/order/delete'">{{ $t('删除') }} </el-button>
    </div>
    <!--处理-->
    <Cancel v-if="openEdit" :open_edit="openEdit" :order_no="orderNo" :order_id="saleBillUuid" @closeDialog="(e) => closeDialogFunc(e, 'edit')" />
    <refund
      v-if="openRefund"
      :open_edit="openRefund"
      :order_no="orderNo"
      :order_id="saleBillUuid"
      :sub_order_id="saleOrderUuid"
      :pay_price="payPrice"
      @closeDialog="(e) => closerefundDialogFunc(e, 'edit')"
    />
    <refundAgain v-if="openRefundAgain" :open_edit="openRefundAgain" :refundOrder="refundOrder" @closeDialog="(e) => closerefundAgainDialogFunc(e)" />
  </div>
</template>

<script setup>
  // 引入Vue3组合式API
  import { ref, reactive, onMounted } from 'vue';
  // 引入Element Plus弹窗和消息
  import { ElMessageBox, ElMessage } from 'element-plus';
  // 引入路由
  import { useRoute, useRouter } from 'vue-router';
  // 引入API
  import OrderApi from '@/api/order.js';
  // 引入组件
  import Cancel from './dialog/cancel.vue';
  import refund from './dialog/refund.vue';
  import refundAgain from './dialog/refundAgain.vue';
  // 引入store
  import { useUserStore } from '@/store';
  import { languageStore } from '@/store/model/language';
  import autoTips from '@/components/autoTips/autoTips.vue';
  // 获取路由实例
  const route = useRoute();
  const router = useRouter();

  // 获取store数据
  const { currency } = useUserStore();

  // 语言
  const language = ref('');
  // 是否加载完成
  const loading = ref(true);
  // 订单数据
  const detail = reactive({
    sale_bill_uuid: 0,
    sale_order_uuid: 0,
    pay_status: [],
    pay_type: [],
    delivery_type: [],
    user: {},
    address: [],
    product: [],
    order_status: [],
    extract: [],
    pay_types: [],
    sale_orders: [],
    supplier: {
      name: '',
    },
  });
  // 额外数据
  const extra = ref({});
  // 是否打开编辑弹窗
  const openEdit = ref(false);
  const openRefund = ref(false);
  const openRefundAgain = ref(false);
  // 当前编辑的对象
  const refundOrder = ref({});
  const orderNo = ref(0);
  const saleBillUuid = ref(0);
  const saleOrderUuid = ref(0);
  const payPrice = ref(0);
  const activeName = ref(0);
  const pageParams = ref({});
  const activities = ref([]);
  // 银行列表
  const bankList = ref([
    { name: 'BANGKOK BANK (BBL)', value: '002' },
    { name: 'KASIKORNBANK (KBANK)', value: '004' },
    { name: 'KRUNG THAI BANK (KTB)', value: '006' },
    { name: 'TMBTHANACHART BANK (TTB)', value: '011' },
    { name: 'SIAM COMMERCIAL BANK (SCB)', value: '014' },
    { name: 'CITIBANK BANGKOK BRANCH (CITI)', value: '017' },
    { name: 'SUMITOMO MITSUI BANK (SMBC)', value: '018' },
    { name: 'STANDARD CHARTERED BANK THAI (SCBT)', value: '020' },
    { name: 'CIMB THAI BANK (CIMBT)', value: '022' },
    { name: 'UNITED OVERSEAS BANK THAI (UOBT)', value: '024' },
    { name: 'BANK OF AYUDHYA (BAY)', value: '025' },
    { name: 'GOVERNMENT SAVINGS BANK (GSB)', value: '030' },
    { name: 'THE HONGKONG AND SHANGHAI BANKING CORPORATION (HSBC)', value: '031' },
    { name: 'GOVERNMENT HOUSING BANK (GHB)', value: '033' },
    { name: 'BANK FOR AGRICULTURE AND AGRICULTURAL COOPERATIVES (BAAC)', value: '034' },
    { name: 'MIZUHO CORPORATE BANK (MHCB)', value: '039' },
    { name: 'ISLAMIC BANK OF THAILAND (ISBT)', value: 'ISBT' },
    { name: 'TISCO BANK (TISCO)', value: 'TISCO' },
    { name: 'KIATNAKIN BANK (KK)', value: '069' },
    { name: 'INDUSTRIAL AND COMMERCIAL BANK OF CHINA (ICBC THAI)', value: '070' },
    { name: 'THAI CREDIT RETAIL BANK (TCRB)', value: '071' },
    { name: 'LAND AND HOUSES BANK (LH BANK)', value: '073' },
  ]);

  // 获取参数
  async function getParams() {
    loading.value = true;
    try {
      const res = await OrderApi.storeOrderdetail(
        {
          sale_bill_uuid: route.query.sale_bill_uuid,
          sale_order_uuid: route.query.sale_order_uuid,
        },
        true
      );

      Object.assign(detail, res.data.detail);
      extra.value = res.data.extra;
      activities.value = [];

      res.data.operation_log.list.map((item) => {
        activities.value.push({
          refund_type: item.refund_type,
          pay_type: item.pay_type,
          content: item.description,
          timestamp:
            item.create_time +
            ' ' +
            $t('操作人：') +
            (item.user_name ? item.user_name + (item.user_email ? `(${item.user_email})` : '') : item.user_email || '') +
            ' ' +
            item.source,
          color: '#0bbd87',
        });
      });
    } catch (error) {
      // 错误处理
    } finally {
      loading.value = false;
    }
  }

  // 取消返回
  function cancelFunc() {
    languageStore().setPageParams(pageParams.value);
    router.back(-1);
  }

  // 打开取消
  function cancelClick(item) {
    orderNo.value = item.order_no;
    saleBillUuid.value = item.sale_bill_uuid;
    openEdit.value = true;
  }

  // 打开退款
  function refundClick(item) {
    orderNo.value = item.order_no;
    saleBillUuid.value = item.sale_bill_uuid;
    saleOrderUuid.value = item.sale_orders[0].sale_order_uuid;
    payPrice.value = 0;
    openRefund.value = true;
  }

  // 关闭弹窗
  function closeDialogFunc(e, f) {
    if (f === 'edit') {
      openEdit.value = e.openDialog;
      if (e.type === 'success') {
        getParams();
      }
    }
  }

  // 关闭退款弹窗
  function closerefundDialogFunc(e, f) {
    if (f === 'edit') {
      openRefund.value = e.openDialog;
      if (e.type === 'success') {
        getParams();
      }
    }
  }

  // 关闭重试退款弹窗
  function closerefundAgainDialogFunc(e) {
    openRefundAgain.value = e.openDialog;
    if (e.type === 'success') {
      getParams();
    }
  }

  // 删除
  async function delClick(item) {
    try {
      await ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
        type: 'warning',
      });

      await OrderApi.storedelete({
        sale_bill_uuid: item.sale_bill_uuid,
        sale_order_uuid: item.is_split === 'undefined' ? item.sale_order_uuid : 0,
      });

      ElMessage({
        message: $t('删除成功'),
        type: 'success',
      });
      router.back(-1);
    } catch (error) {
      // 用户取消或出错
    }
  }

  // 处理重试
  async function handleRetry(item) {
    refundOrder.value = item;
    if (item.value === '90333') {
      openRefundAgain.value = true;
    } else {
      try {
        await ElMessageBox.confirm('确定重试退款操作?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        });

        loading.value = true;
        const params = {
          return_order_uuid: item.return_order_uuid,
          return_amount_uuid: item.return_amount_uuid,
        };

        await OrderApi.orderRefundAgain(params, true);

        ElMessage({
          message: $t('操作成功'),
          type: 'success',
        });
        getParams();
      } catch (error) {
        if (error !== 'cancel') {
          ElMessage({
            type: 'info',
            message: '已取消操作',
          });
        }
      } finally {
        loading.value = false;
      }
    }
  }

  // 获取银行名称
  function bankName(value) {
    let name = '';
    bankList.value.map((item) => {
      if (item.value === value) {
        name = item.name;
      }
    });
    return name;
  }

  // 页面挂载时初始化
  onMounted(() => {
    language.value = languageStore()?.getLanguageKey().language.value;
    pageParams.value = JSON.parse(JSON.stringify(languageStore().getPageParams().pageParams.value));
    languageStore().setPageParams({});

    // 获取列表
    getParams();
  });
</script>

<style scoped lang="scss">
  .common-button-wrapper {
    justify-content: flex-end;
  }
  .el-radio-button {
    margin-right: 0;
  }
  .text-line-through {
    text-decoration: line-through;
    color: var(--el-color-tips);
  }
  .text-line-through:nth-child(2) {
    margin-right: 16px;
  }
  .product-info {
    .info {
      .name {
        display: flex;
        align-items: center;
      }
    }
  }
  .mr-8 {
    margin-right: 8px;
  }
  .timeline-box {
    padding-left: 1px;

    .timeline-main {
      font-size: 14px;
    }
  }
  :deep(.el-timeline-item__content) {
    display: flex;
    align-items: center;
  }
  .flex {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px;
  }
  .cupon {
    cursor: pointer;
    margin-left: 4px;
  }
  .icon {
    cursor: pointer;
    margin-left: 4px;
    font-size: 16px;
  }

  .sub-order {
    display: flex;
    gap: 24px;
    padding: 12px 0 8px 0;
  }
</style>

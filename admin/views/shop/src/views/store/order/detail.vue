<template>
  <!--
      
      时间：2019-10-25
      描述：订单详情
  -->
  <div class="pb50" v-loading="loading">
    <div class="product-content">
      <!--基本信息-->
      <div class="common-form">{{ $t('基本信息') }}</div>
      <div class="table-wrap">
        <el-row>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单类型：') }}</span>
              {{ detail.order_source_text }}{{ detail?.is_merge == '1' ? $t('(合桌)') : '' }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单号：') }}</span>
              {{ detail.order_no }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.user">
            <div class="pb16">
              <span class="gray9">{{ $t('会员：') }}</span>
              <span>{{ $t('会员ID') }}&nbsp;({{ detail.user?.user_id }})</span>
            </div>
          </el-col>
          <template v-if="uniqueUsers.length > 0">
            <el-col :span="6">
              <div class="pb16">
                <span class="gray9">{{ $t('会员：') }}</span>
                {{ $t('会员ID') }}&nbsp;(<template v-for="(item, index) in uniqueUsers"
                  ><span v-if="item">{{ item.user_id }}</span
                  ><span v-if="index < uniqueUsers.length - 1">,</span></template
                >)
              </div>
            </el-col>
          </template>

          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单金额：') }}</span>
              <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
              {{ this.$formatPrice(detail.order_price) }}
              <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
              <span v-if="currency.is_open == 1" style="padding-left: 8px">
                <template v-if="currency.vices.vice_unit_position == '0'">{{ currency.vices?.vice_unit }}</template>
                {{ this.$formatPrice(Number(detail.order_price) * Number(currency.vices?.unit_rate)) }}
                <template v-if="currency.vices.vice_unit_position == '1'">{{ currency.vices?.vice_unit }}</template>
              </span>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.express_price > 0">
            <div class="pb16">
              <span class="gray9">{{ $t('配送费：') }}</span>
              {{ detail.express_price }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.bag_price > 0">
            <div class="pb16">
              <span class="gray9">{{ $t('包装费：') }}</span>
              {{ detail.bag_price }}
            </div>
          </el-col>

          <el-col :span="6" v-if="detail.order_status.value == 30">
            <div class="pb16">
              <span class="gray9">{{ $t('实付款金额：') }}</span>
              <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
              {{ this.$formatPrice(Number(detail.pay_price) - Number(detail.refund_money)) }}
              <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
            </div>
          </el-col>
          <el-col :span="6" v-if="Number(detail.refund_money) > 0">
            <div class="pb16">
              <span class="gray9">{{ $t('退款金额：') }}</span>
              <span>
                <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
                {{ this.$formatPrice(detail.refund_money) }}
                <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
              </span>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.order_status.value == 30" v-for="item in detail.merge_parent_id == 0 ? detail.payType : detail.parentOrder.payType">
            <div class="pb16">
              <span class="gray9">{{ $t('支付方式：') }}</span>
              <span>
                {{ item.name }}
                <template v-if="item.source_text && item.value != 40 && item.value != 10 && item.value != -1"> ({{ item.source_text }}) </template>
                <el-tag class="ml-8" v-if="item.pay_status == '0'" type="danger" size="large">{{ $t('异常') }}</el-tag>
              </span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('用餐方式：') }}</span>
              {{ detail.delivery_type.text }}
            </div>
          </el-col>

          <el-col :span="6" v-if="detail.table_no || detail?.is_merge == '1'">
            <div class="pb16">
              <span class="gray9">{{ $t('桌号：') }}</span>
              <template v-if="detail?.is_merge == '1'">
                {{ tableNo(detail.mergeList) }}
              </template>
              <template v-else>
                {{ detail.table_no }}
                <span v-if="detail?.order_name">- {{ detail?.order_name }}</span>
              </template>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail?.call_no">
            <div class="pb16">
              <span class="gray9">{{ $t('序号：') }}</span>
              {{ detail?.call_no }}
              <span v-if="detail?.order_name">- {{ detail?.order_name }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('交易状态：') }}</span>
              {{ detail.order_status.value == 10 ? $t('待付款') : detail.order_status.value == 20 ? $t('已取消') : $t('已完成') }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.cashier">
            <div class="pb16">
              <span class="gray9">{{ $t('收银员：') }}</span>
              {{ detail.cashier?.real_name || '-' }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.is_buffet == 1">
            <div class="pb16">
              <span class="gray9">{{ $t('自助餐：') }}</span>
              <template v-if="detail?.buffet?.length > 0">{{ buffet }}</template>
              <template v-else>-</template>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.cancel_remark">
            <div class="pb16">
              <span class="gray9">{{ $t('取消原因：') }}</span>
              {{ detail.cancel_remark || '-' }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.is_free > 0">
            <div class="pb16">
              <span class="gray9">{{ $t('免单原因：') }}</span>
              {{ detail.free_tag_text || '-' }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('时间：') }}</span>
              {{ detail.create_time || '-' }} {{ $t('至') }} {{ detail.pay_time || '-' }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.table_no">
            <div class="pb16">
              <span class="gray9">{{ $t('备注：') }}</span>
              {{ detail.table_remark || '-' }}
            </div>
          </el-col>
        </el-row>
      </div>

      <div class="common-form mt16">
        {{ detail?.is_merge == '1' ? $t('订单信息') : $t('商品信息') }}
      </div>

      <el-radio-group v-model="activeName" class="radio-search" v-if="detail?.is_merge == '1' || (detail?.subOrder || []).length > 0">
        <template v-for="(item, index) in tableNameList">
          <el-radio-button :label="index">{{ item }}</el-radio-button>
        </template>
      </el-radio-group>
      <div class="sub-order" v-if="(detail?.subOrder || []).length > 0">
        <p class="sub-order-item">{{ $t('订单号：') }}{{ detail?.subOrder[activeName].order_no }}</p>
        <p class="sub-order-item">
          {{ $t('订单金额：') }}
          <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
          {{ this.$formatPrice(detail?.subOrder[activeName].order_price) }}
          <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
          <span v-if="currency.is_open == 1" style="padding-left: 8px">
            <template v-if="currency.vices.vice_unit_position == '0'">{{ currency.vices?.vice_unit }}</template>
            {{ this.$formatPrice(Number(detail?.subOrder[activeName].order_price) * Number(currency.vices?.unit_rate)) }}
            <template v-if="currency.vices.vice_unit_position == '1'">{{ currency.vices?.vice_unit }}</template>
          </span>
        </p>
        <p class="sub-order-item">
          {{ $t('实付金额：') }}
          <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
          {{ this.$formatPrice(Number(detail?.subOrder[activeName].pay_price) - Number(detail?.subOrder[activeName].refund_money)) }}
          <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
        </p>
        <p class="sub-order-item">
          {{ $t('会员：') }}
          <span v-if="detail?.subOrder[activeName].user">{{ $t('会员ID') }}({{ detail?.subOrder[activeName].user?.user_id }})</span>
          <span v-else>-</span>
        </p>
      </div>
      <div class="table-wrap">
        <el-table size="small" :data="detail?.is_merge == '1' || (detail?.subOrder || []).length > 0 ? tableData[activeName] : tableData" border style="width: 100%">
          <el-table-column prop="name_text" :label="$t('商品')" width="400">
            <template #default="scope">
              <div class="product-info">
                <div class="pic">
                  <img v-if="scope.row.image?.file_path == 'buffet'" src="../../../assets/img/buffet.png" />
                  <img v-else-if="scope.row.image?.file_path == 'clock'" src="../../../assets/img/clock.png" />
                  <img v-else v-img-url="scope.row.image?.file_path" />
                </div>
                <div class="info">
                  <div class="name">
                    <el-tag class="mr-8" type="danger" effect="light" :hit="true" size="large" v-if="scope.row.productReturn?.id">{{ $t('退') }}</el-tag>
                    <el-tag class="mr-8" color="#FF3300" effect="dark" :hit="true" size="large" v-if="scope.row.is_free > 0">{{ $t('赠') }}</el-tag>
                    {{ scope.row.name_text }}
                  </div>
                  <div class="gray9" v-if="scope.row.product_attr != ''">
                    {{ scope.row.product_attr }}
                  </div>
                  <div class="orange" v-if="scope.row.refund"> {{ scope.row.refund.type.text }}-{{ scope.row.refund.status.text }} </div>
                  <div class="price">
                    <span>
                      <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
                      {{ this.$formatPrice(scope.row.product_price) }}
                      <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
                      <span v-if="currency.is_open == 1" style="padding-left: 8px">
                        <template v-if="currency.vices.vice_unit_position == '0'">{{ currency.vices?.vice_unit }}</template>
                        {{ this.$formatPrice(Number(scope.row.product_price) * Number(currency.vices?.unit_rate)) }}
                        <template v-if="currency.vices.vice_unit_position == '1'">{{ currency.vices?.vice_unit }}</template>
                      </span>
                    </span>
                  </div>
                  <div>
                    <span class="gray9" v-if="scope.row.is_free > 0">{{ $t('赠送原因：') }}{{ scope.row.free_tag_text || '-' }}</span>
                    <span class="gray9" v-if="scope.row.productReturn?.reason">{{ $t('退菜原因：') }}{{ scope.row.productReturn?.reason || '-' }}</span>
                  </div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="total_num" :label="$t('购买数量')">
            <template #default="scope">
              <p>{{ scope.row.total_num }}</p>
            </template>
          </el-table-column>
          <el-table-column prop="product_price" :label="$t('商品总价')">
            <template #default="scope">
              <div>
                <template v-if="scope.row.total_consumption_tax_pay_price != scope.row.total_consumption_tax_order_price">
                  <span class="text-line-through" v-if="scope.row.total_consumption_tax_order_price">
                    <template v-if="currency.unit_position == '0'">
                      {{ currency.unit }}
                    </template>
                    {{ this.$formatPrice(Number(scope.row.total_consumption_tax_order_price)) }}
                    <template v-if="currency.unit_position == '1'">
                      {{ currency.unit }}
                    </template>
                  </span>
                  <span class="text-line-through" v-if="currency.is_open == 1 && scope.row.total_consumption_tax_order_price">
                    <template v-if="currency.vices.vice_unit_position == '0'">
                      {{ currency.vices?.vice_unit }}
                    </template>
                    {{ this.$formatPrice(Number(scope.row.total_consumption_tax_order_price) * Number(currency.vices?.unit_rate)) }}
                    <template v-if="currency.vices.vice_unit_position == '1'">
                      {{ currency.vices?.vice_unit }}
                    </template>
                  </span>
                </template>

                <span>
                  <template v-if="currency.unit_position == '0'">
                    {{ currency.unit }}
                  </template>
                  {{ this.$formatPrice(Number(scope.row.total_consumption_tax_pay_price)) }}
                  <template v-if="currency.unit_position == '1'">
                    {{ currency.unit }}
                  </template>
                </span>

                <span v-if="currency.is_open == 1">
                  <template v-if="currency.vices.vice_unit_position == '0'">
                    {{ currency.vices?.vice_unit }}
                  </template>
                  {{ this.$formatPrice(Number(scope.row.total_consumption_tax_pay_price) * Number(currency.vices?.unit_rate)) }}
                  <template v-if="currency.vices.vice_unit_position == '1'">
                    {{ currency.vices?.vice_unit }}
                  </template>
                </span>
                <span class="tips" v-if="Number(scope.row.refund_money) > 0 && scope.row.is_return != '1'">
                  （{{ $t('退款：') + currency.unit + this.$formatPrice(scope.row.refund_money) }}）
                </span>
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
              <template v-if="(activity.pay_type || []).length > 0">
                <template v-if="activity.refund_type == '2'">(</template>
                <span class="flex" v-for="(item, itemIndex) in activity.pay_type" :key="itemIndex">
                  <span>{{ item.name }}:{{ item.unit + this.$formatPrice(item.price) }}</span>
                  <el-tag v-if="item.payment_status == '0'" class="cupon" type="danger" @click="handleRetry(item)">{{ $t('退款失败，重试') }}</el-tag>
                  <el-tooltip v-if="item.value == '90333'" placement="bottom" trigger="click">
                    <el-icon class="icon"><WarningFilled /></el-icon>
                    <template #content>
                      <div> {{ $t('银行') }}{{ $t('：') }}{{ bankName(item.bank_code) }}</div>
                      <div> {{ $t('账户名') }}{{ $t('：') }}{{ item.account_no }} </div>
                      <div> {{ $t('账号') }}{{ $t('：') }}{{ item.account_name }} </div>
                    </template>
                  </el-tooltip>
                  <span v-if="itemIndex < (activity.pay_type || []).length - 1">，</span>
                </span>
                <template v-if="activity.refund_type == '2'">)</template>
              </template>
            </div>
          </el-timeline-item>
        </el-timeline>
      </div>

      <div class="table-wrap" v-if="detail.order_status.value == 20 && detail.cancel_remark != ''">
        <div class="common-form mt16">{{ $t('取消信息') }}</div>
        <div class="table-wrap">
          <el-row>
            <el-col :span="6">
              <div class="pb16">
                <span class="gray9">{{ $t('商家备注') }}:</span>
                {{ detail.cancel_remark }}
              </div>
            </el-col>
          </el-row>
        </div>
      </div>
    </div>
    <div class="common-button-wrapper">
      <el-button size="small" @click="cancelFunc">{{ $t('返回') }}</el-button>
      <el-button v-if="detail.is_refund_button == '1'" @click="refundClick(detail)" type="danger" size="small" v-auth="'/store/operate/refund'">{{ $t('退款') }}</el-button>
      <el-button v-if="detail.is_cancel_button == '1'" @click="cancelClick(detail)" type="danger" size="small" v-auth="'/store/operate/order_cancel'">{{ $t('取消') }} </el-button>
      <el-button v-if="detail.order_status.value == 20" @click="delClick(detail)" type="danger" size="small" v-auth="'/store/order/delete'">{{ $t('删除') }} </el-button>
    </div>
    <!--处理-->
    <Cancel v-if="open_edit" :open_edit="open_edit" :order_no="order_no" :order_id="order_id" @closeDialog="closeDialogFunc($event, 'edit')"> </Cancel>
    <!--处理-->
    <refund v-if="open_refund" :open_edit="open_refund" :order_no="order_no" :order_id="order_id" :pay_price="pay_price" @closeDialog="closerefundDialogFunc($event, 'edit')">
    </refund>

    <refundAgain v-if="open_refundAgain" :open_edit="open_refundAgain" :refundOrder="refundOrder" @closeDialog="closerefundAgainDialogFunc($event)"> </refundAgain>
  </div>
</template>

<script>
  import OrderApi from '@/api/order.js';
  import Cancel from './dialog/cancel.vue';
  import refund from './dialog/refund.vue';
  import refundAgain from './dialog/refundAgain.vue';
  import { useUserStore } from '@/store';
  import { languageStore } from '@/store/model/language';
  const { currency } = useUserStore();
  export default {
    components: {
      Cancel,
      refund,
      refundAgain,
    },
    data() {
      return {
        currency: currency,
        active: 0,
        /*是否加载完成*/
        loading: true,
        /*订单数据*/
        detail: {
          order_id: 0,
          pay_status: [],
          pay_type: [],
          delivery_type: [],
          user: {},
          address: [],
          product: [],
          order_status: [],
          extract: [],
          delivery_status: [],
          supplier: {
            name: '',
          },
        },
        /*是否打开编辑弹窗*/
        open_edit: false,
        open_refund: false,
        open_refundAgain: false,
        /*当前编辑的对象*/
        refundOrder: {},
        order_no: 0,
        order_id: 0,
        pay_price: 0,
        buffet: [],
        tableData: [],
        activeName: 0,
        tableNameList: [],
        pageParams: {},
        activities: [],
        bankList: [
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
        ],
      };
    },
    created() {
      this.pageParams = JSON.parse(JSON.stringify(languageStore().getPageParams().pageParams.value));
      languageStore().setPageParams({});
      /*获取列表*/
      this.getParams();
    },
    computed: {
      uniqueUsers() {
        const userIds = new Set(); // 使用 Set 来存储唯一的 user_id
        return (this.detail.subOrder || [])
          .filter((item) => {
            if (item.user && !userIds.has(item.user.user_id)) {
              userIds.add(item.user.user_id);
              return true; // 保留这个用户
            }
            return false; // 过滤掉重复的用户
          })
          .map((item) => item.user); // 返回去重后的用户对象
      },
    },

    methods: {
      next() {
        if (this.active++ > 4) this.active = 0;
      },
      /*获取参数*/
      getParams() {
        let self = this;
        // 取到路由带过来的参数
        const params = this.$route.query.order_id;
        OrderApi.storeOrderdetail(
          {
            order_id: params,
          },
          true
        )
          .then((data) => {
            self.loading = false;
            self.detail = data.data.detail;
            self.buffet = [];
            self.detail.buffet.map((item) => {
              self.buffet.push(item.name_text);
            });

            self.buffet = self.buffet.join('+');
            self.tableData = [];
            self.detail.buffetCustomerType.map((item) => {
              self.tableData.push({
                name_text: `${item.buffet_name_text}(${item.customer_type_name_text})`,
                image: { file_path: 'buffet' },
                product_attr: '',
                refund: '',
                product_price: item.price,
                total_num: item.num,
                total_consumption_tax_pay_price: item.total_consumption_tax_pay_price,
                total_consumption_tax_order_price: item.total_consumption_tax_order_price,
                refund_money: Number(item.consumption_tax_pay_price) * Number(item.refund_num),
                is_return: item.is_return,
              });
            });

            self.detail.delay.map((item) => {
              self.tableData.push({
                name_text: item.name_text,
                image: { file_path: 'clock' },
                product_attr: '',
                refund: '',
                product_price: item.price,
                total_num: self.detail.meal_num,
                total_consumption_tax_pay_price: Number(item.price) * Number(item.num),
                total_consumption_tax_order_price: Number(item.price) * Number(item.num),
                refund_money: item.refund_money,
              });
            });

            self.detail.product.map((item) => {
              self.tableData.push({
                name_text: item.product_name_text,
                image: item.image,
                product_attr: item.product_attr,
                refund: item.refund,
                product_price: item.product_price,
                total_num: item.total_num,
                total_consumption_tax_pay_price: item.total_consumption_tax_pay_price,
                total_consumption_tax_order_price: item.total_consumption_tax_order_price,
                refund_money: Number(item.consumption_tax_pay_price) * Number(item.refund_num),
                free_tag_text: item.free_tag_text,
                is_free: item.is_free,
                productReturn: item.productReturn,
                is_return: item.is_return,
              });
            });

            if (self.detail?.is_merge == '1') {
              self.dataSplicing(this.detail.mergeList, 1);
            }
            if ((self.detail?.subOrder || []).length > 0) {
              self.dataSplicing(this.detail.subOrder, 2);
            }

            self.activities = [];
            data.data.operationLog.map((item) => {
              self.activities.push({
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
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      /*合单的数据拼接*/
      dataSplicing(data, type) {
        this.tableData = [];
        this.tableNameList = [];
        (data || []).map((item, index) => {
          if (type == 1) {
            this.tableNameList.push(item.table_no ? item.table_no : item.call_no || '-');
          } else {
            this.tableNameList.push(item.table_no ? item.table_no + '-' + item.order_name : item.call_no + '-' + item.order_name || '-');
          }

          this.tableData.push([]);
          if (item.buffetCustomerType.length > 0) {
            (item.buffetCustomerType || []).map((items) => {
              this.tableData[index].push({
                name_text: `${items.buffet_name_text}(${items.customer_type_name_text})`,
                image: { file_path: 'buffet' },
                product_attr: '',
                refund: '',
                product_price: items.price,
                total_num: items.num,
                total_consumption_tax_pay_price: items.total_consumption_tax_pay_price,
                total_consumption_tax_order_price: items.total_consumption_tax_order_price,
                refund_money: Number(items.consumption_tax_pay_price) * Number(items.refund_num),
                is_return: item.is_return,
              });
            });
          }

          if (item.delay.length > 0) {
            (item.delay || []).map((items) => {
              this.tableData[index].push({
                name_text: items.name_text,
                image: { file_path: 'clock' },
                product_attr: '',
                refund: '',
                product_price: items.price,
                total_num: items.num,
                total_consumption_tax_pay_price: Number(items.price) * Number(items.num),
                total_consumption_tax_order_price: Number(items.price) * Number(items.num),
                refund_money: items.refund_money,
              });
            });
          }

          if (item.product.length > 0) {
            (item.product || []).map((items) => {
              this.tableData[index].push({
                name_text: items.product_name_text,
                image: items.image,
                product_attr: items.product_attr,
                refund: items.refund,
                product_price: items.product_price,
                total_num: items.total_num,
                total_consumption_tax_pay_price: items.total_consumption_tax_pay_price,
                total_consumption_tax_order_price: items.total_consumption_tax_order_price,
                refund_money: Number(items.consumption_tax_pay_price) * Number(items.refund_num),
                free_tag_text: items.free_tag_text,
                is_free: items.is_free,
                productReturn: items.productReturn,
                is_return: items.is_return,
              });
            });
          }
        });
        if (this.tableNameList.length > 0) {
          this.activeName = 0;
        }
      },
      /*取消*/
      cancelFunc() {
        languageStore().setPageParams(this.pageParams);
        this.$router.back(-1);
      },

      /*打开取消*/
      cancelClick(item) {
        this.order_no = item.order_no;
        this.order_id = item.order_id;
        this.open_edit = true;
      },
      refundClick(item) {
        this.order_no = item.order_no;
        this.order_id = item.order_id;
        this.pay_price = (Number(item.pay_price) - Number(item.refund_money)).toFixed(2);
        this.open_refund = true;
      },

      /*关闭弹窗*/
      closeDialogFunc(e, f) {
        if (f == 'edit') {
          this.open_edit = e.openDialog;
          if (e.type == 'success') {
            this.getParams();
          }
        }
      },
      /*关闭弹窗*/
      closerefundDialogFunc(e, f) {
        if (f == 'edit') {
          this.open_refund = e.openDialog;
          if (e.type == 'success') {
            this.getParams();
          }
        }
      },

      closerefundAgainDialogFunc(e) {
        this.open_refundAgain = e.openDialog;
        if (e.type == 'success') {
          this.getParams();
        }
      },

      delClick(item) {
        let self = this;
        ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
          type: 'warning',
        }).then(() => {
          OrderApi.storedelete({
            order_id: item.order_id,
          }).then((data) => {
            this.$ElMessage({
              message: $t('删除成功'),
              type: 'success',
            });
            this.$router.back(-1);
          });
        });
      },

      handleRetry(item) {
        this.refundOrder = item;
        if (item.value == '90333') {
          this.open_refundAgain = true;
        } else {
          ElMessageBox.confirm('确定重试退款操作?', '提示', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          })
            .then(() => {
              this.loading = true;
              const params = {
                payment_order_id: item.payment_order_id,
                refund_money: item.refund_money,
                refund_destination_id: item.refund_destination_id,
              };
              OrderApi.orderRefundAgain(params, true)
                .then((data) => {
                  this.loading = false;
                  this.$ElMessage({
                    message: $t('操作成功'),
                    type: 'success',
                  });
                  this.getParams();
                })
                .catch((error) => {
                  this.loading = false;
                });
            })
            .catch(() => {
              this.$ElMessage({
                type: 'info',
                message: '已取消操作',
              });
            });
        }
      },

      bankName(value) {
        let name = '';
        this.bankList.map((item) => {
          if (item.value == value) {
            name = item.name;
          }
        });
        return name;
      },

      tableNo(arr) {
        let result = '-';
        if (arr && arr.length > 0) {
          let nameArr = [];
          (arr || []).map((item) => {
            nameArr.push(item.table_no);
          });
          result = nameArr.join('+');
        }
        return result;
      },
    },
  };
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

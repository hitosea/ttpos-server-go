<template>
  <!--

    	时间：2019-10-26
    	描述：商品管理-商品编辑-高级设置
    -->
  <div class="buy-set-content">
    <!-- <div class="common-form">积分设置</div>
    <el-form-item for="no_click"  label="是否开启积分赠送：">
      <el-radio-group v-model="form.model.is_points_gift">
        <el-radio :label="1">开启</el-radio>
        <el-radio :label="0">关闭</el-radio>
      </el-radio-group>
    </el-form-item> -->
    <!--其他设置-->
    <div class="common-form mt50">{{ $t('其他设置') }}</div>
    <el-form-item
      for="no_click"
      v-if="form.model.product_status != 40"
      :label="$t('商品状态：')"
      :rules="[{ required: true, message: $t('选择商品状态') }]"
      prop="model.product_status"
    >
      <el-radio-group v-model="form.model.product_status">
        <el-radio :value="10">{{ $t('上架') }}</el-radio>
        <el-radio :value="20">{{ $t('下架') }}</el-radio>
      </el-radio-group>
    </el-form-item>

    <template v-for="(item, index) in form.model.productTaxes" v-if="userInfo.isOpenTax == '1' && form.model.type == 10">
      <el-form-item
        for="no_click"
        :label="returnType(item.product_tax_type)"
        :prop="`form.model.productTaxes[${index}].tax_category_id`"
        :rules="[
          {
            required: true,
            validator: () => {
              return item.tax_category_id ? true : false;
            },
            message: returnMessage(item.product_tax_type),
          },
        ]"
      >
        <el-select v-model="item.tax_category_id" clearable class="max-w460" size="default" :placeholder="$t('请选择')">
          <template v-for="cat in taxList" :key="cat.id">
            <el-option :value="cat.id" :label="cat.name"></el-option>
          </template>
        </el-select>
      </el-form-item>
    </template>

    <!--数量计算方法-->
    <el-form-item for="no_click" :label="$t('计价方式：')" v-if="form.model.type == 10" :rules="[{ required: true, message: $t('选择计价方式') }]">
      <el-radio-group v-model="form.model.num_type">
        <el-radio :value="0">{{ $t('整数') }}</el-radio>
        <el-radio :value="1">{{ $t('小数') }}</el-radio>
      </el-radio-group>
      <div class="gray9 line-height-tips">{{ $t('按整数计价：按”个/份/件“卖，数量只能是整数的商品。') }}<br />{{ $t('按小数计价：按”斤/米/升“卖，数量可带小数的商品。') }}</div>
    </el-form-item>

    <el-form-item
      for="no_click"
      v-if="form.model.type == 10"
      :label="$t('显示：')"
      :rules="[
        {
          required: true,
          validator: () => {
            return form.model.is_show_cashier || form.model.is_show_tablet || form.model.is_show_kitchen ? true : false;
          },
          message: $t('请输入名称'),
        },
      ]"
      prop="model"
    >
      <el-checkbox v-model="form.model.is_show_cashier" :true-value="1" :false-value="2" :label="$t('收银机')" size="large" />
      <el-checkbox
        v-if="is_open_tablet"
        v-model="form.model.is_show_tablet"
        :true-value="1"
        :false-value="2"
        :label="$t('平板')"
        size="large"
        :disabled="form.model.num_type == 1"
      />
      <el-checkbox v-if="is_open_kitchen_kds" v-model="form.model.is_show_kitchen" :true-value="1" :false-value="2" :label="$t('厨显')" size="large" />
      <el-checkbox
        v-if="is_open_assistant"
        v-model="form.model.is_show_assistant"
        :true-value="1"
        :false-value="2"
        :label="$t('点餐助手')"
        size="large"
        :disabled="form.model.num_type == 1"
      />
      <el-checkbox
        v-if="is_open_assistant"
        v-model="form.model.is_show_h5"
        :true-value="1"
        :false-value="2"
        :label="$t('扫码点餐')"
        size="large"
        :disabled="form.model.num_type == 1"
      />
      <el-checkbox
        v-if="is_open_delivery"
        v-model="form.model.is_show_delivery"
        :true-value="1"
        :false-value="2"
        :label="$t('外送')"
        size="large"
        :disabled="form.model.num_type == 1"
      />
    </el-form-item>

    <el-form-item v-if="form.model.type == 10" for="no_click" :label="$t('商品排序：')">
      <el-input-number :controls="false" :min="0" :max="999" :placeholder="$t('接近0，排序等级越高')" v-model="form.model.product_sort" class="max-w460"></el-input-number>
    </el-form-item>

    <el-form-item for="no_click" :label="$t('限购数量：')" v-if="form.model.type == 10">
      <el-input-number :controls="false" :min="0" :max="999" v-model="form.model.limit_num" class="max-w460"></el-input-number>
      <div class="gray9">{{ $t('每单/每桌购买的最大数量，0为不限购') }}</div>
    </el-form-item>
    <template v-if="showMore && form.model.type == 10">
      <el-form-item for="no_click" :label="$t('特色分类：')" v-if="form.model.type == 10">
        <el-select v-model="form.model.special_id" clearable class="max-w460" size="default" :placeholder="$t('请选择特色分类')">
          <template v-for="cat in form.special" :key="cat.category_id">
            <el-option :value="cat.category_id" :label="cat.name_text"></el-option>
            <template v-for="cat_c in cat.child" :key="cat_c.category_id">
              <el-option :value="cat_c.category_id" :label="cat_c.name_text">|—{{ cat_c.name_text }}</el-option>
            </template>
          </template>
        </el-select>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('商品卖点：')">
        <el-input type="textarea" :placeholder="$t('请输入商品卖点')" v-model="form.model.selling_point" show-word-limit :maxlength="50" class="max-w460"></el-input>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('供应商：')" v-if="baseSale == '1'">
        <el-select v-model="form.model.erp_supplier_id" filterable clearable class="max-w460" size="default" :placeholder="$t('请选择供应商')">
          <template v-for="item in supplierList" :key="item.id">
            <el-option :value="item.id" :label="item.name"></el-option>
          </template>
        </el-select>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('打印标签：')" prop="model.label_id">
        <el-select v-model="form.model.label_id" clearable class="max-w460" size="default" :placeholder="$t('请选择')">
          <el-option :value="0" :label="$t('无')"></el-option>
          <template v-for="cat in form.labelList" :key="cat.label_id">
            <el-option :value="cat.label_id" :label="cat.label_name_text"></el-option>
          </template>
        </el-select>
      </el-form-item>
    </template>
    <el-form-item for="no_click" label="" v-if="form.model.type == 10">
      <p
        @click="
          () => {
            showMore = !showMore;
          }
        "
        class="more-set"
        >{{ !showMore ? $t('展开更多设置') : $t('收起更多设置') }}</p
      >
    </el-form-item>
    <!--会员折扣设置-->
    <template v-if="form.model.type == 10 && is_open_member == '1'">
      <div class="common-form mt50">{{ $t('会员折扣设置') }}</div>
      <el-form-item for="no_click" :label="$t('是否开启会员折扣：')">
        <el-radio-group v-model="form.model.is_enable_grade">
          <el-radio :value="1">{{ $t('开启') }}</el-radio>
          <el-radio :value="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('会员折扣设置：')" v-if="form.model.is_enable_grade == 1">
        <el-radio-group v-model="form.model.is_alone_grade">
          <el-radio :value="0">{{ $t('默认折扣') }}</el-radio>
          <!-- <el-radio :label="1">{{ $t('仅需支付') }}</el-radio> -->
        </el-radio-group>
        <div class="gray9" v-if="form.model.is_alone_grade == 0">{{ $t('默认折扣：默认为用户所属会员等级的折扣率') }}</div>
        <div class="gray9" v-if="form.model.is_alone_grade == 1">{{ $t('仅需支付：用户购买此商品仅需支付的金额或比例') }}</div>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('折扣佣金类型：')" v-if="form.model.is_alone_grade == 1 && form.model.is_enable_grade == 1">
        <el-radio-group v-model="form.model.alone_grade_type" @change="changeGradeType">
          <el-radio :value="10">{{ $t('百分比') }}</el-radio>
          <el-radio :value="20">{{ $t('固定金额') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item for="no_click" label="" v-if="form.model.is_alone_grade == 1 && form.model.is_enable_grade == 1">
        <div class="percent-w50">
          <el-table :data="form.gradeList" border size="">
            <el-table-column prop="name" :label="$t('会员等级')"> </el-table-column>
            <el-table-column prop="name" :label="$t('折扣')">
              <template #default="scope">
                <div class="d-s-c">
                  <el-form-item
                    for="no_click"
                    class="product-equity"
                    :rules="[
                      {
                        validator: () => {
                          return scope.row.product_equity ? true : false;
                        },
                        message: $t('请输入折扣'),
                      },
                    ]"
                    prop="model.image"
                  >
                    <el-input-number
                      v-model="scope.row.product_equity"
                      :min="form.model.alone_grade_type == 10 ? 1 : 0"
                      :max="form.model.alone_grade_type == 10 ? 100 : minPrice"
                      :controls="false"
                      :placeholder="$t('请输入折扣')"
                    ></el-input-number>
                    <span class="ml10">{{ form.model.alone_grade_type == 10 ? grade_unit : currency.unit }}</span>
                  </el-form-item>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-form-item>
    </template>

    <!--整单折扣-->
    <div class="common-form mt50">{{ $t('整单折扣') }}</div>
    <el-form-item for="no_click" :label="$t('开启整单折扣：')" prop="model.open_overall_discount" :rules="[{ required: true, message: $t('请选择是否开启整单折扣') }]">
      <el-radio-group v-model="form.model.open_overall_discount">
        <el-radio :value="1">{{ $t('开启') }}</el-radio>
        <el-radio :value="0">{{ $t('关闭') }}</el-radio>
      </el-radio-group>
      <div class="gray9">{{ $t('开启整单折扣时，设置折扣的商品会参与整单打折；关闭后则不再参与') }}</div>
    </el-form-item>
  </div>
</template>

<script>
  import { useUserStore } from '@/store';
  import PorductApi from '@/api/product.js';
  import PurchaseApi from '@/api/purchase.js';
  const { currency, userInfo } = useUserStore();
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const baseSale = supplier.value?.sale_stock || 0;
  const is_open_tablet = supplier.value?.is_open_tablet || 0;
  const is_open_assistant = supplier.value?.is_open_assistant || 0;
  const is_open_kitchen_kds = supplier.value?.is_open_kitchen_kds || 0;
  const is_open_member = supplier.value?.is_open_member || 0;
  const is_open_delivery = supplier.value?.delivery_status || 0;
  export default {
    data() {
      return {
        unit: '%',
        grade_unit: '%',
        currency: currency,
        minPrice: 0,
        userInfo: userInfo,
        taxList: [],
        supplierList: [],
        is_open_tablet: is_open_tablet,
        is_open_assistant: is_open_assistant,
        is_open_kitchen_kds: is_open_kitchen_kds,
        is_open_member: is_open_member,
        is_open_delivery: is_open_delivery,
        baseSale: baseSale,
        showMore: false,
      };
    },
    created() {
      if (this.form.model.alone_grade_type == '20') {
        this.grade_unit = '元';
      }
      if (this.form.model.agent_money_type == '20') {
        this.unit = '元';
      }
      if (this.userInfo.isOpenTax == '1') {
        this.getTaxData();
      }
      this.getData();
      //权限判断
      if (!this.is_open_tablet) {
        this.form.model.is_show_tablet = 2;
      }
      if (!this.is_open_assistant) {
        this.form.model.is_show_assistant = 2;
      }
      if (!this.is_open_kitchen_kds) {
        this.form.model.is_show_kitchen = 2;
      }
    },
    inject: ['form'],
    watch: {
      'form': {
        handler(val) {
          let price = [];
          val.model.sku.map((item) => {
            price.push(item.product_price);
          });
          this.minPrice = Math.min(...price);
        },
        immediate: true,
        deep: true,
      },
      'form.model.num_type': {
        handler(val) {
          if (val == 1) {
            this.form.model.is_show_tablet = 2;
            this.form.model.is_show_assistant = 2;
            this.form.model.is_show_h5 = 2;
            this.form.model.is_show_delivery = 2;
          }
        },
      },
    },
    methods: {
      /*获取基础数据*/
      getTaxData: function () {
        let self = this;
        PorductApi.getTaxList({}, true)
          .then((res) => {
            this.taxList = res.data.list;
            let idArr = [];
            this.taxList.map((item) => {
              idArr.push(item.id);
            });
            this.form.model.productTaxes.map((item) => {
              if (!idArr.includes(item.tax_category_id)) {
                item.tax_category_id = '';
              }
            });
          })
          .catch((error) => {});
      },

      //获取供应商
      getData() {
        let self = this;
        let Params = {};
        Params.list_rows = 1000;
        PurchaseApi.supplierList(Params, true)
          .then((data) => {
            self.loading = false;
            self.supplierList = data.data.list.data;
            self.$nextTick(() => {
              let arr = [];
              self.supplierList.map((item) => {
                arr.push(item.id);
              });
              if (!arr.includes(self.form.model.erp_supplier_id)) {
                self.form.model.erp_supplier_id = '';
              }
            });
          })
          .catch((error) => {});
      },

      /*换算单位*/
      changeMoneyType: function (val) {
        if (val == '10') {
          this.unit = '%';
        } else {
          this.unit = '元';
        }
      },
      /*换算单位*/
      changeGradeType: function (val) {
        this.form.gradeList.map((item, index) => {
          this.form.gradeList[index].product_equity = null;
        });
        if (val == '10') {
          this.grade_unit = '%';
        } else {
          this.grade_unit = '元';
        }
      },

      returnType(type) {
        let result = '';
        if (type == '1') {
          result = $t('堂食税类：');
        } else {
          result = $t('外带税类：');
        }
        return result;
      },
      returnMessage(type) {
        let result = '';
        if (type == '1') {
          result = $t('请选择堂食税类');
        } else {
          result = $t('请选择外带税类');
        }
        return result;
      },
    },
  };
</script>
<style lang="scss" scoped>
  :deep(.el-input__wrapper) {
    padding-left: 7px !important;
    padding-right: 7px !important;
  }

  .product-equity {
    display: flex;
    align-items: center;
    width: 100%;
    margin-top: 16px;

    :deep(.el-form-item__content) {
      flex-wrap: nowrap;
    }
  }
  .more-set {
    color: var(--el-color-primary);
    cursor: pointer;
    font-size: 14px;
  }
  .line-height-tips {
    line-height: 1.4;
    margin-top: 4px;
  }
</style>

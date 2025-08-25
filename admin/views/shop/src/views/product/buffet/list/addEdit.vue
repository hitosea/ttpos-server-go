<template>
  <el-dialog :title="title" v-model="dialogVisible" @close="handleClose" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <UniqueNameForm
        ref="uniqueNameFormRef"
        :labelPrefix="$t('自助餐名称')"
        :placeholder="$t('请输入自助餐名称')"
        apiSource="buffet"
        :apiId="form.id"
        :overrideLanguages="form.name"
      />

      <el-form-item
        for="no_click"
        :label="$t('排序')"
        prop="sort"
        :rules="[
          {
            required: true,
            message: $t('排序不能为空'),
          },
          {
            type: 'number',
            message: $t('排序必须为数字'),
          },
        ]"
      >
        <el-input-number
          :controls="false"
          :precision="0"
          :min="0"
          :max="999"
          :placeholder="$t('接近0，排序等级越高')"
          v-model.number="form.sort"
          autocomplete="off"
        ></el-input-number>
      </el-form-item>

      <template v-if="userInfo.isOpenTax == '1'"
        ><template v-for="(item, index) in form.buffetTaxes" :key="index">
          <el-form-item
            for="no_click"
            :label="returnType(item.buffet_tax_type)"
            :prop="`buffetTaxes.${index}.tax_category_id`"
            :rules="[
              {
                required: true,
                message: returnMessage(item.buffet_tax_type),
              },
            ]"
          >
            <el-select v-model="item.tax_category_id" clearable size="default">
              <template v-for="cat in taxList" :key="cat.id">
                <el-option :value="cat.id" :label="cat.name"></el-option>
              </template>
            </el-select>
          </el-form-item> </template
      ></template>

      <el-form-item for="no_click" :label="$t('限制用餐时间')" prop="is_time_limit" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.is_time_limit">
          <el-radio :value="0">{{ $t('不限制') }}</el-radio>
          <el-radio :value="1">{{ $t('限制') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item for="no_click" v-if="form.is_time_limit == 1" label="" class="display-none" prop="time_limit" :rules="[{ required: true, message: $t('请输入用餐时间') }]">
        <el-input-number :controls="false" :min="0" :max="999" :placeholder="$t('请输入用餐时间')" v-model.number="form.time_limit"></el-input-number>
        {{ $t('分') }}
      </el-form-item>

      <el-form-item for="no_click" :label="$t('状态')" prop="status" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.status">
          <el-radio :value="1">{{ $t('开启') }}</el-radio>
          <el-radio :value="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('组合')" prop="is_comb" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.is_comb">
          <el-radio :value="1">{{ $t('开启') }}</el-radio>
          <el-radio :value="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="this.is_open_tablet || this.is_open_scan" for="no_click" :label="nameReturn()" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.is_remain_continue">
          <el-radio :value="1">{{ $t('开') }}</el-radio>
          <el-radio :value="0">{{ $t('关') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <template v-if="form.is_remain_continue == '1'">
        <el-form-item for="no_click" label="" :rules="[{ required: true, message: $t('请输入时间') }]" prop="remain_continue_time">
          {{ $t('剩余') }}
          <el-input-number
            :controls="false"
            :precision="0"
            :min="0"
            :max="maxNum"
            style="width: 200px !important; margin: 0 4px"
            :placeholder="$t('请输入时间')"
            v-model.number="form.remain_continue_time"
          ></el-input-number>
          {{ $t('分') }}{{ $t('不可继续点单') }}
          <div class="gray9">{{ $t('如：设置20分钟，自助餐时间为90分钟，则用餐70分钟后，不可继续点自助餐内商品') }}</div>
        </el-form-item>
        <el-form-item for="no_click" label="" :rules="[{ required: true, message: $t('请输入时间') }]" prop="remain_continue_notice_time">
          {{ $t('剩余') }}
          <el-input-number
            :controls="false"
            :precision="0"
            :min="0"
            :max="maxNum"
            style="width: 200px !important; margin: 0 4px"
            :placeholder="$t('请输入时间')"
            v-model.number="form.remain_continue_notice_time"
          ></el-input-number>
          {{ $t('分') }}{{ $t('提醒不可继续点单') }}
        </el-form-item>
      </template>

      <el-form-item
        for="no_click"
        :label="$t('顾客类型')"
        prop="customer_type"
        required
        :rules="[
          {
            validator: (rule, value, callback) => {
              return form.customer_type.length > 0 ? callback() : callback($t('请选择顾客类型'));
            },
          },
        ]"
      >
        <el-button type="primary" @click="addCustomerType" :disabled="form.customer_type.length >= 5">{{ $t('添加') }}</el-button>
        <div class="customer-type" v-if="form.customer_type.length > 0">
          <template v-for="(item, index) in form.customer_type" :key="index">
            <div class="customer-button">
              <el-form-item
                for="no_click"
                label=""
                style="margin-top: 16px; width: 100%"
                :prop="`customer_type.${index}.customer_type_id`"
                :rules="[
                  {
                    required: true,
                    message: $t('请选择顾客类型'),
                  },
                ]"
              >
                <el-select v-model="item.customer_type_id" filterable clearable :placeholder="$t('请选择顾客类型')">
                  <template v-for="(item, index) in customerList" :key="index">
                    <el-option :value="item.id" :label="item.name_text">{{ item.name_text }}</el-option>
                  </template>
                </el-select>
              </el-form-item>
              <el-form-item
                for="no_click"
                label=""
                style="margin-top: 16px"
                :prop="`customer_type.${index}.price`"
                :rules="[
                  {
                    required: true,
                    message: $t('请输入价格'),
                  },
                ]"
              >
                <el-input-number
                  :controls="false"
                  :min="0"
                  :max="100000000"
                  @change="numChange(index)"
                  style="width: 200px !important"
                  :placeholder="$t('请输入价格')"
                  v-model.number="item.price"
                ></el-input-number>
              </el-form-item>
              <el-icon class="delete-icon" @click="handleDeleteCustomer(index)">
                <CircleCloseFilled />
              </el-icon>
            </div>
          </template>
        </div>
      </el-form-item>

      <el-form-item
        for="no_click"
        :label="$t('商品')"
        prop="product_ids"
        :rules="[
          {
            required: true,
            message: $t('请选中商品'),
          },
        ]"
      >
        <el-button type="primary" @click="selectList('select')">{{ $t('选中商品') }}</el-button>
        <div class="select-list" v-if="select_list.length > 0">
          <template v-for="(item, index) in select_list" :key="index">
            <div class="select-button">
              <div class="select-p">
                <autoTips :content="item.product_name_text">{{ item.product_name_text }}</autoTips>
              </div>
              <div class="select-check">
                <el-checkbox v-model="item.is_show_cashier" :true-value="1" :false-value="2" :label="$t('收银机')" size="large" />
                <el-checkbox v-if="is_open_tablet" v-model="item.is_show_tablet" :true-value="1" :false-value="2" :label="$t('平板')" size="large" />
                <el-checkbox v-if="is_open_kitchen_kds" v-model="item.is_show_kitchen" :true-value="1" :false-value="2" :label="$t('厨显')" size="large" />
                <el-checkbox v-if="is_open_assistant" v-model="item.is_show_assistant" :true-value="1" :false-value="2" :label="$t('点餐助手')" size="large" />
              </div>
              <el-icon class="select-icon" @click="deleteOne(index, item.product_id)">
                <CircleCloseFilled />
              </el-icon>
            </div>
          </template>
        </div>
      </el-form-item>

      <el-form-item
        for="no_click"
        :label="$t('限购')"
        prop="buy_limit_products"
        :rules="[
          {
            required: true,
            validator: () => {
              return form.buy_limit_products.length == 0 && form.buy_limit_status == 1 ? false : true;
            },
            message: $t('请选中商品'),
          },
        ]"
      >
        <el-radio-group v-model="form.buy_limit_status">
          <el-radio :value="1">{{ $t('开启') }}</el-radio>
          <el-radio :value="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
        <div class="limit-list" v-if="form.buy_limit_status == 1">
          <el-button type="primary" @click="selectList('limit')" :disabled="!limit_ids">{{ $t('选中商品') }}</el-button>
          <div class="limit-product">
            <template v-for="(item, index) in form.buy_limit_products" :key="index">
              <div class="limit-product-list">
                <div class="limit-product-box">
                  <el-input type="text" v-model="item.name" readonly></el-input>
                  <el-form-item
                    for="no_click"
                    label=""
                    style="margin-top: 16px"
                    :prop="`buy_limit_products.${index}.limit_num`"
                    :rules="[
                      {
                        required: true,
                        message: $t('请输入限购数量'),
                      },
                      {
                        type: 'number',
                        message: $t('限购数量必须为数字'),
                      },
                    ]"
                  >
                    <el-input-number
                      :controls="false"
                      :min="0"
                      :max="999"
                      style="width: 200px !important"
                      :placeholder="$t('请输入限购数量')"
                      v-model.number="item.limit_num"
                    ></el-input-number>
                  </el-form-item>

                  <el-icon class="delete-icon" @click="handleDelete(index)">
                    <CircleCloseFilled />
                  </el-icon>
                </div>
              </div>
            </template>
          </div>
        </div>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('整单折扣')" prop="open_overall_discount" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.open_overall_discount">
          <el-radio :value="1">{{ $t('开启') }}</el-radio>
          <el-radio :value="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">{{ $t('取消') }}</el-button>
        <el-tooltip effect="dark" placement="top" :content="$t('如长时间无响应，请刷新后重试。')">
          <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
        </el-tooltip>
      </div>
    </template>
    <productList
      v-if="open_product"
      :open_product="open_product"
      :limit_ids="limit_ids"
      :selectType="selectType"
      :multiple_selection="multiple_selection"
      :material_type="10"
      @closeDialogFunc="closeDialogFunc($event)"
    >
    </productList>
  </el-dialog>
</template>

<script>
  import ProductApi from '@/api/product.js';
  import autoTips from './autoTips.vue';
  import productList from '@/components/productList/productList.vue';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';

  import { useUserStore } from '@/store';

  const { userInfo, computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_tablet = supplier.value?.is_open_tablet || 0;
  const is_open_assistant = supplier.value?.is_open_assistant || 0;
  const is_open_kitchen_kds = supplier.value?.is_open_kitchen_kds || 0;
  const is_open_scan = supplier.value?.is_open_scan || 0;

  export default {
    name: 'ProductBuffetListAddEdit',
    components: { productList, autoTips, UniqueNameForm },
    data() {
      return {
        dialogVisible: false,
        loading: false,
        open_product: false,

        form: {
          id: undefined,
          name: {},
          sort: null,
          is_time_limit: 1,
          time_limit: 90,
          status: 1,
          is_comb: 1,
          buy_limit_status: 0,
          buy_limit_products: [],
          products: [],
          customer_type: [],
          is_remain_continue: 1,
          remain_continue_notice_time: 5,
          remain_continue_time: 20,
          buffetTaxes: [
            {
              buffet_tax_type: '1',
              tax_category_id: '',
            },
          ],
          product_ids: [],
          open_overall_discount: 1,
        },
        select_list: [],
        limit_list: [],
        multiple_selection: [],
        limit_ids: '',
        selectType: '',
        customerList: [],
        taxList: [],
        userInfo: userInfo,
        is_open_tablet: is_open_tablet,
        is_open_assistant: is_open_assistant,
        is_open_kitchen_kds: is_open_kitchen_kds,
        is_open_scan: is_open_scan,
      };
    },
    props: {
      open_dialog: {
        type: Boolean,
        default: false,
      },
      title: {
        default: '',
      },
      editData: {
        default: '',
      },
    },
    computed: {
      maxNum() {
        let result = 999;
        if (this.form.is_time_limit == 1 && this.form.time_limit > 0) {
          result = this.form.time_limit - 1;
        }
        return result;
      },
    },
    created() {
      this.dialogVisible = this.open_dialog;
      this.getCustomer();
      if (this.userInfo.isOpenTax == '1') {
        this.getTaxData();
      }
      if (this.editData) {
        const copyData = JSON.parse(JSON.stringify(this.editData));
        this.form.id = copyData.id;
        try {
          const _names = typeof copyData.name === 'string' ? JSON.parse(copyData.name) : copyData.name ?? {};
          this.form.name = _names;
        } catch (error) {
          console.error('parse name faild', error);
        }
        this.form.sort = Number(copyData.sort);
        this.form.is_time_limit = copyData.time_limit > 0 ? 1 : 0;
        this.form.time_limit = Number(copyData.time_limit);
        this.form.status = copyData.status;
        this.form.is_comb = copyData.is_comb;
        this.form.is_remain_continue = copyData.is_remain_continue;
        this.form.remain_continue_time = copyData.remain_continue_time;
        this.form.remain_continue_notice_time = copyData.remain_continue_notice_time;
        this.select_list = copyData.buffetProducts.map((item) => ({ ...item, product_name_text: item.product.product_name_text }));
        this.form.price = Number(copyData.price);
        this.form.product_ids = copyData.buffetProducts.map((item) => item.product_id);
        this.form.open_overall_discount = copyData.open_overall_discount;

        this.form.customer_type = copyData.buffetCustomerType.map((item) => ({ ...item, price: Number(item.price || 0) }));
        this.form.buy_limit_status = copyData.buy_limit_status;
        this.limit_ids = this.form.product_ids.join(',');
        this.limit_list = copyData.buffetLimitProducts;
        this.form.buy_limit_products = this.limit_list.map(({ product, product_id, limit_num }) => ({ name: product.product_name_text, product_id, limit_num }));
        //税率处理
        this.form.buffetTaxes = [];
        if (copyData.buffetTaxes.length == 0) {
          this.form.buffetTaxes = [
            {
              buffet_tax_type: '1',
              tax_category_id: '',
            },
          ];
        } else {
          this.form.buffetTaxes = copyData.buffetTaxes.map((item) => ({ buffet_tax_type: item.buffet_tax_type, tax_category_id: item.tax_category_id }));
        }
      }

      if (!this.is_open_tablet && !this.is_open_scan) {
        this.form.is_remain_continue = '0';
      }
    },
    methods: {
      /*获取基础数据*/
      getTaxData() {
        const self = this;
        ProductApi.getTaxList({}, true)
          .then((res) => {
            self.taxList = res.data.list;
          })
          .catch(() => {});
      },
      /*获取基础数据*/
      getCustomer() {
        const self = this;
        self.loading = true;
        let params = {};
        params.page = 1;
        params.list_rows = 100;
        ProductApi.getCustomerList(params, true)
          .then((data) => {
            self.loading = false;
            self.customerList = data.data.list;
            const ids = this.customerList.map((item) => item.id);
            this.form.customer_type.forEach((item, index) => {
              if (!ids.includes(item.customer_type_id)) {
                this.form.customer_type[index].customer_type_id = '';
              }
            });
          })
          .catch(() => {
            self.loading = false;
          });
      },

      async submit() {
        const self = this;
        // 调用表单验证方法，valid为验证结果
        self.loading = true;
        try {
          const validForm = await self.$refs.formRef.validate();
          if (!validForm) return;

          const validUniqueName = await self.$refs.uniqueNameFormRef.validate();
          if (!validUniqueName) return;

          const _name = self.$refs.uniqueNameFormRef.data;
          const params = JSON.parse(JSON.stringify(self.form));
          params.name = JSON.stringify(_name);

          // 将customer_type字段转换为数组
          params.customer_type = self.form.customer_type.map(({ customer_type_id, price }) => ({ customer_type_id, price }));
          // 将buy_limit_products字段转换为数组
          params.buy_limit_products = (self.form?.buy_limit_products || []).map(({ product_id, limit_num }) => ({ product_id, limit_num }));
          // 将select_list字段转换为数组
          params.products = self.select_list.map(({ product_id, is_show_cashier, is_show_tablet, is_show_kitchen, is_show_assistant }) => ({
            product_id,
            is_show_cashier,
            is_show_tablet,
            is_show_kitchen,
            is_show_assistant,
          }));

          // 如果时间限制开启，且时间限制小于等于用餐时间或用餐时间提醒时间，则提示错误信息
          if (self.form.is_time_limit == 1 && (self.form.time_limit <= self.form.remain_continue_time || self.form.time_limit <= self.form.remain_continue_notice_time)) {
            self.$ElMessage({
              message: self.$t('平板时间不能大于用餐时间'),
              type: 'warning',
            });
            return;
          }

          //权限判断
          params.products.map((item) => {
            // 如果平板权限关闭，则将is_show_tablet字段赋值为2
            if (!self.is_open_tablet) {
              item.is_show_tablet = 2;
            }
            // 如果助手权限关闭，则将is_show_assistant字段赋值为2
            if (!self.is_open_assistant) {
              item.is_show_assistant = 2;
            }
            // 如果厨房权限关闭，则将is_show_kitchen字段赋值为2
            if (!self.is_open_kitchen_kds) {
              item.is_show_kitchen = 2;
            }
          });

          try {
            if (self.editData) {
              params.buffet_id = params.id;
              await ProductApi.editBuffet(params, true);
            } else {
              await ProductApi.addBuffet(params, true);
            }
            self.$ElMessage({
              message: self.editData ? self.$t('编辑成功') : self.$t('添加成功'),
              type: 'success',
            });
            // 关闭对话框
            self.handleClose(true);
          } catch (err) {
            //
          }
        } catch (error) {
          self.scrollToError();
        } finally {
          self.loading = false;
        }
      },

      scrollToError() {
        setTimeout(() => {
          const errorItems = document.querySelectorAll('.el-form-item__error');
          if (errorItems.length > 0) {
            const firstErrorItem = errorItems[0];
            firstErrorItem.scrollIntoView({ behavior: 'smooth', block: 'center' });
          }
        }, 200);
      },

      selectList(e) {
        if (e == 'select') {
          this.selectType = e;
          this.multiple_selection = this.select_list;
        }
        if (e == 'limit') {
          this.selectType = e;
          this.multiple_selection = this.limit_list;
          this.limit_ids = this.form.product_ids.join(',');
        }
        this.open_product = true;
      },

      /*关闭弹窗*/
      closeDialogFunc(e) {
        this.open_product = e.openDialog;
        if (e.type == 'select') {
          e.data.map((item) => {
            if (!this.form.product_ids.includes(item.product_id)) {
              this.select_list.push({
                product_id: item.product_id,
                product_name_text: item.product_name_text,
                is_show_cashier: 1,
                is_show_kitchen: 1,
                is_show_tablet: 1,
                is_show_assistant: 1,
              });
            }
          });

          this.form.product_ids = [];
          this.select_list.map((item) => {
            this.form.product_ids.push(item.product_id);
          });

          this.$refs.formRef.validateField('product_ids');

          this.limit_ids = this.form.product_ids.join(',');
        }
        if (e.type == 'limit') {
          let map = new Map();
          [this.limit_list, e.data].flat().forEach((obj) => map.set(obj.product_id, obj));
          this.limit_list = Array.from(map.values());
          let arr = [];
          this.form.buy_limit_products.map((item) => {
            arr.push(item.product_id);
          });
          this.limit_list.map((item) => {
            if (!arr.includes(item.product_id)) {
              this.form.buy_limit_products.push({
                name: item.product_name_text,
                product_id: item.product_id,
                limit_num: null,
              });
            }
          });
          this.$refs.formRef.validateField('buy_limit_products');
        }
      },

      deleteOne(index, product_id) {
        this.select_list.splice(index, 1);
        this.form.product_ids = [];
        this.select_list.map((item) => {
          this.form.product_ids.push(item.product_id);
        });
        this.limit_ids = this.form.product_ids.join(',');
        this.form.buy_limit_products.map((item, index) => {
          if (product_id == item.product_id) {
            this.handleDelete(index);
          }
        });
        this.$refs.formRef.validateField('product_ids');
      },

      handleDelete(index) {
        this.form.buy_limit_products.splice(index, 1);
        this.limit_list.splice(index, 1);
        this.$refs.formRef.validateField('buy_limit_products');
      },

      /*关闭弹窗*/
      handleClose(isSuccess = false, data) {
        this.$emit('closeDialog', {
          type: isSuccess ? 'success' : 'error',
          openDialog: false,
          data: data,
        });
      },
      // 添加顾客类型
      addCustomerType() {
        this.form.customer_type.push({
          customer_type_id: '',
          price: null,
        });
      },
      // 删除顾客类型
      handleDeleteCustomer(index) {
        this.form.customer_type.splice(index, 1);
      },

      numChange(index) {
        this.$nextTick(() => {
          this.form.customer_type[index].price = Number(this.$priceTwo(this.form.customer_type[index].price));
        });
      },

      //名字返回
      nameReturn() {
        if (this.is_open_tablet) {
          if (this.is_open_scan) {
            return this.$t('平板/扫码H5时间');
          }
          return this.$t('平板时间');
        } else {
          if (this.is_open_scan) {
            return this.$t('扫码H5时间');
          }
        }
        return '';
      },

      returnType(type) {
        return type == '1' ? this.$t('堂食税类：') : this.$t('外带税类：');
      },
      returnMessage(type) {
        return type == '1' ? this.$t('请选择堂食税类') : this.$t('请选择外带税类');
      },
      /*翻译*/
      translate(lang) {
        this.languageList.map((item) => {
          lang.map((items) => {
            let key = item.name;
            if (key == 'zhtw') {
              key = 'zh-TW';
            }
            if (items[key]) {
              this.form.name[item.key] = items[key];
            }
          });
        });
      },
    },
  };
</script>

<style lang="scss" scoped>
  .select-list {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 16px;
    padding: 7px 7px 0 0;
    max-height: 400px;
    overflow: auto;
  }

  .customer-type {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 16px;
    padding: 7px 7px 0 0;
    max-height: 450px;
    overflow: auto;

    .customer-button {
      width: 100%;
      display: flex;
      border: solid 1px var(--el-color-tips);
      border-radius: 4px;
      padding: 6px 12px;
      gap: 12px;
      align-items: center;
      position: relative;

      .delete-icon {
        position: absolute;
        right: -7px;
        top: -7px;
        cursor: pointer;
        color: #c80000;
      }
    }
  }

  .limit-list {
    width: 100%;
    margin-top: 12px;

    .limit-product {
      max-height: 400px;
      overflow: auto;
    }

    .limit-product-list {
      width: 100%;
      display: flex;
      flex-direction: column;
      gap: 12px;
      margin-top: 16px;
      padding-right: 7px;

      .limit-product-box {
        display: flex;
        border: solid 1px var(--el-color-tips);
        border-radius: 4px;
        padding: 6px 12px;
        gap: 12px;
        align-items: center;
        position: relative;

        .delete-icon {
          position: absolute;
          right: -7px;
          top: -7px;
          cursor: pointer;
          color: #c80000;
        }
      }
    }
  }

  .display-none {
    :deep(.el-form-item__content) {
      display: flex;
      flex-wrap: nowrap;
      gap: 12px;
    }
  }

  .select-button {
    flex: 1;
    min-width: 100%;
    border: solid 1px var(--el-color-tips);
    color: var(--el-color-tips);
    border-radius: 4px;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: space-between;

    .select-p {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      position: relative;
    }

    .select-check {
      flex-shrink: 0;
      padding-right: 12px;
    }

    .select-icon {
      position: absolute;
      right: -7px;
      top: -7px;
      cursor: pointer;
      color: #c80000;
    }
  }
</style>

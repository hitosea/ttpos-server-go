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
      <el-form-item v-if="is_open_tablet || is_open_scan" for="no_click" :label="nameReturn()" :rules="[{ required: true, message: '' }]">
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

<script setup>
// 中文注释：改为箭头函数语法，统一在语句末尾添加分号
import { ref, reactive, computed, watch, onMounted, nextTick, getCurrentInstance } from 'vue';
import ProductApi from '@/api/product.js';
import autoTips from './autoTips.vue';
import productList from '@/components/productList/productList.vue';
import UniqueNameForm from '@/components/product/UniqueNameForm.vue';
import { useUserStore } from '@/store';
import { storeToRefs } from 'pinia';

const props = defineProps({
  open_dialog: { type: Boolean, default: false },
  title: { type: String, default: '' },
  editData: { type: [String, Object], default: '' },
});
const emit = defineEmits(['closeDialog']);

const { proxy } = getCurrentInstance();

const userStore = useUserStore();
const { userInfo } = storeToRefs(userStore);
const { computedSupplier } = userStore;
const supplier = computedSupplier().supplier;
const is_open_tablet = computed(() => supplier.value?.is_open_tablet || 0);
const is_open_assistant = computed(() => supplier.value?.is_open_assistant || 0);
const is_open_kitchen_kds = computed(() => supplier.value?.is_open_kitchen_kds || 0);
const is_open_scan = computed(() => supplier.value?.is_open_scan || 0);

const dialogVisible = ref(false);
const loading = ref(false);
const open_product = ref(false);

const form = reactive({
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
});

const select_list = ref([]);
const limit_list = ref([]);
const multiple_selection = ref([]);
const limit_ids = ref('');
const selectType = ref('');
const customerList = ref([]);
const taxList = ref([]);

const formRef = ref();
const uniqueNameFormRef = ref();

const maxNum = computed(() => {
  let result = 999;
  if (form.is_time_limit == 1 && form.time_limit > 0) {
    result = form.time_limit - 1;
  }
  return result;
});

watch(
  () => props.open_dialog,
  (val) => {
    dialogVisible.value = val;
  },
  { immediate: true }
);

const initFromEditData = () => {
  const copyData = props.editData ? JSON.parse(JSON.stringify(props.editData)) : null;
  if (!copyData) return;

  form.id = copyData.id;
  try {
    const _names = typeof copyData.name === 'string' ? JSON.parse(copyData.name) : copyData.name ?? {};
    form.name = _names;
  } catch (error) {
    console.error('parse name faild', error);
  }
  form.sort = Number(copyData.sort);
  form.is_time_limit = copyData.time_limit > 0 ? 1 : 0;
  form.time_limit = Number(copyData.time_limit);
  form.status = copyData.status;
  form.is_comb = copyData.is_comb;
  form.is_remain_continue = copyData.is_remain_continue;
  form.remain_continue_time = copyData.remain_continue_time;
  form.remain_continue_notice_time = copyData.remain_continue_notice_time;

  select_list.value = (copyData.buffetProducts || []).map((item) => ({ ...item, product_name_text: item.product.product_name_text }));
  form.price = Number(copyData.price);
  form.product_ids = (copyData.buffetProducts || []).map((item) => item.product_id);
  form.open_overall_discount = copyData.open_overall_discount;

  form.customer_type = (copyData.buffetCustomerType || []).map((item) => ({ ...item, price: Number(item.price || 0) }));
  form.buy_limit_status = copyData.buy_limit_status;
  limit_ids.value = form.product_ids.join(',');
  limit_list.value = copyData.buffetLimitProducts || [];
  form.buy_limit_products = limit_list.value.map(({ product, product_id, limit_num }) => ({ name: product.product_name_text, product_id, limit_num }));

  form.buffetTaxes = [];
  if ((copyData.buffetTaxes || []).length == 0) {
    form.buffetTaxes = [
      {
        buffet_tax_type: '1',
        tax_category_id: '',
      },
    ];
  } else {
    form.buffetTaxes = copyData.buffetTaxes.map((item) => ({ buffet_tax_type: item.buffet_tax_type, tax_category_id: item.tax_category_id }));
  }
};

watch(
  () => props.editData,
  () => {
    initFromEditData();
  },
  { immediate: true }
);

onMounted(async () => {
  await getCustomer();
  if (userInfo.value.isOpenTax == '1') {
    await getTaxData();
  }
  if (!is_open_tablet.value && !is_open_scan.value) {
    form.is_remain_continue = '0';
  }
});

const getTaxData = async () => {
  try {
    const res = await ProductApi.getTaxList({}, true);
    taxList.value = res.data.list;
  } catch (e) {
    // 忽略错误
  }
};

const getCustomer = async () => {
  loading.value = true;
  try {
    const params = { page: 1, list_rows: 100 };
    const data = await ProductApi.getCustomerList(params, true);
    customerList.value = data.data.list;
    const ids = customerList.value.map((item) => item.id);
    (form.customer_type || []).forEach((item, index) => {
      if (!ids.includes(item.customer_type_id)) {
        form.customer_type[index].customer_type_id = '';
      }
    });
  } catch (e) {
    // 忽略错误
  } finally {
    loading.value = false;
  }
};

const submit = async () => {
  loading.value = true;
  try {
    const validForm = await formRef.value.validate();
    if (!validForm) return;

    const validUniqueName = await uniqueNameFormRef.value.validate();
    if (!validUniqueName) return;

    const _name = uniqueNameFormRef.value.data;
    const params = JSON.parse(JSON.stringify(form));
    params.name = JSON.stringify(_name);

    params.customer_type = (form.customer_type || []).map(({ customer_type_id, price }) => ({ customer_type_id, price }));
    params.buy_limit_products = (form.buy_limit_products || []).map(({ product_id, limit_num }) => ({ product_id, limit_num }));
    params.products = (select_list.value || []).map(({ product_id, is_show_cashier, is_show_tablet, is_show_kitchen, is_show_assistant }) => ({
      product_id,
      is_show_cashier,
      is_show_tablet,
      is_show_kitchen,
      is_show_assistant,
    }));

    if (form.is_time_limit == 1 && (form.time_limit <= form.remain_continue_time || form.time_limit <= form.remain_continue_notice_time)) {
      proxy.$ElMessage({
        message: proxy.$t('平板时间不能大于用餐时间'),
        type: 'warning',
      });
      return;
    }

    params.products.map((item) => {
      if (!is_open_tablet.value) {
        item.is_show_tablet = 2;
      }
      if (!is_open_assistant.value) {
        item.is_show_assistant = 2;
      }
      if (!is_open_kitchen_kds.value) {
        item.is_show_kitchen = 2;
      }
      return item;
    });

    try {
      if (props.editData) {
        params.buffet_id = params.id;
        await ProductApi.editBuffet(params, true);
      } else {
        await ProductApi.addBuffet(params, true);
      }
      proxy.$ElMessage({
        message: props.editData ? proxy.$t('编辑成功') : proxy.$t('添加成功'),
        type: 'success',
      });
      handleClose(true);
    } catch (err) {
      // 忽略错误
    }
  } catch (error) {
    scrollToError();
  } finally {
    loading.value = false;
  }
};

const scrollToError = () => {
  setTimeout(() => {
    const errorItems = document.querySelectorAll('.el-form-item__error');
    if (errorItems.length > 0) {
      const firstErrorItem = errorItems[0];
      firstErrorItem.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }, 200);
};

const selectList = (e) => {
  if (e == 'select') {
    selectType.value = e;
    multiple_selection.value = select_list.value;
  }
  if (e == 'limit') {
    selectType.value = e;
    multiple_selection.value = limit_list.value;
    limit_ids.value = form.product_ids.join(',');
  }
  open_product.value = true;
};

const closeDialogFunc = (e) => {
  open_product.value = e.openDialog;
  if (e.type == 'select') {
    e.data.map((item) => {
      if (!form.product_ids.includes(item.product_id)) {
        select_list.value.push({
          product_id: item.product_id,
          product_name_text: item.product_name_text,
          is_show_cashier: 1,
          is_show_kitchen: 1,
          is_show_tablet: 1,
          is_show_assistant: 1,
        });
      }
      return item;
    });

    form.product_ids = [];
    select_list.value.map((item) => {
      form.product_ids.push(item.product_id);
      return item;
    });

    formRef.value?.validateField('product_ids');

    limit_ids.value = form.product_ids.join(',');
  }
  if (e.type == 'limit') {
    const map = new Map();
    [limit_list.value, e.data].flat().forEach((obj) => map.set(obj.product_id, obj));
    limit_list.value = Array.from(map.values());

    const arr = [];
    (form.buy_limit_products || []).map((item) => {
      arr.push(item.product_id);
      return item;
    });

    limit_list.value.map((item) => {
      if (!arr.includes(item.product_id)) {
        form.buy_limit_products.push({
          name: item.product_name_text,
          product_id: item.product_id,
          limit_num: null,
        });
      }
      return item;
    });

    formRef.value?.validateField('buy_limit_products');
  }
};

const deleteOne = (index, product_id) => {
  select_list.value.splice(index, 1);
  form.product_ids = [];
  select_list.value.map((item) => {
    form.product_ids.push(item.product_id);
    return item;
  });
  limit_ids.value = form.product_ids.join(',');
  (form.buy_limit_products || []).map((item, i) => {
    if (product_id == item.product_id) {
      handleDelete(i);
    }
    return item;
  });
  formRef.value?.validateField('product_ids');
};

const handleDelete = (index) => {
  form.buy_limit_products.splice(index, 1);
  limit_list.value.splice(index, 1);
  formRef.value?.validateField('buy_limit_products');
};

const handleClose = (isSuccess = false, data) => {
  emit('closeDialog', {
    type: isSuccess ? 'success' : 'error',
    openDialog: false,
    data: data,
  });
};

const addCustomerType = () => {
  form.customer_type.push({
    customer_type_id: '',
    price: null,
  });
};

const handleDeleteCustomer = (index) => {
  form.customer_type.splice(index, 1);
};

const numChange = (index) => {
  nextTick(() => {
    form.customer_type[index].price = Number(proxy.$priceTwo(form.customer_type[index].price));
  });
};

const nameReturn = () => {
  if (is_open_tablet.value) {
    if (is_open_scan.value) {
      return proxy.$t('平板/扫码H5时间');
    }
    return proxy.$t('平板时间');
  } else {
    if (is_open_scan.value) {
      return proxy.$t('扫码H5时间');
    }
  }
  return '';
};

const returnType = (type) => {
  return type == '1' ? proxy.$t('堂食税类：') : proxy.$t('外带税类：');
};

const returnMessage = (type) => {
  return type == '1' ? proxy.$t('请选择堂食税类') : proxy.$t('请选择外带税类');
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

<template>
  <el-dialog
    class=""
    @close="handleClose(false)"
    v-model="dialogVisible"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :title="editData?.id ? $t('编辑方案') : $t('添加方案')"
  >
    <div class="form">
      <el-form size="small" ref="formRef" :model="form" label-position="top">
        <div class="common-form">{{ $t('基础配置') }}</div>
        <el-form-item
          for="no_click"
          :label="$t('方案名称')"
          prop="name"
          :rules="[
            { required: true, message: $t('请输入方案名称') },
            { validator: uniqueNameValidator('order_scheme', editData?.id, 'SINGLE'), trigger: 'blur' },
          ]"
        >
          <el-input v-model="form.name" :placeholder="$t('请输入方案名称')"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('使用渠道')" prop="use_channel" :rules="[{ required: true, message: $t('请选择使用渠道') }]">
          <el-checkbox-group v-model="form.use_channel" @change="handleUseChannelChange">
            <el-checkbox label="10" size="small">
              {{ $t('点餐方式') }}
            </el-checkbox>
            <el-checkbox label="20" size="small">
              {{ $t('桌台方式') }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item v-if="form.use_channel.indexOf('20') != -1" for="no_click" :label="$t('桌台区域')">
          <el-select v-model="form.table_area_ids" multiple :placeholder="$t('桌台区域')">
            <el-option v-for="item in area_list" :key="item.area_id" :label="item.area_name" :value="item.area_id"> </el-option>
          </el-select>
        </el-form-item>
        <el-form-item for="no_click" prop="must_type" :rules="[{ required: true, message: $t('请选择必点类型') }]">
          <template #label>
            <span>{{ $t('必点类型') }}</span>
            <el-tooltip class="item" effect="dark" placement="bottom">
              <template #content>
                {{ $t('设置每人必点：仅桌台方式支持设置，将按照开台人数计算；') }}<br />{{
                  $t('设置每笔订单必点：桌台方式下按照每个桌台计算，非桌台方式则按照每笔订单计算')
                }}</template
              >
              <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
            </el-tooltip>
          </template>
          <el-radio-group v-model="form.must_type">
            <el-radio :label="1" :disabled="disabledRadio" class="mb4">
              {{ $t('每人必点1份') }}
              <i class="tips">{{ $t('适合餐具、茶位费、蘸料等每人必点的场景') }}</i>
            </el-radio>
            <el-radio :label="2">
              {{ $t('每笔订单必点1份') }}
              <i class="tips">{{ $t('适合纸巾、火锅锅底等每单或每桌必点，不按人数算的场景') }}</i>
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item for="no_click" prop="must_rule" :rules="[{ required: true, message: '' }]">
          <template #label>
            <span>{{ $t('必点规则') }}</span>
            <el-tooltip class="item" effect="dark" placement="bottom">
              <template #content>
                {{ $t('固定商品：针对每种必点商品，每人或每笔订单需要选够您设置的必点数量；') }}<br />{{
                  $t('可选商品：在您设置的必点商品范围内，每人或每笔订单需要选够您设置的必点数量')
                }}</template
              >
              <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
            </el-tooltip>
          </template>
          <el-radio-group v-model="form.must_rule">
            <el-radio :label="1">{{ $t('固定商品') }}</el-radio>
            <el-radio :label="2">{{ $t('可选商品') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item
          for="no_click"
          :label="$t('必点商品')"
          prop="product_list"
          :rules="[
            {
              validator: () => {
                return product_list.length > 0 ? true : false;
              },
              required: true,
              message: $t('请选择商品'),
            },
          ]"
        >
          <template #label>
            {{ $t('必点商品') }}
            <span class="tips">{{ $t('如需设置餐具、茶位费、火锅底料等为必点，需先将其创建为商品') }}</span>
          </template>
          <div class="product-body">
            <div>
              <el-button size="small" type="primary" @click="addProduct" :loading="loading">
                {{ $t('选择商品') }}（{{ $t('已选商品') }}{{ product_list.length }}{{ $t('个') }}）
              </el-button>
            </div>
            <div class="product-list" v-if="product_list.length > 0">
              <template v-for="(item, index) in product_list" :key="index">
                <div class="product-item">
                  {{ item.product_name_text }}
                  <i class="el-icon el-tag__close" @click="removeProduct(index)">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">
                      <path
                        fill="currentColor"
                        d="M764.288 214.592 512 466.88 259.712 214.592a31.936 31.936 0 0 0-45.12 45.12L466.752 512 214.528 764.224a31.936 31.936 0 1 0 45.12 45.184L512 557.184l252.288 252.288a31.936 31.936 0 0 0 45.12-45.12L557.12 512.064l252.288-252.352a31.936 31.936 0 1 0-45.12-45.184z"
                      ></path>
                    </svg>
                  </i>
                </div>
              </template>
            </div>
          </div>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('状态')" prop="status" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.status">
            <el-radio :label="1">{{ $t('开启') }}</el-radio>
            <el-radio :label="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <div class="common-form">{{ $t('高级配置') }}</div>
        <el-form-item for="no_click" prop="auto_cart" :rules="[{ required: true, message: '' }]">
          <template #label>
            <span>{{ $t('自动加入购物车') }}</span>
            <el-tooltip class="item" effect="dark" placement="bottom">
              <template #content>
                {{ $t('固定商品：按所设置的必选要求将必点商品加入购物车') }}<br />{{ $t('可选商品：弹窗展示所有备选必点商品由服务员或顾客进行选择后加入购物车') }}</template
              >
              <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
            </el-tooltip>
          </template>
          <el-radio-group v-model="form.auto_cart">
            <el-radio :label="1">{{ $t('开启') }}</el-radio>
            <el-radio :label="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" prop="auto_change" :rules="[{ required: true, message: '' }]">
          <template #label>
            <span>{{ $t('顾客可修改必点数量') }}</span>
            <el-tooltip class="item" effect="dark" placement="bottom">
              <template #content> {{ $t('若可以修改，顾客可将数量改为“0”') }}</template>
              <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
            </el-tooltip>
          </template>
          <el-radio-group v-model="form.auto_change">
            <el-radio :label="1">{{ $t('开启') }}</el-radio>
            <el-radio :label="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" prop="auto_check" :rules="[{ required: true, message: '' }]">
          <template #label>
            <span>{{ $t('下单时检查必点商品') }}</span>
            <el-tooltip class="item" effect="dark" placement="bottom">
              <template #content> {{ $t('下单时检查购物车或已下单商品内是否有必点商品') }}</template>
              <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
            </el-tooltip>
          </template>
          <el-radio-group v-model="form.auto_check">
            <el-radio :label="1">{{ $t('开启') }}</el-radio>
            <el-radio :label="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" prop="auto_checkout" :rules="[{ required: true, message: '' }]">
          <template #label>
            <span>{{ $t('结账时检查必点商品') }}</span>
            <el-tooltip class="item" effect="dark" placement="bottom">
              <template #content> {{ $t('结账/免单时检查购物车或已下单商品内是否有必点商品') }}</template>
              <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
            </el-tooltip>
          </template>
          <el-radio-group v-model="form.auto_checkout">
            <el-radio :label="1">{{ $t('开启') }}</el-radio>
            <el-radio :label="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <div>
          <el-button @click="handleClose(false)">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
        </div>
      </div>
    </template>
  </el-dialog>
  <!-- 商品选择器 -->
  <ProductSelector
    v-if="openProductSelector"
    :open="openProductSelector"
    @close="handleProductSelectorClose"
    selectorType="all"
    type="product"
    num_type="1"
    :selectedProductIds="select_product_ids"
  >
  </ProductSelector>
</template>
<script setup>
  import { ref, computed, onMounted, getCurrentInstance, watch } from 'vue';
  import { uniqueNameValidator } from '@/utils/form.js';
  import SvgIcon from '@/components/svg-icon/SvgIcon.vue';
  import ProductSelector from '@/components/product/Selector.vue';
  import SettingApi from '@/api/setting.js';

  const { proxy } = getCurrentInstance();
  const emit = defineEmits(['close']);
  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
    editData: {
      type: Object,
      default: () => {},
    },
    area_list: {
      type: Array,
      default: () => [],
    },
  });

  const dialogVisible = computed(() => props.open);

  const loading = ref(false);
  const form = ref({
    name: '',
    use_channel: [],
    table_area_ids: [],
    must_type: null,
    must_rule: 1,
    product_ids: [],
    status: 1,
    auto_cart: 1,
    auto_change: 1,
    auto_check: 1,
    auto_checkout: 1,
  });

  const select_product_ids = ref([]);
  const formRef = ref(null);
  const product_list = ref([]);
  const openProductSelector = ref(false);
  const disabledRadio = ref(false);

  const removeProduct = (index) => {
    product_list.value.splice(index, 1);
  };

  const addProduct = () => {
    select_product_ids.value = [];
    if (product_list.value.length > 0) {
      product_list.value.map((item) => {
        select_product_ids.value.push(item.product_id);
      });
    }
    openProductSelector.value = true;
  };

  const handleProductSelectorClose = (list, categories) => {
    if (Array.isArray(list)) {
      product_list.value = list;
    }
    if (Array.isArray(categories)) {
      product_list.value.map((item) => {
        item.path_name_text = '';
        categories.map((item2) => {
          if (item.category_id == item2.category_id) {
            item.path_name_text = item2.path_name_text;
          }
          if (item2.child.length > 0) {
            item2.child.map((item3) => {
              if (item.category_id == item3.category_id) {
                item.path_name_text = item3.path_name_text;
              }
            });
          }
        });
      });
    }
    formRef.value.validateField('product_list');
    openProductSelector.value = false;
  };

  const handleUseChannelChange = (e) => {
    if (e.indexOf('20') != -1) {
      form.value.table_area_ids = [];
      if ((props.area_list.length > 0 && form.value.table_area_ids.length == 0) || !props.editData?.id) {
        props.area_list.map((item) => {
          form.value.table_area_ids.push(item.area_id);
        });
      }
    }
  };

  const handleClose = (e) => {
    emit('close', e);
  };
  const onSubmit = () => {
    formRef.value.validate((valid) => {
      if (valid) {
        let params = {};
        params = JSON.parse(JSON.stringify(form.value));
        params.product_ids = product_list.value.map((item) => item.product_id);
        params.table_area_ids.length == 0 ? (params.table_area_ids = '') : '';
        params.use_channel.indexOf('20') == -1 ? (params.table_area_ids = '') : '';
        loading.value = true;
        if (props.editData?.id) {
          SettingApi.orderSchemeEdit(params, true)
            .then((res) => {
              proxy.$ElMessage({
                type: 'success',
                message: $t('更新成功'),
              });
              handleClose(true);
            })
            .catch(() => {})
            .finally(() => {
              loading.value = false;
            });
        } else {
          SettingApi.orderSchemeAdd(params, true)
            .then((res) => {
              proxy.$ElMessage({
                type: 'success',
                message: $t('添加成功'),
              });
              handleClose(true);
            })
            .catch(() => {})
            .finally(() => {
              loading.value = false;
            });
        }
      }
    });
  };

  watch(
    () => form.value.use_channel,
    (val) => {
      if (val.indexOf('10') != -1) {
        form.value.must_type = 2;
        disabledRadio.value = true;
      } else {
        disabledRadio.value = false;
      }
    }
  );

  onMounted(() => {
    if (props.editData?.id) {
      form.value = {
        id: props.editData.id,
        name: props.editData.name,
        use_channel: props.editData.use_channel,
        table_area_ids: props.editData.table_area_ids.map((item) => Number(item)),
        must_type: props.editData.must_type,
        must_rule: props.editData.must_rule,
        product_ids: props.editData.product_ids,
        status: props.editData.status,
        auto_cart: props.editData.auto_cart,
        auto_change: props.editData.auto_change,
        auto_check: props.editData.auto_check,
        auto_checkout: props.editData.auto_checkout,
      };
      product_list.value = props.editData.product_ids;
    }
  });
</script>
<style lang="scss" scoped>
  .data-box-icon {
    color: var(--el-color-tips);
    margin-left: 4px;
  }

  .product-body {
    display: flex;
    flex-direction: column;
    width: 100%;
    overflow: auto;
    .product-list {
      display: flex;
      box-shadow: 0 0 0 1px var(--el-input-border-color, var(--el-border-color)) inset;
      border-radius: 4px;
      padding: 6px 11px;
      margin-top: 12px;
      gap: 8px;
      flex-wrap: wrap;
      .product-item {
        color: var(--el-tag-text-color);
        display: inline-flex;
        justify-content: center;
        align-items: center;
        vertical-align: middle;
        font-size: var(--el-tag-font-size);
        line-height: 1;
        border-width: 1px;
        border-style: solid;
        box-sizing: border-box;
        white-space: nowrap;
        padding: 0 7px;
        height: 20px;
        font-size: 12px;
        border-radius: 4px;
        background-color: var(--el-color-info-light-9);
        border: solid 1px var(--el-color-info-light-8);
      }
      .el-tag__close {
        font-size: 12px;
        margin-left: 4px;
        cursor: pointer;
      }
    }
  }
  :deep(.el-select--small .el-select__wrapper) {
    min-height: 32px !important;
    height: auto;
    padding: 4px 8px !important;
  }
  .mb4 {
    margin-bottom: 4px;
  }
</style>

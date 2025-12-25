<template>
  <div>
    <template v-for="(item, index) in languageList" :key="index">
      <el-form-item
        for="no_click"
        :label="$t('套餐单位：')"
        :rules="[{ required: true, message: $t('请填写套餐单位') }]"
        :prop="`model.product_unit.${item.key}`"
        v-if="item.name == languageKey"
      >
        <el-select
          v-model="form.model.product_unit[item.key]"
          @change="(e) => selectChange(e, index)"
          filterable
          clearable
          class="max-w460 mr8"
          size="default"
          :placeholder="$t('请选择') + `(${item.value})`"
          :disabled="erp_is_open == 1"
        >
          <template v-for="items in restaurantsObj[item.key]" :key="items.index">
            <el-option :value="items.index" :label="items.value"></el-option>
          </template>
        </el-select>
        <el-button size="small" type="primary" :disabled="erp_is_open == 1" class="el-icon-circle-plus" @click="addUnit">{{ $t('添加单位') }}+</el-button>
      </el-form-item>
    </template>

    <div class="common-form mt50">{{ $t('套餐商品') }}</div>

    <div class="package-group-list">
      <el-card class="package-card" shadow="never" v-for="(group, groupIndex) in form.model.package_group" :key="groupIndex">
        <template v-for="(item, index) in languageList" :key="index">
          <el-form-item
            label-position="left"
            label-width="auto"
            class="package-group-item"
            for="no_click"
            :label="$t('分组名称') + `(${item.value})`"
            :rules="[{ required: true, validator: (rule, value, callback) => validateGroupName(rule, value, callback, groupIndex) }]"
            :prop="`model.package_group.${groupIndex}.group_name.${item.key}`"
            v-if="item.name == languageKey"
          >
            <el-input
              type="text"
              :placeholder="$t('请输入分组名称')"
              :maxlength="150"
              class="max-w460"
              :disabled="erp_is_open == 1"
              v-model="form.model.package_group[groupIndex].group_name[item.key]"
            ></el-input>
            <el-button class="ml8" type="primary" @click="translate(groupIndex)">{{ $t('翻译') }}</el-button>
            <el-select v-model="form.model.package_group[groupIndex].group_type" class="max-w230 ml16" :disabled="erp_is_open == 1">
              <el-option :value="0" :label="$t('固定')"></el-option>
              <el-option :value="1" :label="$t('可选')"></el-option>
            </el-select>

            <el-icon class="delete-icon" :class="{ 'is-disabled': form.model.package_group.length === 1 }" @click="handleDelete(groupIndex)">
              <Delete />
            </el-icon>
          </el-form-item>
        </template>

        <el-form-item
          for="no_click"
          :label="$t('本组可选数量') + '&nbsp;&nbsp;&nbsp;' + form.model.package_group[groupIndex].product_list.length + $t('选')"
          label-width="auto"
          label-position="left"
          v-if="form.model.package_group[groupIndex].group_type == 1"
        >
        <numInput class="max-w80" v-model="form.model.package_group[groupIndex].optional_min_count" :min="0" :max="99999999" :precision="0" :disabled="erp_is_open == 1" />
        <span>-</span>
        <numInput class="max-w80" v-model="form.model.package_group[groupIndex].optional_count" :min="0" :max="99999999" :precision="0" :disabled="erp_is_open == 1" />
        </el-form-item>

        <el-form-item :prop="`model.package_group.${groupIndex}.product_list`" :rules="[{ required: true, validator: validatePackageGroup, message: $t('请添加套餐商品') }]">
          <el-table :data="form.model.package_group[groupIndex].product_list" style="width: 100%" border max-height="250">
            <el-table-column :label="$t('序号')" width="80" type="index" />
            <el-table-column prop="product_name_text" :label="$t('商品名称')" />
            <el-table-column prop="spec_name_text" :label="$t('规格')" width="120" />
            <el-table-column prop="product_price" :label="$t('加价')" width="120">
              <template #default="scope">
                <el-form-item
                  class="mt16"
                  :prop="`model.package_group.${groupIndex}.product_list.${scope.$index}.add_price`"
                  :rules="[{ required: true, message: $t('请输入加价') }]"
                >
                  <numInput v-model="scope.row.add_price" :min="0" :max="99999999" :precision="2" :disabled="erp_is_open == 1" />
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="num" :label="$t('数量')" width="120">
              <template #default="scope">
                <el-form-item class="mt16" :prop="`model.package_group.${groupIndex}.product_list.${scope.$index}.num`" :rules="[{ required: true, message: $t('请输入数量') }]">
                  <numInput v-model="scope.row.num" :min="0" :max="999" :precision="0" :disabled="erp_is_open == 1" />
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="sort" :label="$t('排序')" width="120">
              <template #default="scope">
                <el-form-item class="mt16" :prop="`model.package_group.${groupIndex}.product_list.${scope.$index}.sort`" :rules="[{ required: true, message: $t('请输入排序') }]">
                  <numInput v-model="scope.row.sort" :min="0" :max="999" :precision="0" @change="handleSortChange(groupIndex)" :disabled="erp_is_open == 1" />
                </el-form-item>
              </template>
            </el-table-column>
            <!-- <el-table-column prop="stock_num" :label="$t('商品库存')" width="160">
              <template #default="scope">
                {{ scope.row.stock_num }}
              </template>
            </el-table-column> -->

            <el-table-column prop="is_required" :label="$t('必选')" width="80" v-if="form.model.package_group[groupIndex].group_type == 1">
              <template #default="scope">
                <el-checkbox v-model="scope.row.is_required" :true-value="1" :false-value="0" :disabled="erp_is_open == 1"></el-checkbox>
              </template>
            </el-table-column>
            <el-table-column prop="is_default" :label="$t('默认')" width="80" v-if="form.model.package_group[groupIndex].group_type == 1">
              <template #default="scope">
                <el-checkbox v-model="scope.row.is_default" :true-value="1" :false-value="0" :disabled="erp_is_open == 1"></el-checkbox>
              </template>
            </el-table-column>

            <el-table-column prop="action" :label="$t('操作')" width="100" fixed="right">
              <template #default="scope">
                <el-button type="primary" @click="handleDeleteGoods(groupIndex, scope.$index)" :disabled="erp_is_open == 1">{{ $t('删除') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-form-item>
        <div class="mt16">
          <el-button class="w-full" type="default" @click="addGoods(groupIndex)" :disabled="erp_is_open == 1">{{ $t('添加商品') }}</el-button>
        </div>
      </el-card>
    </div>

    <el-button :disabled="form.model.package_group.length >= 100" class="mt16" type="primary" @click="addGroup">{{ $t('添加分组') }}</el-button>

    <div class="common-form mt50">{{ $t('库存') }}</div>

    <!--  <el-form-item for="no_click" :label="$t('可售量')">
      <numInput type="text" :placeholder="$t('请输入库存')" disabled v-model="nowStock" :maxlength="50" class="max-w460"></numInput>
      <div class="gray9">{{ $t('根据子商品库存动态计算，库存变化自动更新') }}</div>
    </el-form-item> -->
    <!-- TODO: 是否开启库存,2025年08月07日09:55:45，需求暂时隐藏 -->
    <!-- <el-form-item for="no_click" :label="$t('是否开启库存')">
      <el-radio-group v-model="form.model.is_open_stock">
        <el-radio :value="1">{{ $t('是') }}</el-radio>
        <el-radio :value="0">{{ $t('否') }}</el-radio>
      </el-radio-group>
    </el-form-item> -->

    <!-- <el-form-item v-if="form.model.is_open_stock" for="no_click" :prop="`model.package_stock`" :label="$t('套餐可售总量')" :rules="[{ required: true, message: $t('请输入库存') }]">
      <numInput type="text" :placeholder="$t('请输入库存')" :precision="0" v-model="form.model.package_stock" :min="0" :max="99999999" class="max-w460"></numInput>
      <div class="gray9">{{ $t('库存为0时套餐自动售罄') }}</div>
    </el-form-item> -->

    <el-form-item for="no_click" :label="$t('库存计算方式：')">
      <el-radio-group v-model="form.model.deduct_stock_type" :disabled="erp_is_open == 1">
        <el-radio :value="10">{{ $t('下单减库存') }}</el-radio>
        <el-radio :value="20">{{ $t('付款减库存') }}</el-radio>
      </el-radio-group>
    </el-form-item>

    <el-dialog v-model="dialogVisible" :title="$t('翻译')" :close-on-click-modal="false">
      <UniqueNameForm
        v-if="dialogVisible"
        ref="uniqueNameFormRef"
        :labelPrefix="$t('分组名称')"
        :isUnique="false"
        apiSource="product"
        :overrideLanguages="overrideLanguages"
        :otherGroupNames="otherGroupNames"
        :maxlength="150"
      />
      <template #footer>
        <div class="flex justify-end">
          <el-button @click="dialogVisible = false">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="translateConfirm">{{ $t('确定') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 商品选择器 -->
    <ProductSelector
      v-if="openProductSelector"
      :open="openProductSelector"
      @close="handleProductSelectorClose"
      selectorType="all"
      :selectedProductIds="selectedProductIds"
      :haveSku="true"
      :haveStatusClose="true"
    >
    </ProductSelector>

    <!--添加-->
    <Add v-if="open_add_feed" :open_add="open_add_feed" @closeDialog="closeDialogFunc($event, 'add')"></Add>
  </div>
</template>
<script setup>
  import { ref, inject, nextTick, computed, defineEmits, reactive, watch } from 'vue';
  import { languageStore } from '@/store/model/language.js';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';
  import ProductSelector from '@/components/product/Selector.vue';
  import Add from '../../../expand/unit/add.vue';
  import { useUserStore } from '@/store';
  import { length } from 'localforage';
  const { erp_is_open } = useUserStore();
  const emit = defineEmits(['validateField']);
  const languageList = languageStore().getLanguageList().languageList.value;
  const languageKey = languageStore().getLanguageKey().language.value;
  const languageData = JSON.stringify(languageStore().getLanguageKeyForm());

  // 初始化语言对象
  let languageObj = {};
  languageList.forEach((item) => {
    languageObj[item.key] = [];
  });

  const restaurantsObj = reactive(languageObj);

  const form = inject('form');

  const dialogVisible = ref(false);

  const uniqueNameFormRef = ref(null);

  const openProductSelector = ref(false);

  const selectedProductIds = ref([]);

  const selectIndex = ref(0);

  const overrideLanguages = ref({});

  const otherGroupNames = ref([]);

  const open_add_feed = ref(false);

  const nowStock = computed(() => {
    // 计算规则：相同商品在不同分组中的数量先合并相加，然后用该商品库存除以合并后的需求数量，最后取所有商品中的最小值，向下取整

    // 如果没有商品组或者商品组为空，返回0
    if (!form.model.package_group || form.model.package_group.length === 0) {
      return 0;
    }

    // 汇总相同商品的需求数量与库存
    const productDemandMap = new Map(); // key: product_id, value: { needTotal, stock }

    form.model.package_group.forEach((group) => {
      if (!group.product_list || group.product_list.length === 0) return;
      group.product_list.forEach((product) => {
        // 忽略数量为0或空的商品
        if (!product.num || product.num <= 0) return;

        const productId = product.product_id;
        if (productId === null || productId === undefined) return;

        const prev = productDemandMap.get(productId) || { needTotal: 0, stock: undefined };
        const needTotal = prev.needTotal + Number(product.num || 0);

        // 记录库存（取最小库存以更为保守，避免不同条目库存不一致导致的高估）
        let stock = prev.stock;
        if (product.stock_num !== null && product.stock_num !== undefined && product.stock_num >= 0) {
          stock = stock === undefined ? Number(product.stock_num) : Math.min(Number(stock), Number(product.stock_num));
        }

        productDemandMap.set(productId, { needTotal, stock });
      });
    });

    // 如果没有有效的合并数据，返回0
    if (productDemandMap.size === 0) {
      return 0;
    }

    // 计算每个商品的可售套餐数，取最小值
    let minStock = Infinity;
    let hasValidProduct = false;

    for (const [, info] of productDemandMap.entries()) {
      const need = Number(info.needTotal || 0);
      // 需求必须大于0
      if (need <= 0) continue;
      // 库存必须为有效数字
      if (info.stock === undefined || info.stock === null || isNaN(Number(info.stock)) || Number(info.stock) < 0) continue;

      const stockRatio = Number(info.stock) / need;
      minStock = Math.min(minStock, stockRatio);
      hasValidProduct = true;
    }

    // 如果没有任何有效商品可参与计算，返回0
    if (!hasValidProduct || minStock === Infinity) {
      return 0;
    }

    // 向下取整
    return Math.floor(minStock);
  });

  const addUnit = () => {
    open_add_feed.value = true;
  };

  /*关闭弹窗*/
  const closeDialogFunc = (e, f) => {
    if (f == 'add') {
      open_add_feed.value = e.openDialog;
      if (e.type == 'success' && e.data) {
        //
        form.unit.unshift(e.data);
      }
    }
  };

  // 方法定义
  const selectChange = (e) => {
    languageList.forEach((item) => {
      form.model.product_unit[item.key] = restaurantsObj[item.key][e]?.value || '';
      form.model.unit_id = restaurantsObj[item.key][e]?.unit_id || '';
    });
  };

  // 翻译
  const translate = async (groupIndex) => {
    if (erp_is_open == 1) {
      return;
    }
    // 设置当前分组名称
    overrideLanguages.value = form.model.package_group[groupIndex].group_name;
    // 保存当前选中的分组索引
    selectIndex.value = groupIndex;
    // 打开翻译弹窗
    dialogVisible.value = true;
    await nextTick();

    // 把除了自己以外的其他分组名称获取到
    otherGroupNames.value = form.model.package_group.filter((item, index) => index !== groupIndex).map((item) => item.group_name);

    // 校验表单
    uniqueNameFormRef.value.validate();
  };

  // 翻译确认
  const translateConfirm = async () => {
    // 校验是否与其他分组名称重复
    const isRepeat = await uniqueNameFormRef.value.validate();
    if (!isRepeat) {
      return;
    }
    // 将翻译结果回写到原数据
    Object.assign(form.model.package_group[selectIndex.value].group_name, uniqueNameFormRef.value.data);
    dialogVisible.value = false;

    // 校验分组名称
    emit('validateField', `model.package_group.${selectIndex.value}.group_name.${languageKey}`);
  };

  const addGroup = () => {
    form.model.package_group.push({
      group_name: JSON.parse(languageData),
      group_type: 0,
      optional_min_count: 0,
      optional_count: 1,
      product_list: [],
    });
  };

  // 添加商品
  const addGoods = (groupIndex) => {
    openProductSelector.value = true;
    selectedProductIds.value = [];
    selectIndex.value = groupIndex;
    form.model.package_group[groupIndex].product_list.map((item) => {
      selectedProductIds.value.push(item.product_id);
    });
  };

  // 删除商品
  const handleDeleteGoods = (groupIndex, index) => {
    form.model.package_group[groupIndex].product_list.splice(index, 1);
  };

  // 删除分组
  const handleDelete = (index) => {
    if (erp_is_open == 1) {
      return;
    }
    if (form.model.package_group.length === 1) {
      return;
    }
    form.model.package_group.splice(index, 1);
  };

  // 商品选择器关闭
  const handleProductSelectorClose = async (selectedProducts) => {
    openProductSelector.value = false;
    if (!selectedProducts) {
      return;
    }
    // 先与当前分组商品对比，获取需要删除的商品id组和需要新增的商品id组
    const deleteProductIds = [];
    const addProductIds = [];

    // 获取所选商品的ID列表
    const selectedProductIds = (selectedProducts || []).map((product) => (typeof product === 'object' ? product.product_id : product));

    // 找出需要删除的商品ID
    (form.model.package_group[selectIndex.value].product_list || []).map((item) => {
      if (!selectedProductIds.includes(item.product_id)) {
        deleteProductIds.push(item.product_id);
      }
    });

    // 找出需要新增的商品
    const existingProductIds = (form.model.package_group[selectIndex.value].product_list || []).map((p) => p.product_id);
    (selectedProducts || []).map((item) => {
      const productId = typeof item === 'object' ? item.product_id : item;
      if (!existingProductIds.includes(productId)) {
        addProductIds.push(item);
      }
    });

    // 先删除商品
    form.model.package_group[selectIndex.value].product_list = form.model.package_group[selectIndex.value].product_list.filter(
      (item) => !deleteProductIds.includes(item.product_id)
    );

    // 等待下一帧，确保DOM更新完成
    await nextTick();

    // 获取当前分组商品最大排序，处理空数组情况
    const currentProductList = form.model.package_group[selectIndex.value].product_list || [];
    let maxSort = 0;

    if (currentProductList.length > 0) {
      const sortValues = currentProductList.map((item) => item.sort || 0).filter((sort) => !isNaN(sort));

      if (sortValues.length > 0) {
        maxSort = Math.max(...sortValues);
      }
    }

    // 再插入新增商品（创建新对象，避免修改原始数据）
    addProductIds.forEach((item, index) => {
      const newProduct = {
        ...item, // 复制所有原有属性
        sort: maxSort + index + 1, // 递增排序
        num: 1, // 保持原有num值，如果没有则为null
        add_price: 0, // 加价
        is_required: 1, // 是否必选  0-否 1-是
        is_default: 1, // 是否默认 0-否 1-是
      };
      form.model.package_group[selectIndex.value].product_list.push(newProduct);
    });
  };

  // 排序改变
  const handleSortChange = (groupIndex) => {
    // 排序改变后，需要重新排序
    form.model.package_group[groupIndex].product_list.sort((a, b) => a.sort - b.sort);
  };

  // 校验套餐商品
  const validatePackageGroup = (rule, value, callback) => {
    if (value.length === 0) {
      callback(new Error($t('请添加套餐商品')));
    } else {
      callback();
    }
  };

  // 校验分组名称
  const validateGroupName = (rule, value, callback, groupIndex) => {
    // 检查当前输入框的值是否为空
    if (!value) {
      callback(new Error($t('请填写分组名称')));
      return;
    }
    // 以 languageList 中的key值为基准，检查该分组下所有语言版本是否都已填写
    const groupName = form.model.package_group[groupIndex].group_name;
    const allLanguagesFilled = languageList.every((item) => {
      const langValue = groupName[item.key];
      return langValue && langValue.trim();
    });

    if (!allLanguagesFilled) {
      callback(new Error($t('请翻译分组名称')));
    } else {
      callback();
    }
  };

  // 工具函数定义
  const isValidJSON = (str) => {
    try {
      JSON.parse(str);
      return true; // 如果解析成功，返回 true
    } catch (e) {
      return false; // 如果解析失败，返回 false
    }
  };

  watch(
    () => form.unit,
    (val) => {
      val.map((item, index) => {
        let unit_name = isValidJSON(item.unit_name) ? JSON.parse(item.unit_name) : {};
        languageList.forEach((items) => {
          if (unit_name[items.key] != null) {
            restaurantsObj[items.key].push({
              value: unit_name[items.key] == '' ? '-' : unit_name[items.key],
              index: index,
              unit_id: item.unit_id,
            });
          }
        });
      });
    },
    { deep: true, immediate: true }
  );
</script>
<style scoped lang="scss">
  .flex {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .justify-end {
    justify-content: flex-end;
  }

  .w-full {
    width: 100%;
  }
  .package-group-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .package-card {
    width: 100%;
  }
  .delete-icon {
    cursor: pointer;
    font-size: 24px;
    margin-left: auto;
  }
  .is-disabled {
    cursor: not-allowed;
    color: #c0c4cc;
  }
  .package-group-item {
    :deep(.el-form-item__label) {
      margin-bottom: 0;
    }
  }

  .max-w230 {
    max-width: 230px;
  }

  .max-w80 {
    max-width: 80px;
  }

  .ml16 {
    margin-left: 16px;
  }
</style>

<template>
  <div class="product-add">
    <!--form表单-->
    <el-form size="small" ref="formRef" class="product-form" :model="form" label-position="top" label-width="180px">
      <!--基础信息-->
      <div class="product-form-flex" ref="formContainer">
        <Basic ref="BasicRef" @validateField="validateField"></Basic>

        <!--规格设置-->
        <Spec v-if="form.model.type != 30"></Spec>

        <!--套餐设置-->
        <Package v-if="form.model.type == 30" @validateField="validateField"></Package>

        <!-- 属性设置-->
        <Attr ref="AttrRef" v-if="form.model.type == 10" @validateField="validateField"></Attr>

        <!-- 加料设置-->
        <Ingredients ref="IngredientsRef" v-if="form.model.type == 10" @validateField="validateField"></Ingredients>

        <!--高级设置-->
        <Buyset></Buyset>
      </div>
      <!--提交-->
      <div class="common-button-wrapper">
        <el-button size="small" @click="() => cancelFunc(1)">{{ $t('取消') }}</el-button>
        <el-tooltip effect="dark" placement="top" :content="$t('如长时间无响应，请刷新后重试。')">
          <el-button size="small" type="primary" @click="() => onSubmit(1)" :loading="loading">{{ $t('确定') }}</el-button>
        </el-tooltip>
        <el-button size="small" type="primary" @click="() => onSubmit(2)" :loading="loading">{{ $t('确定后继续添加') }}</el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup>
  import { ref, reactive, provide, onMounted, getCurrentInstance } from 'vue';
  import { useRouter } from 'vue-router';
  import { ElMessage } from 'element-plus';
  import ProductApi from '@/api/product.js';
  import Basic from './part/Basic.vue';
  import Attr from './part/Attr.vue';
  import Ingredients from './part/Ingredients.vue';
  import Spec from './part/Spec.vue';
  import Package from './part/Package.vue';
  import Buyset from './part/Buyset.vue';
  import { languageStore } from '@/store/model/language.js';
  import { EEUIRELOAD } from '@/utils/platform.js';
  import { useUserStore } from '@/store/index';

  // 获取当前实例
  const { proxy } = getCurrentInstance();
  const router = useRouter();

  // 获取用户信息和语言数据
  const languageData = JSON.stringify(languageStore().getLanguageKeyForm());
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;

  // 响应式数据
  const app_id_ref = ref(app_id);
  const loading = ref(false);
  const pageParams = ref({});

  // 表单数据
  const form = reactive({
    model: {
      type: 10,
      product_name: JSON.parse(languageData),
      category_id: null,
      erp_supplier_id: null,
      image: [],
      img_name: '',
      selling_point: '',
      spec_type: 20,
      deduct_stock_type: 10,
      num_type: 0,
      is_alone_grade: 0,
      sku: [
        {
          spec_name: JSON.parse(languageData),
          product_price: null,
          stock_num: null,
          product_weight: '',
          cost_price: 0,
          material: [],
          spec_id: null,
          barcode: '',
          barcodeUniqueness: true,
        },
      ],
      product_attr: [],
      product_feed: [],
      feed_required: 0,
      feed_open_max_select: 0,
      feed_max_select: 0,
      min_buy: 1,
      product_unit: JSON.parse(languageData),
      unit_id: '',
      content: '',
      product_status: 10,
      material: [],
      is_show_cashier: 1,
      is_show_tablet: 1,
      is_show_kitchen: 1,
      is_show_assistant: 1,
      is_show_h5: 1,
      is_show_delivery: 1,
      sales_initial: 0,
      product_sort: 0,
      limit_num: 0,
      special_id: '',
      is_points_gift: 1,
      is_agent: 0,
      is_ind_agent: 0,
      agent_money_type: 10,
      first_money: 0,
      second_money: 0,
      third_money: 0,
      is_enable_grade: 1,
      alone_grade_type: 10,
      label_id: '',
      productTaxes: [
        {
          product_tax_type: '1',
          tax_category_id: '',
        },
        {
          product_tax_type: '2',
          tax_category_id: '',
        },
      ],
      open_overall_discount: 1,
      package_price: null,
      is_open_stock: 1,
      package_stock: null,
      package_group: [
        {
          group_name: JSON.parse(languageData),
          product_list: [],
        },
      ],
      product_printer_uuids: [],
    },
    category: [],
    feed: [],
    attribute: [],
    unit: [],
    spec: [],
    labelList: [],
    special: [],
    delivery: [],
    gradeList: [],
    specData: null,
    basicSetting: {},
    agentSetting: {},
    single_select_list: [],
    many_select_list: [[]],
    ing_select_list: [[]],
    productPrinterList: [],
  });

  // 模板引用
  const formRef = ref(null);
  const BasicRef = ref(null);
  const AttrRef = ref(null);
  const IngredientsRef = ref(null);

  // 提供form给子组件
  provide('form', form);

  // 组件挂载时初始化
  onMounted(() => {
    pageParams.value = JSON.parse(JSON.stringify(languageStore().getPageParams().pageParams.value));
    languageStore().setPageParams({});
    getBaseData();
  });

  // 方法定义
  const getBaseData = async () => {
    try {
      const res = await ProductApi.storeGetBaseData({}, true);
      loading.value = false;
      Object.assign(form, res.data);
    } catch (error) {
      loading.value = false;
    }
  };

  /*转JSON字符串*/
  const convertJson = (list) => {
    let obj = {};
    list.forEach((item) => {
      obj[item.grade_id] = item.product_equity;
    });
    return JSON.stringify(obj);
  };

  const validateField = (e) => {
    formRef.value.validateField(e);
  };

  /*提交*/
  const onSubmit = async (e) => {
    loading.value = true;
    let valid = false;

    await formRef.value.validate((res) => {
      valid = res;
    });

    // 验证表单
    const validUniqueName = await BasicRef.value.$refs.uniqueNameFormRef.validate();

    // 如果表单验证通过，则处理参数
    if (validUniqueName && valid) {
      let params = {};
      params = JSON.parse(JSON.stringify(form.model));

      // 将产品名称和产品单位转换为字符串
      const _name = BasicRef.value.$refs.uniqueNameFormRef.data;
      params.product_name = JSON.stringify(_name);

      params.product_unit = JSON.stringify(params.product_unit);

      params.spec_type = 20; //规格类型固定20
      // 处理sku中的规格
      params.sku.map((item, index) => {
        params.sku[index].spec_name = JSON.stringify(item.spec_name);
      });

      // 如果分类id为对象且不为空，则处理分类id
      if (typeof params.category_id == 'object' && params.category_id) {
        params.category_id = Number(params.category_id[params.category_id.length - 1]);
      }

      // 材料的数据添加处理
      if (params.type == 20) {
        let data = {};
        data = {
          type: params.type,
          product_name: params.product_name,
          category_id: params.category_id,
          image: params.image,
          img_name: params.img_name,
          selling_point: params.selling_point,
          erp_supplier_id: params.erp_supplier_id,
          product_unit: params.product_unit,
          unit_id: params.unit_id,
          spec_type: 10,
          sku: params.sku,
          product_status: params.product_status,
          product_sort: params.product_sort,
        };
        params = data;
      }

      // 如果套餐类型，则处理套餐数据
      if (params.type == 30) {
        params.package_group.forEach((group) => {
          group.group_name = JSON.stringify(group.group_name);
          // group.product_list 只需要保留product_id、num、sort，其他字段删除
          let productList = [];
          group.product_list.forEach((product) => {
            productList.push({
              product_id: product.product_id,
              num: product.num,
              sort: product.sort,
            });
          });
          group.product_list = productList;
        });
        // 删除sku
        params.sku = [];
        // 套餐类型不显示外送
        params.is_show_delivery = 0;
      }

      // 将等级列表转换为json
      params.alone_grade_equity = convertJson(form.gradeList);

      // 调用接口添加产品
      try {
        await ProductApi.storeAddProduct(
          {
            params: JSON.stringify(params),
          },
          true
        );

        loading.value = false;
        ElMessage({
          message: proxy.$t('添加成功'),
          type: 'success',
        });
        cancelFunc(e);
      } catch (res) {
        loading.value = false;
        //验证规格的条形码
        if ((res.data || []).length > 0 && typeof res.data == 'Array') {
          await res.data.map((item, index) => {
            form.model.sku[index].barcodeUniqueness = item;
          });
        }
        await formRef.value.validate(() => {
          moveToError();
        });
      }
    } else {
      loading.value = false;
      // 如果表单验证不通过，则滚动到错误位置
      moveToError();
    }
  };

  const moveToError = () => {
    // 设置一个定时器，在200毫秒后，获取所有带有el-form-item__error类的元素，并将第一个元素滚动到视图中
    setTimeout(() => {
      const errorItems = document.querySelectorAll('.el-form-item__error');
      if (errorItems.length > 0) {
        const firstErrorItem = errorItems[0];
        firstErrorItem.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    }, 200);
  };

  /*保存为草稿*/
  const Draft = () => {
    form.model.product_status = 30;
    onSubmit();
  };

  /*取消*/
  const cancelFunc = (e) => {
    if (e == 1) {
      languageStore().setPageParams(pageParams.value);
      router.push('/' + app_id_ref.value + '/product/store/index');
    } else {
      languageStore().setPageParams(pageParams.value);
      EEUIRELOAD();
    }
  };
</script>

<style lang="scss" scoped>
  .basic-setting-content {
    //
  }

  .product-add {
    height: calc(100% - 14px);
    overflow: hidden;
  }

  .product-form {
    height: 100%;
    overflow: hidden;
    display: flex;
    flex-direction: column;

    .product-form-flex {
      flex: 1 1 auto;
      overflow-y: auto;
    }

    .common-button-wrapper {
      flex: 0 0 auto;
      flex-shrink: 0;
    }

    :deep(.el-select__placeholder) {
      font-size: 14px;
    }
  }
</style>

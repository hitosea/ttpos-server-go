<template>
  <div class="product-add" v-loading="loading">
    <!--form表单-->
    <el-form size="small" ref="formRef" :model="form" class="product-form" label-position="top" label-width="180px" v-if="!loading">
      <!--基础信息-->
      <div class="product-form-flex">
        <Basic ref="BasicRef" @validateField="validateField" disableChangeType></Basic>

        <!--规格设置-->
        <Spec v-if="form.model.type != 30"></Spec>

        <!--套餐设置-->
        <Package v-if="form.model.type == 30" @validateField="validateField"></Package>

        <!-- 属性设置-->
        <Attr ref="AttrRef" v-if="form.model.type == 10" @validateField="validateField"></Attr>

        <!-- 加料设置-->
        <Ingredients ref="IngredientsRef" v-if="form.model.type == 10" @validateField="validateField"></Ingredients>

        <!--高级设置-->
        <Buyset ref="BuysetRef" ></Buyset>
      </div>
      <!--提交-->
      <div class="common-button-wrapper">
        <el-button size="small" @click="() => cancelFunc()">{{ $t('取消') }}</el-button>
        <el-tooltip effect="dark" placement="top" :content="$t('如长时间无响应，请刷新后重试。')">
          <el-button size="small" type="primary" @click="() => onSubmit()" :loading="save_loading">{{ $t('确定') }}</el-button>
        </el-tooltip>
      </div>
    </el-form>
    <!-- 调整库存弹窗 2025年12月12日13:49:13 任务37468 -->
    <!-- <el-dialog v-if="dialogVisible" v-model="dialogVisible" :title="$t('调整库存')" width="700" align-center append-to-body>
      <el-form size="small" :inline="true" ref="tiaoRef" :model="tiao" label-position="top">
        <el-table size="small" ref="multipleTable" :data="form.model.sku" border style="width: 100%" v-loading="loading">
          <el-table-column prop="product.type" width="300" :label="$t('规格名称')" v-if="form.model.type == '10'">
            <template #default="scope">
              {{ scope.row.spec_name_text || scope.row.spec_name[languageKey] || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="product.type" :label="$t('原库存')">
            <template #default="scope">
              {{ form.model.type == '10' ? oldForm.model?.sku[scope.$index]?.stock_num || 0 : oldForm.model?.sku[scope.$index]?.material_stock || 0 }}
            </template>
          </el-table-column>
          <el-table-column prop="product.type" :label="$t('调整后库存')">
            <template #default="scope">
              {{ form.model.type == '10' ? scope.row.stock_num || 0 : scope.row.material_stock || 0 }}
            </template>
          </el-table-column>
        </el-table>
        <el-form-item style="width: 100%; margin-top: 12px" :label="$t('备注')" :rules="[{ required: true, message: $t('请输入调整备注') }]" prop="stock_remark">
          <el-input size="small" v-model="tiao.stock_remark" :placeholder="$t('请输入调整备注')"></el-input>
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button type="primary" @click="() => onSubmit(1)" :loading="save_loading"> {{ $t('确定') }}</el-button>
        </div>
      </template>
    </el-dialog> -->
  </div>
</template>

<script setup>
  import { ref, reactive, provide, onMounted, getCurrentInstance, nextTick } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { ElMessage } from 'element-plus';
  import ProductApi from '@/api/product.js';
  import Basic from './part/Basic.vue';
  import Attr from './part/Attr.vue';
  import Ingredients from './part/Ingredients.vue';
  import Spec from './part/Spec.vue';
  import Package from './part/Package.vue';
  import Buyset from './part/Buyset.vue';
  import { languageStore } from '@/store/model/language.js';
  import { useUserStore } from '@/store/index';
  import { formatModel } from '@/utils/base.js';

  // 获取当前实例
  const { proxy } = getCurrentInstance();
  const router = useRouter();
  const route = useRoute();

  // 获取用户信息和语言数据
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;
  const languageData = JSON.stringify(languageStore().getLanguageKeyForm());
  const languageList = languageStore().getLanguageList().languageList.value;
  let language = languageStore().language;
  let languageKey = '0';
  languageList.map((item) => {
    if (item.name == language) {
      languageKey = item.key;
    }
  });

  // 响应式数据
  const app_id_ref = ref(app_id);
  const old_audit = ref(20);
  const product_id = ref(0);
  const scene = ref('edit');
  const loading = ref(true);
  const save_loading = ref(false);
  const stockNumChange = ref(false);
//   const dialogVisible = ref(false);
  const languageKey_ref = ref(languageKey);
  const pageParams = ref({});

  // 表单数据
  const form = reactive({
    model: {},
    category: [],
    feed: [],
    attribute: [],
    unit: [],
    spec: [],
    labelList: [],
    special: [],
    gradeList: [],
    specData: null,
    isSpecLocked: false,
    basicSetting: {},
    agentSetting: {},
    label_id: '',
    single_select_list: [],
    many_select_list: [[]],
    ing_select_list: [[]],
    productPrinterList: [],
  });

  // 模型数据
  const model = reactive({
    type: 10,
    product_name: '',
    category_id: null,
    image: [],
    img_name: '',
    selling_point: '',
    erp_supplier_id: '',
    spec_type: 20,
    deduct_stock_type: 10,
    num_type: 0,
    is_alone_grade: 0,
    sku: [
      {
        product_price: '',
        stock_num: '',
        product_weight: '',
        cost_price: '',
        material: [],
      },
    ],
    stock_remark: '',
    product_attr: [],
    product_feed: [],
    feed_required: 0,
    feed_open_max_select: 0,
    feed_max_select: 0,
    min_buy: 1,
    product_unit: '',
    unit_id: '',
    content: '',
    product_status: 10,
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
    open_overall_discount: 0,
    package_price: null,
    is_open_stock: 1,
    package_stock: null,
    package_group: [
      {
        group_name: JSON.parse(languageData),
        group_type: 0, // 0-固定套餐 1-可选套餐
        optional_count: 1, // 可选套餐数量
        product_list: [],
      },
    ],
    product_printer_uuids: [],
  });

  const oldForm = ref({});
  const tiao = reactive({
    stock_remark: '',
  });

  // 模板引用
  const formRef = ref(null);
  const BasicRef = ref(null);
  const AttrRef = ref(null);
  const IngredientsRef = ref(null);
  const tiaoRef = ref(null);
  const BuysetRef = ref(null);
  // 提供form给子组件
  provide('form', form);

  // 组件挂载时初始化
  onMounted(() => {
    pageParams.value = JSON.parse(JSON.stringify(languageStore().getPageParams().pageParams.value));
    languageStore().setPageParams({});
    product_id.value = route.query.product_id;
    scene.value = route.query.scene;
    getData();
  });

  // 方法定义
  const getData = async () => {
    try {
      const res = await ProductApi.storeGetEditData(
        {
          product_id: product_id.value,
          scene: scene.value,
        },
        true
      );

      loading.value = false;

      try {
        if (res.data.model.alone_grade_equity != null && typeof res.data.model.alone_grade_equity != 'undefined') {
          let equitys = res.data.model.alone_grade_equity;
          res.data.gradeList.forEach((item) => {
            item.product_equity = equitys[item.grade_id];
          });
        }

        Object.assign(form, res.data);
        oldForm.value = JSON.parse(JSON.stringify(res.data));

        // 处理数据
        form.model.product_status = res.data.model.product_status.value;
        try {
          form.model.product_name = JSON.parse(form.model.product_name || '{}');
        } catch (e) {}


        form.category.map((item) => {
          if (form.model.category_id == item.category_id && item.child.length > 0) {
            form.model.category_id = '';
          }
        });

        try {
          form.model.product_unit = JSON.parse(form.model.product_unit || '{}');
        } catch (e) {}

        if (form.model.special_id == 0) {
          form.model.special_id = '';
        }

        if (form.model.erp_supplier_id == 0) {
          form.model.erp_supplier_id = '';
        }
        form.many_select_list = [];

        // 处理规格的材料组
        (form.model.sku || []).map((item, index) => {
          //删除多余字段
          delete item.create_time;
          delete item.update_time;

          //处理条唯一性
          item.barcodeUniqueness = true;

          form.many_select_list.push([]);
          if (item.spec_name) {
            try {
              form.model.sku[index].spec_name = JSON.parse(item.spec_name || '{}');
            } catch (e) {}
          } else {
            form.model.sku[index].spec_name = JSON.parse(languageData);
          }

          if (form.model.type == 10) {
            form.model.sku[index].spec_id = form.model.sku[index].spec_sku_id;
            item.material.map((items, indexs) => {
              form.many_select_list[index].push(items.materialProduct);
              form.many_select_list[index][indexs].product_id = form.model.sku[index].material[indexs].material_id;
              form.model.sku[index].material[indexs].product_id = form.model.sku[index].material[indexs].material_id;

              form.many_select_list[index][indexs].sku = [];
              form.many_select_list[index][indexs].sku[0] = {};
              form.many_select_list[index][indexs].sku[0].material_stock = form.many_select_list[index][indexs].product_material_stock;
              form.many_select_list[index][indexs].sku[0].product_id = form.model.sku[index].material[indexs].material_id;
            });
            if (form.many_select_list[0].length > 0) {
              form.single_select_list = JSON.parse(JSON.stringify(form.many_select_list[0]));
            }
            form.productPrinterList = res.data.productPrinterList;
            form.model.product_printer_uuids = [];
            (res.data.model.productPrinters || []).map((item) => {
              form.model.product_printer_uuids.push(item.product_printer_uuid);
            });
          }
        });

        //处理属性
        (form.model.product_attr || []).map((item) => {
          //处理旧数据向下兼容
          if (item.attribute_open_max_select === undefined) {
            item.attribute_open_max_select = 0;
          }
          if (item.attribute_required === undefined) {
            item.attribute_required = 0;
          }
          if (item.attribute_max_select === undefined) {
            item.attribute_required = 0;
          }
          if (item.parent_id === undefined) {
            form.attribute.map((items) => {
              if (item.attribute_name_text == items.parent_attribute_name_text) {
                item.parent_id = items.parent_id;
              }
            });
          }

          if (item.default_select === undefined) {
            item.default_select = [];
            item.attribute_ids = [];
            item.attribute_value_text.map((item2) => {
              form.attribute.map((items) => {
                if (item2 == items.attribute_name_text) {
                  item.attribute_ids.push(items.attribute_id);
                  item.default_select.push(0);
                }
              });
            });
          }
          //去掉这个 不用每次都提交
          delete item.attribute_value_text;
        });

        //  原本的product_feed会弃用，使用 feed
        form.model.product_feed = form.model.feed;
        // 处理商品的加料的材料
        if (form.model.type == 10 && form.model.product_feed.length > 0) {
          form.ing_select_list = [];
          form.model.product_feed.map((item, index) => {
            //删除多余字段
            delete item.create_time;
            delete item.update_time;

            item.default_select = Number(item.default_select);
            form.ing_select_list.push([]);
            form.model.product_feed[index].material.map((items, indexs) => {
              form.model.product_feed[index].material[indexs].product_id = items.material_id;
              form.ing_select_list[index].push(items.materialProduct);
              form.ing_select_list[index][indexs].sku = [];
              form.ing_select_list[index][indexs].sku[0] = {};
              form.ing_select_list[index][indexs].sku[0].material_stock = items.materialProduct.product_material_stock;
              form.ing_select_list[index][indexs].sku[0].product_id = items.material_id;
            });
          });
        }

        // 处理套餐
        if (form.model.type == 30) {
          form.model.package_group = JSON.parse(JSON.stringify(form.model.package.package_group));
          form.model.package_group.forEach((group) => {
            group.group_name = JSON.parse(group.group_name || '{}');
            group.group_type = group.group_type || 0;
            group.optional_count = group.optional_count || 1;
          });
          form.model.package_price = form.model.package.package_price;
          form.model.package_stock = form.model.package.package_stock;
          form.model.is_open_stock = form.model.package.is_open_stock;
        }
        
        try {
            form.model.selling_point_i18n = JSON.parse(form.model.selling_point_i18n);
        } catch (e) {
          console.log(e);
        }

      } catch (error) {
        console.log(error);
      }
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
    // 调用表单验证方法，valid为验证结果
    let valid = false;
    await formRef.value.validate((res) => {
      valid = res;
    });

    //单独验证产品名称
    const validUniqueName = await BasicRef.value.$refs.uniqueNameFormRef.validate();

    // 如果验证通过，则处理数据
    if (validUniqueName && valid) {
      //处理数据
      let params = formatModel(JSON.parse(JSON.stringify(model)), JSON.parse(JSON.stringify(form.model)));

      params.scene = JSON.parse(JSON.stringify(scene.value));
      params.image = JSON.parse(JSON.stringify(ImgKeepId(params.image)));
      params.product_id = JSON.parse(JSON.stringify(product_id.value));
      params.sku = JSON.parse(JSON.stringify(form.model.sku));
      params.alone_grade_equity = JSON.parse(JSON.stringify(convertJson(form.gradeList)));
      //处理数据
      const _name = BasicRef.value.$refs.uniqueNameFormRef.data;
      params.product_name = JSON.stringify(_name);

      params.product_unit = JSON.stringify(params.product_unit);

      params.spec_type = 20; //规格类型固定20
      // 处理sku中的规格
      let isError = false;
      params.sku.map((item, index) => {
        if (!item.spec_id && params.type == 10) {
          isError = true;
        }
        params.sku[index].spec_name = JSON.stringify(item.spec_name);
      });

      if (isError) {
        save_loading.value = false;
        ElMessage({
          message: $t('添加失败，请检查系统的语言设置。'),
          type: 'warning',
        });
        return;
      }

      params.stock_remark = tiao.stock_remark;

      // 处理sku中的材料
      (params.sku || []).map((skuItem, skuIndex) => {
        params.sku[skuIndex].material = [];
        (form.model.sku[skuIndex].material || []).map((materialItem) => {
          params.sku[skuIndex].material.push({
            product_id: materialItem.product_id,
            material_num: materialItem.material_num,
          });
        });
      });

      //税率处理
      params.productTaxes = [];
      form.model.productTaxes.map((item) => {
        params.productTaxes.push({
          tax_category_id: item.tax_category_id,
          product_tax_type: item.product_tax_type,
        });
      });

      // 如果分类id为对象且存在，则将其转换为数字
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
          stock_remark: tiao.stock_remark,
        };
        params = data;
      }

      // 如果套餐类型，则处理套餐数据
      if (params.type == 30) {
        params.package_group.forEach((group) => {
          group.group_name = JSON.stringify(group.group_name);
          group.group_type = group.group_type || 0;
          group.optional_count = group.optional_count || 1;
          // group.product_list 只需要保留product_id、num、sort，其他字段删除
          let productList = [];
          group.product_list.forEach((product) => {
            productList.push({
              product_id: product.product_id,
              num: product.num,
              sort: product.sort,
              item_id: product.item_id || 0,
              add_price: product.add_price, // 加价
              is_required: product.is_required, // 是否必选
              is_default: product.is_default, // 是否默认
            });
          });
          group.product_list = productList;
        });
        // 删除sku
        params.sku = [];
        // 套餐类型不显示外送
        params.is_show_delivery = 0;
      }

      // 处理商品卖点
      if (BuysetRef.value && BuysetRef.value.uniqueNameFormAreaTextRef && BuysetRef.value.uniqueNameFormAreaTextRef.data) {
        params.selling_point_i18n = JSON.stringify(BuysetRef.value.uniqueNameFormAreaTextRef.data);
      }

      //库存变动的时候
      // 如果旧表单类型为20且库存数量发生变化，则设置stockNumChange为true
      if (oldForm.value.model.type == '20' && oldForm.value.model.sku[0].material_stock != form.model.sku[0].material_stock) {
        stockNumChange.value = true;
      }
      // 如果旧表单类型为10，则比较库存数量是否发生变化
      if (oldForm.value.model.type == '10') {
        let oldArr = [];
        let newArr = [];
        oldForm.value.model.sku.map((item) => {
          oldArr.push(item.stock_num);
        });
        form.model.sku.map((item) => {
          newArr.push(item.stock_num);
        });
        oldArr = oldArr.join(',');
        newArr = newArr.join(',');
        // 如果库存数量发生变化，则设置stockNumChange为true
        if (oldArr != newArr) {
          stockNumChange.value = true;
        }
      }
    //   // 如果库存数量发生变化且e不等于1，则显示对话框
    //   if (stockNumChange.value && e != 1) {
    //     dialogVisible.value = true;
    //     save_loading.value = false;
    //     return;
    //   }
      // 如果库存数量发生变化且e等于1，则调用tiao的validate方法
      if (stockNumChange.value && e == 1) {
        tiaoRef.value.validate(() => {});
        // 如果tiao的stock_remark为空，则返回
        if (!tiao.stock_remark) return;
      }

      // 设置保存按钮的loading状态为true
      // 调用storeEditProduct方法，传入参数product_id和params
      try {
        save_loading.value = true;
        await ProductApi.storeEditProduct(
          {
            product_id: product_id.value,
            params: JSON.stringify(params),
          },
          true
        );

        save_loading.value = false;
        ElMessage({
          message: $t('保存成功'),
          type: 'success',
        });
        cancelFunc();
      } catch (res) {
        save_loading.value = false;
        // dialogVisible.value = false;
        if ((res.data || []).length > 0) {
          await res.data.map((item, index) => {
            form.model.sku[index].barcodeUniqueness = item;
          });
        }
        await formRef.value.validate(() => {
          moveToError();
        });
      }
    } else {
      save_loading.value = false;
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

  /*图片数值只保留id*/
  const ImgKeepId = (list) => {
    let arr = [];
    for (let i = 0, length = list.length; i < length; i++) {
      let obj = {
        image_id: list[i].image_id,
        file_id: list[i].file_id,
      };
      arr.push(obj);
    }
    return arr;
  };

  /*保存为草稿*/
  const Draft = () => {
    form.model.product_status = 30;
    onSubmit();
  };

  /*取消*/
  const cancelFunc = () => {
    languageStore().setPageParams(pageParams.value);
    router.push('/' + app_id_ref.value + '/product/store/index');
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
  }

  .mb50 {
    margin-bottom: 50px;
  }
</style>

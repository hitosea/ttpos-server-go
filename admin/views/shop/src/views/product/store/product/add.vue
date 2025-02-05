<template>
  <div class="product-add">
    <!--form表单-->
    <el-form size="small" ref="form" class="product-form" :model="form" label-position="top" label-width="180px">
      <!--基础信息-->
      <div class="product-form-flex" ref="formContainer">
        <Basic ref="BasicRef" @validateField="validateField"></Basic>

        <!--规格设置-->
        <Spec></Spec>

        <!-- 属性设置-->
        <Attr ref="AttrRef" v-if="form.model.type == 10" @validateField="validateField"></Attr>

        <!-- 加料设置-->
        <Ingredients ref="IngredientsRef" v-if="form.model.type == 10" @validateField="validateField"></Ingredients>

        <!--商品详情-->
        <!-- <Content></Content> -->

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

<script>
  import ProductApi from '@/api/product.js';
  import Basic from './part/Basic.vue';
  import Attr from './part/Attr.vue';
  import Ingredients from './part/Ingredients.vue';
  import Spec from './part/Spec.vue';
  // import Content from './part/Content.vue';
  import Buyset from './part/Buyset.vue';
  import { languageStore } from '@/store/model/language.js';
  import { EEUIRELOAD } from '@/utils/platform.js';

  const languageData = JSON.stringify(languageStore().getLanguageData().languageData.value);
  // eslint-disable-next-line no-unused-vars
  const languageList = languageStore().getLanguageList().languageList.value;
  import { useUserStore } from '@/store/index';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;

  export default {
    name: 'ProductStoreProductAdd',
    components: {
      /*基础信息*/
      Basic,
      /*规格信息*/
      Spec,
      /* 属性信息*/
      Attr,
      /*加料设置*/
      Ingredients,
      /*商品详情*/
      // Content,
      /*高级设置*/
      Buyset,
    },
    data() {
      return {
        app_id: app_id,
        /*是否正在加载*/
        loading: false,
        active: false,

        /*form表单数据*/
        form: {
          model: {
            /*商品类型*/
            type: 10,
            /*商品名称*/
            product_name: JSON.parse(languageData),
            /*商品分类*/
            category_id: null,
            /*供应商id*/
            erp_supplier_id: null,
            /*商品图片*/
            image: [],
            /*商品图片名称*/
            img_name: '',
            /*商品卖点*/
            selling_point: '',
            /*规格类别,默认10单规格，20多规格*/
            spec_type: 20,
            /*库存计算方式,默认20付款减库存，10下单减库存*/
            deduct_stock_type: 20,
            /*检查用户等级*/
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
                barcodeUniqueness: true, //条形码是否唯一
              },
            ],
            product_attr: [],
            product_feed: [],
            feed_required: 0,
            feed_open_max_select: 0,
            feed_max_select: 0,
            /* 最小购买量 */
            min_buy: 1,
            /* 商品单位 */
            product_unit: JSON.parse(languageData),
            unit_id: '',
            /* 属性*/
            /*商品详情内容*/
            content: '',
            /*商品状态*/
            product_status: 10,
            /*商品材料*/
            material: [],
            /*收银机是否显示*/
            is_show_cashier: 1,
            /*平板是否显示*/
            is_show_tablet: 1,
            /*厨显是否显示*/
            is_show_kitchen: 1,
            /*点餐助手是否显示*/
            is_show_assistant: 1,
            /*H5是否显示*/
            is_show_h5: 1,
            /*初始销量*/
            sales_initial: 0,
            /*商品排序，默认100*/
            product_sort: 0,
            /*限购数量*/
            limit_num: 0,
            special_id: '',
            /*是否开启积分赠送,默认1为开启，0为关闭*/
            is_points_gift: 1,
            is_agent: 0,
            /*是否开启单独分销,默认0为关闭，1为开启*/
            is_ind_agent: 0,
            /*分销佣金类型,默认10为百分比，20为固定金额*/
            agent_money_type: 10,
            /*一级佣金*/
            first_money: 0,
            /*二级佣金*/
            second_money: 0,
            /*三级佣金*/
            third_money: 0,
            /*会员折扣设置,默认1为单独设置折扣,0为默认折扣*/
            is_enable_grade: 1,
            /*等级金额类型,默认10为百分比，20为固定金额*/
            alone_grade_type: 10,
            /*打印标签*/
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
          },
          /*商品分类*/
          category: [],
          feed: [],
          attribute: [],
          unit: [],
          spec: [],
          labelList: [],
          special: [],
          /*运费模板*/
          delivery: [],
          /*会员等级*/
          gradeList: [],
          /*规格数据*/
          specData: null,
          /*分销基础设置*/
          basicSetting: {},
          /*分销佣金设置*/
          agentSetting: {},
          /*单规格的材料组*/
          single_select_list: [],
          /*多规格的材料组*/
          many_select_list: [[]],
          /*加料的材料组*/
          ing_select_list: [[]],
        },
      };
    },
    provide: function () {
      return {
        form: this.form,
      };
    },
    created() {
      this.pageParams = JSON.parse(JSON.stringify(languageStore().getPageParams().pageParams.value));
      languageStore().setPageParams({});
      /*获取基础数据*/
      this.getBaseData();
    },

    methods: {
      /*获取基础数据*/
      getBaseData: function () {
        let self = this;
        ProductApi.storeGetBaseData({}, true)
          .then((res) => {
            self.loading = false;
            Object.assign(self.form, res.data);
          })
          .catch(() => {
            self.loading = false;
          });
      },

      /*转JSON字符串*/
      convertJson(list) {
        let obj = {};
        list.forEach((item) => {
          obj[item.grade_id] = item.product_equity;
        });
        return JSON.stringify(obj);
      },

      validateField(e) {
        this.$refs.form.validateField(e);
      },

      /*提交*/
      async onSubmit(e) {
        // 获取当前实例
        let self = this;
        // 设置加载状态
        self.loading = true;
        //单独验证产品名称
        let valid = false;

        await self.$refs.form.validate((res) => {
          valid = res;
        });

        // 验证表单
        const validUniqueName = await self.$refs.BasicRef.$refs.uniqueNameFormRef.validate();

        // 如果表单验证通过，则处理参数
        if (validUniqueName && valid) {
          let params = {};
          params = JSON.parse(JSON.stringify(self.form.model));
          // 将产品名称和产品单位转换为字符串
          const _name = self.$refs.BasicRef.$refs.uniqueNameFormRef.data;
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

          // 将等级列表转换为json
          params.alone_grade_equity = self.convertJson(self.form.gradeList);

          // 调用接口添加产品
          ProductApi.storeAddProduct(
            {
              params: JSON.stringify(params),
            },
            true
          )
            .then(() => {
              // 设置加载状态
              self.loading = false;
              self.$ElMessage({
                message: self.$t('添加成功'),
                type: 'success',
              });
              self.cancelFunc(e);
            })
            .catch(async (res) => {
              self.loading = false;
              if ((res.data || []).length > 0) {
                await res.data.map((item, index) => {
                  self.form.model.sku[index].barcodeUniqueness = item;
                });
              }
              await self.$refs.form.validate(() => {
                self.moveToError();
              });
            });
        } else {
          self.loading = false;
          // 如果表单验证不通过，则滚动到错误位置
          self.moveToError();
        }
      },

      moveToError() {
        // 设置一个定时器，在200毫秒后，获取所有带有el-form-item__error类的元素，并将第一个元素滚动到视图中
        setTimeout(() => {
          const errorItems = document.querySelectorAll('.el-form-item__error');
          if (errorItems.length > 0) {
            const firstErrorItem = errorItems[0];
            firstErrorItem.scrollIntoView({ behavior: 'smooth', block: 'center' });
          }
        }, 200);
      },

      /*保存为草稿*/
      Draft() {
        let self = this;
        self.form.model.product_status = 30;
        self.onSubmit();
      },

      /*取消*/
      cancelFunc(e) {
        if (e == 1) {
          languageStore().setPageParams(this.pageParams);
          this.$router.push('/' + this.app_id + '/product/store/index');
        } else {
          languageStore().setPageParams(this.pageParams);
          EEUIRELOAD();
        }
      },
    },
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

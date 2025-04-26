<template>
  <div class="product-add" v-loading="loading">
    <!--form表单-->
    <el-form size="small" ref="form" :model="form" class="product-form" label-position="top" label-width="180px" v-if="!loading">
      <!--基础信息-->
      <div class="product-form-flex">
        <Basic ref="BasicRef" @validateField="validateField" disableChangeType></Basic>
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
        <el-button size="small" @click="() => cancelFunc()">{{ $t('取消') }}</el-button>
        <el-tooltip effect="dark" placement="top" :content="$t('如长时间无响应，请刷新后重试。')">
          <el-button size="small" type="primary" @click="() => onSubmit()" :loading="save_loading">{{ $t('确定') }}</el-button>
        </el-tooltip>
      </div>
    </el-form>
    <!-- 调整库存弹窗 -->
    <el-dialog v-if="dialogVisible" v-model="dialogVisible" :title="$t('调整库存')" width="700" align-center append-to-body>
      <el-form size="small" :inline="true" ref="tiao" :model="tiao" label-position="top">
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
          <el-button type="primary" @click="() => onSubmit(1)"> {{ $t('确定') }}</el-button>
        </div>
      </template>
    </el-dialog>
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
  import { useUserStore } from '@/store/index';
  import { formatModel } from '@/utils/base.js';
  import IndexApi from '@/api/index.js';

  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;
  const languageData = JSON.stringify(languageStore().getLanguageData().languageData.value);
  const languageList = languageStore().getLanguageList().languageList.value;
  let language = languageStore().language;
  let languageKey = '0';
  languageList.map((item) => {
    if (item.name == language) {
      languageKey = item.key;
    }
  });

  export default {
    name: 'ProductStoreProductEdit',
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
        /* 审核状态*/
        old_audit: 20,
        /*商品ID*/
        product_id: 0,
        /*判断是编辑*/
        scene: 'edit',
        /*是否正在加载*/
        loading: true,
        /*是否正在提交保存*/
        save_loading: false,
        /*form表单数据*/
        form: {
          model: {},
          /*商品分类*/
          category: [],
          feed: [],
          attribute: [],
          unit: [],
          spec: [],
          labelList: [],
          special: [],
          /*会员等级*/
          gradeList: [],
          /*规格数据*/
          specData: null,
          /*是否锁住*/
          isSpecLocked: false,
          /*分销基础设置*/
          basicSetting: {},
          /*分销佣金设置*/
          agentSetting: {},
          /*打印标签*/
          label_id: '',
          /*单规格的材料组*/
          single_select_list: [],
          /*多规格的材料组*/
          many_select_list: [[]],
          /*加料的材料组*/
          ing_select_list: [[]],
        },
        /*模型数据*/
        model: {
          /*商品类型*/
          type: 10,
          /*商品名称*/
          product_name: '',
          /*商品分类*/
          category_id: null,
          /*商品图片*/
          image: [],
          /*商品图片名称*/
          img_name: '',
          /*商品卖点*/
          selling_point: '',
          /*供应商*/
          erp_supplier_id: '',
          /*规格类别,默认10单规格，20多规格*/
          spec_type: 20,
          /*库存计算方式,默认20付款减库存，10下单减库存*/
          deduct_stock_type: 20,
          /*检查用户等级*/
          is_alone_grade: 0,
          /*单规格*/
          sku: [
            {
              product_price: '',
              stock_num: '',
              product_weight: '',
              cost_price: '',
              material: [],
            },
          ],
          /* 调整库存的备注 */
          stock_remark: '',
          product_attr: [],
          product_feed: [],
          feed_required: 0,
          feed_open_max_select: 0,
          feed_max_select: 0,
          /* 最小购买量 */
          min_buy: 1,
          /* 商品单位 */
          product_unit: '',
          unit_id: '',
          /*商品详情内容*/
          content: '',
          /*商品状态*/
          product_status: 10,
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
        },
        oldForm: {},
        /*库存变没变*/
        tiao: {
          stock_remark: '',
        },

        stockNumChange: false,
        dialogVisible: false,
        languageKey: languageKey,
        pageParams: {},
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
      /*获取列表*/
      this.product_id = this.$route.query.product_id;
      this.scene = this.$route.query.scene;
      this.getData();
    },

    methods: {
      /**
       * 获取基础数据
       */
      getData: function () {
        let self = this;
        ProductApi.storeGetEditData(
          {
            product_id: self.product_id,
            scene: self.scene,
          },
          true
        )
          .then((res) => {
            self.loading = false;
            try {
              if (res.data.model.alone_grade_equity != null && typeof res.data.model.alone_grade_equity != 'undefined') {
                let equitys = res.data.model.alone_grade_equity;
                res.data.gradeList.forEach((item) => {
                  item.product_equity = equitys[item.grade_id];
                });
              }

              Object.assign(self.form, res.data);
              self.oldForm = JSON.parse(JSON.stringify(res.data));
              // 处理数据
              self.form.model.product_status = res.data.model.product_status.value;
              try {
                self.form.model.product_name = JSON.parse(self.form.model.product_name || '{}');
              } catch (e) {}

              self.form.category.map((item) => {
                if (self.form.model.category_id == item.category_id && item.child.length > 0) {
                  self.form.model.category_id = '';
                }
              });

              try {
                self.form.model.product_unit = JSON.parse(self.form.model.product_unit || '{}');
              } catch (e) {}

              if (self.form.model.special_id == 0) {
                self.form.model.special_id = '';
              }

              if (self.form.model.erp_supplier_id == 0) {
                self.form.model.erp_supplier_id = '';
              }
              this.form.many_select_list = [];

              // 处理规格的材料组
              self.form.model.sku.map((item, index) => {
                //删除多余字段
                delete item.create_time;
                delete item.update_time;

                //处理条唯一性
                item.barcodeUniqueness = true;

                this.form.many_select_list.push([]);
                if (item.spec_name) {
                  try {
                    self.form.model.sku[index].spec_name = JSON.parse(item.spec_name || '{}');
                  } catch (e) {}
                } else {
                  self.form.model.sku[index].spec_name = JSON.parse(languageData);
                }

                if (self.form.model.type == 10) {
                  self.form.model.sku[index].spec_id = self.form.model.sku[index].spec_sku_id;
                  item.material.map((items, indexs) => {
                    this.form.many_select_list[index].push(items.materialProduct);
                    this.form.many_select_list[index][indexs].product_id = this.form.model.sku[index].material[indexs].material_id;
                    this.form.model.sku[index].material[indexs].product_id = this.form.model.sku[index].material[indexs].material_id;

                    this.form.many_select_list[index][indexs].sku = [];
                    this.form.many_select_list[index][indexs].sku[0] = {};
                    this.form.many_select_list[index][indexs].sku[0].material_stock = this.form.many_select_list[index][indexs].product_material_stock;
                    this.form.many_select_list[index][indexs].sku[0].product_id = this.form.model.sku[index].material[indexs].material_id;
                  });
                  if (this.form.many_select_list[0].length > 0) {
                    this.form.single_select_list = JSON.parse(JSON.stringify(this.form.many_select_list[0]));
                  }
                }
              });

              //处理属性
              self.form.model.product_attr.map((item) => {
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
                  self.form.attribute.map((items) => {
                    if (item.attribute_name_text == items.parent_attribute_name_text) {
                      item.parent_id = items.parent_id;
                    }
                  });
                }

                if (item.default_select === undefined) {
                  item.default_select = [];
                  item.attribute_ids = [];
                  item.attribute_value_text.map((item2) => {
                    self.form.attribute.map((items) => {
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
              self.form.model.product_feed = self.form.model.feed;
              // 处理商品的加料的材料
              if (self.form.model.type == 10 && self.form.model.product_feed.length > 0) {
                self.form.ing_select_list = [];
                self.form.model.product_feed.map((item, index) => {
                  //删除多余字段
                  delete item.create_time;
                  delete item.update_time;

                  item.default_select = Number(item.default_select);
                  self.form.ing_select_list.push([]);
                  self.form.model.product_feed[index].material.map((items, indexs) => {
                    self.form.model.product_feed[index].material[indexs].product_id = items.material_id;
                    self.form.ing_select_list[index].push(items.materialProduct);
                    self.form.ing_select_list[index][indexs].sku = [];
                    self.form.ing_select_list[index][indexs].sku[0] = {};
                    self.form.ing_select_list[index][indexs].sku[0].material_stock = items.materialProduct.product_material_stock;
                    self.form.ing_select_list[index][indexs].sku[0].product_id = items.material_id;
                  });
                });
              }
            } catch (error) {
              console.log(error);
            }
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
        // 定义一个变量self，指向当前组件实例
        let self = this;
        self.save_loading = true;

        // 调用表单验证方法，valid为验证结果
        let valid = false;
        await self.$refs.form.validate((res) => {
          valid = res;
        });

        //单独验证产品名称
        const validUniqueName = await self.$refs.BasicRef.$refs.uniqueNameFormRef.validate();

        // 如果验证通过，则处理数据
        if (validUniqueName && valid) {
          //处理数据
          let params = formatModel(JSON.parse(JSON.stringify(self.model)), JSON.parse(JSON.stringify(self.form.model)));

          params.scene = JSON.parse(JSON.stringify(self.scene));
          params.image = JSON.parse(JSON.stringify(self.ImgKeepId(params.image)));
          params.product_id = JSON.parse(JSON.stringify(self.product_id));
          params.sku = JSON.parse(JSON.stringify(self.form.model.sku));
          params.alone_grade_equity = JSON.parse(JSON.stringify(self.convertJson(self.form.gradeList)));
          //处理数据
          const _name = self.$refs.BasicRef.$refs.uniqueNameFormRef.data;
          params.product_name = JSON.stringify(_name);

          params.product_unit = JSON.stringify(params.product_unit);

          params.spec_type = 20; //规格类型固定20
          // 处理sku中的规格
          let isError = false;
          params.sku.map((item, index) => {
            if (!item.spec_id) {
              isError = true;
            }
            params.sku[index].spec_name = JSON.stringify(item.spec_name);
          });

          if (isError) {
            self.save_loading = false;
            self.$ElMessage({
              message: self.$t('添加失败，请检查系统的语言设置。'),
              type: 'warning',
            });
            return;
          }

          params.stock_remark = self.tiao.stock_remark;

          // 处理sku中的材料
          params.sku.map((skuItem, skuIndex) => {
            params.sku[skuIndex].material = [];
            self.form.model.sku[skuIndex].material.map((materialItem) => {
              params.sku[skuIndex].material.push({
                product_id: materialItem.product_id,
                material_num: materialItem.material_num,
              });
            });
          });

          //税率处理
          params.productTaxes = [];
          self.form.model.productTaxes.map((item) => {
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
              stock_remark: self.tiao.stock_remark,
            };
            params = data;
          }
          //库存变动的时候
          // 如果旧表单类型为20且库存数量发生变化，则设置stockNumChange为true
          if (this.oldForm.model.type == '20' && this.oldForm.model.sku[0].material_stock != this.form.model.sku[0].material_stock) {
            this.stockNumChange = true;
          }
          // 如果旧表单类型为10，则比较库存数量是否发生变化
          if (this.oldForm.model.type == '10') {
            let oldArr = [];
            let newArr = [];
            this.oldForm.model.sku.map((item) => {
              oldArr.push(item.stock_num);
            });
            this.form.model.sku.map((item) => {
              newArr.push(item.stock_num);
            });
            oldArr = oldArr.join(',');
            newArr = newArr.join(',');
            // 如果库存数量发生变化，则设置stockNumChange为true
            if (oldArr != newArr) {
              this.stockNumChange = true;
            }
          }
          // 如果库存数量发生变化且e不等于1，则显示对话框
          if (this.stockNumChange && e != 1) {
            this.dialogVisible = true;
            self.save_loading = false;
            return;
          }
          // 如果库存数量发生变化且e等于1，则调用tiao的validate方法
          if (this.stockNumChange && e == 1) {
            self.$refs.tiao.validate(() => {});
            // 如果tiao的stock_remark为空，则返回
            if (!self.tiao.stock_remark) return;
          }

          // 设置保存按钮的loading状态为true
          // 调用storeEditProduct方法，传入参数product_id和params
          ProductApi.storeEditProduct(
            {
              product_id: self.product_id,
              params: JSON.stringify(params),
            },
            true
          )
            // 如果请求成功，则设置保存按钮的loading状态为false，并显示保存成功的提示信息，调用cancelFunc方法
            .then(() => {
              self.save_loading = false;
              self.$ElMessage({
                message: self.$t('保存成功'),
                type: 'success',
              });
              self.cancelFunc();
            })
            // 如果请求失败，则设置保存按钮的loading状态为false
            .catch(async (res) => {
              self.save_loading = false;
              this.dialogVisible = false;
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
          self.save_loading = false;
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

      /*图片数值只保留id*/
      ImgKeepId(list) {
        let arr = [];
        for (let i = 0, length = list.length; i < length; i++) {
          let obj = {
            image_id: list[i].image_id,
            file_id: list[i].file_id,
          };
          arr.push(obj);
        }
        return arr;
      },

      /*保存为草稿*/
      Draft() {
        let self = this;
        self.form.model.product_status = 30;
        self.onSubmit();
      },

      /*取消*/
      cancelFunc() {
        languageStore().setPageParams(this.pageParams);
        this.$router.push('/' + this.app_id + '/product/store/index');
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
  }

  .mb50 {
    margin-bottom: 50px;
  }
</style>

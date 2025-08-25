<template>
  <!--

    	时间：2019-10-26
    	描述：商品管理-商品编辑-规格/库存-多规格-表格
    -->
  <div class="mt16">
    <el-form-item :label="$t('规格明细：')" v-if="form.model.sku.length > 0">
      <!--规格类别-->
      <Type></Type>
      <!--多规格表格-->
      <div class="ww100">
        <el-table size="" :data="form.model.sku" border style="width: 100%; margin-top: 20px">
          <el-table-column :label="$t('规格名称')" width="360">
            <template #default="scope">
              <div label="" class="spec-name" style="margin-bottom: 0">
                <template v-for="(item, index) in languageList">
                  <el-form-item
                    for="no_click"
                    :key="index"
                    v-if="item.name == languageKey"
                    :prop="`scope.row.spec_name[${item.key}]`"
                    class="flex-wrap"
                    :rules="[
                      {
                        validator: () => {
                          return scope.row.spec_name[item.key] ? true : false;
                        },
                        message: $t('请选择规格名称'),
                      },
                    ]"
                  >
                    <el-select
                      v-model="scope.row.spec_name[item.key]"
                      @change="(e) => selectChange(e, scope.$index)"
                      filterable
                      clearable
                      class="max-w460"
                      size="default"
                      :placeholder="$t('请选择') + `(${item.value})`"
                    >
                      <template v-for="items in restaurantsObj[item.key]">
                        <el-option :value="items.index" :label="items.value"></el-option>
                      </template>
                    </el-select>
                    <el-button size="small" type="primary" class="el-icon-circle-plus mr0" @click="addSku">{{ $t('添加规格') }}+</el-button>
                  </el-form-item>
                </template>
              </div>
            </template>
          </el-table-column>

          <el-table-column v-if="baseSale == '1'" :label="$t('采购单价')" minWidth="160">
            <template #default="scope">
              <el-form-item for="no_click" label="" style="margin-bottom: 0">
                <numInput :min="0" :max="100000000" :precision="2" :placeholder="$t('请输入采购单价')" v-model="scope.row.purchase_price"></numInput>
              </el-form-item>
            </template>
          </el-table-column>
          <el-table-column :label="'*' + $t('库存')" minWidth="160">
            <template #default="scope">
              <el-form-item
                for="no_click"
                label=""
                style="margin-bottom: 0"
                :prop="`scope.row.stock_num`"
                :rules="[
                  {
                    validator: (rule, value, callback) => {
                      if (!scope.row.stock_num) {
                        callback(new Error($t('请输入库存')));
                        return;
                      }
                      if (form.model.num_type == 0) {
                        if (!Number.isInteger(Number(scope.row.stock_num))) {
                          callback(new Error($t('请输入整数')));
                          return;
                        }
                      }
                      callback();
                    },
                    trigger: 'change',
                  },
                ]"
              >
                <numInput
                  :disabled="scope.row.material.length > 0"
                  :min="0"
                  :max="99999999"
                  :precision="2"
                  :placeholder="$t('请输入库存')"
                  v-model="scope.row.stock_num"
                ></numInput>
              </el-form-item>
            </template>
          </el-table-column>
          <el-table-column v-if="baseSale == '1'" :label="$t('商品条码')" minWidth="160">
            <template #default="scope">
              <el-form-item
                for="no_click"
                label=""
                :prop="`model.sku.${scope.$index}.barcode`"
                style="margin-bottom: 0"
                :rules="[
                  {
                    validator: (rule, value, callback) => {
                      // 判断长度是否为12或13位，并且只能输入数字
                      if (value && value.length !== 12 && value.length !== 13) {
                        callback(new Error($t('长度12-13位数字')));
                        return;
                      }

                      if (!scope.row.barcodeUniqueness) {
                        callback(new Error($t('商品条码重复')));
                        return;
                      }
                      callback();
                    },
                    trigger: 'change',
                  },
                ]"
              >
                <el-input
                  v-model="scope.row.barcode"
                  :maxlength="13"
                  @input="
                    () => {
                      scope.row.barcodeUniqueness = true;
                    }
                  "
                  :placeholder="$t('请输入商品条码')"
                ></el-input>
              </el-form-item>
            </template>
          </el-table-column>
          <el-table-column :label="'*' + $t('商品价格')" minWidth="160">
            <template #default="scope">
              <el-form-item
                for="no_click"
                label=""
                style="margin-bottom: 0"
                :prop="`scope.row.product_price`"
                :rules="[
                  {
                    validator: () => {
                      return scope.row.product_price ? true : false;
                    },
                    message: $t('请输入商品价格'),
                    trigger: 'change',
                  },
                ]"
              >
                <numInput :min="0" :max="100000000" :precision="2" :placeholder="$t('请输入商品价格')" v-model="scope.row.product_price"></numInput>
              </el-form-item>
            </template>
          </el-table-column>
          <el-table-column v-if="baseSale == '1'" :label="$t('材料')" minWidth="330">
            <template #default="scope">
              <el-form-item for="no_click" label="" style="margin-bottom: 0">
                <el-button type="primary" :style="form.many_select_list[scope.$index].length > 0 ? 'margin-top: 16px;' : ''" @click="addMaterials(scope.$index)">{{
                  $t('添加材料')
                }}</el-button>
              </el-form-item>
              <div class="materials-one" label="" v-for="(item, index) in form.many_select_list[scope.$index]">
                <el-form-item for="no_click" label="" class="max-w230">
                  <el-input v-model="item.product_name_text" disabled></el-input>
                </el-form-item>
                <el-form-item
                  for="no_click"
                  label=""
                  class="max-w230"
                  :prop="`form.model.sku.[${scope.$index}].material.[${index}].material_num`"
                  :rules="[
                    {
                      validator: () => {
                        return form.model.sku[scope.$index].material[index].material_num && form.model.sku[scope.$index].material[index].material_num != 0 ? true : false;
                      },
                      message: $t('请输入数量'),
                    },
                  ]"
                >
                  <el-input-number :min="0" :max="999" :controls="false" v-model="form.model.sku[scope.$index].material[index].material_num" :placeholder="$t('请输入数量')">
                  </el-input-number>
                </el-form-item>
                <span class="mt--16">{{ item.product_unit_text }}</span>
                <el-icon class="delete-icon" @click="handleDelete(scope.$index, index)">
                  <Delete />
                </el-icon>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="" fixed="right" width="100">
            <template #default="scope">
              <el-form-item label="" style="margin-bottom: 0">
                <el-button type="primary" :disabled="form.model.sku.length <= 1" link @click="deleteAttr(scope.$index)">{{ $t('删除') }}</el-button>
              </el-form-item>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-form-item>
    <productList
      v-if="open_product"
      :open_product="open_product"
      :index="index"
      :multiple_selection="multiple_selection"
      :material_type="20"
      @closeDialogFunc="closeDialogFunc($event)"
    >
    </productList>
    <!-- 新增规格 -->
    <Add v-if="open_add" :open_add="open_add" :addform="model" @closeDialog="closeDialog($event)"></Add>
  </div>
</template>

<script>
  import productList from '@/components/productList/productList.vue';
  import { languageStore } from '@/store/model/language.js';
  import mAutocomplete from '@/components/m-autocomplete/index.vue';
  import { useUserStore } from '@/store';
  import Type from './Type.vue';
  import Add from '../../../../../expand/spec/add.vue';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const baseSale = supplier.value?.sale_stock || 0;
  const languageList = languageStore().getLanguageList().languageList.value;
  const languageKey = languageStore().getLanguageKey().language.value;

  export default {
    components: {
      productList,
      mAutocomplete,
      /*规格类别属性*/
      Type,
      Add,
    },
    data() {
      let languageObj = {};
      languageList.forEach((item) => {
        languageObj[item.key] = [];
      });
      return {
        languageList: languageList,
        languageKey: languageKey,
        restaurants: [],
        restaurantsObj: languageObj,
        /*批量设置sku属性*/
        batchData: {
          product_price: '',
          line_price: '',
          stock_num: '',
          product_weight: '',
        },
        /*图片是否打开*/
        isupload: false,
        //上传图片选择的下标
        spec_index: -1,
        //材料
        open_product: false,
        multiple_selection: [],
        index: 0,
        baseSale: baseSale,
        open_add: false,
      };
    },
    inject: ['form'],
    watch: {
      form: {
        handler(val) {
          let languageObj = {};
          languageList.forEach((item) => {
            languageObj[item.key] = [];
          });
          this.restaurantsObj = languageObj;
          val.spec.map((item, index) => {
            let spec_name = JSON.parse(item.spec_name);
            languageList.forEach((items) => {
              if (spec_name[items.key] != null) {
                this.restaurantsObj[items.key].push({
                  value: spec_name[items.key] == '' ? '-' : spec_name[items.key],
                  index: index,
                  spec_id: item.spec_id,
                });
              }
            });
          });

          // 用材料的库存来计算这个规格的库存
          (val.model.sku || []).map((item, index) => {
            //无材料的时候库存取整；2025年07月07日14:22:32；2.3.0 - 商家后台/收银机 -小计商品下单数量：50.5 - 库存/销量未记录为：50.5
            // if ((item.material || []).length == 0) {
            //   this.form.model.sku[index].stock_num = Number(String(item.stock_num).replace(/(\.\d{0,0})\d*/, '$1'));
            // }

            //处理条形码
            this.$nextTick(() => {
              if (this.form.model.sku[index].barcode) {
                //只能输入纯数字
                this.form.model.sku[index].barcode = item.barcode.match(/[0-9]*/g).join('');
              }
            });

            let arr = [];
            (item.material || []).map((items, indexs) => {
              //有材料的时候4位小数
              if (items.material_num != null) {
                this.form.model.sku[index].material[indexs].material_num = String(items.material_num).replace(/(\.\d{1,4})\d*/, '$1');
              }

              let num = 0;
              num = Number(this.form.many_select_list[index][indexs].sku[0]?.material_stock) / Number(items.material_num);
              num = Math.floor(num);
              arr.push(num);
            });
            if ((item.material || []).length > 0) {
              this.form.model.sku[index].stock_num =
                arr.sort((a, b) => a - b)[0] == Infinity ? null : arr.sort((a, b) => a - b)[0] > 99999999 ? 99999999 : arr.sort((a, b) => a - b)[0];
            }
          });
        },
        deep: true,
        immediate: true,
      },
    },
    mounted() {},
    methods: {
      deleteAttr(i) {
        if (this.form.model.sku.length > 1) {
          this.form.model.sku.splice(i, 1);
          this.form.many_select_list.splice(i, 1);
          if (i == 0) {
            this.form.single_select_list = [];
            this.form.single_select_list = this.form.many_select_list[0];
          }
        }
      },

      addSku() {
        this.open_add = true;
      },

      selectChange(e, index) {
        languageList.forEach((item) => {
          this.form.model.sku[index].spec_name[item.key] = this.restaurantsObj[item.key][e]?.value || '';
          this.form.model.sku[index].spec_id = this.restaurantsObj[item.key][e]?.spec_id || '';
        });
        //多规格的时候填入材料
        this.form.model.sku[index].material = [];
        this.form.many_select_list[index] = [];
        // v.1.0.8 去掉材料
        // if (!e) return;
        // if (this.form.spec[e].material.length > 0) {
        //   this.form.spec[e].material.map((item) => {
        //     this.form.model.sku[index].material.push({
        //       material_num: Number(item.material_num),
        //       product_id: item.materialProduct.product_id,
        //     });
        //     this.form.many_select_list[index].push(item);
        //   });
        // }
        // //多规格的时候填入材料
        // this.form.many_select_list[index].map((items, indexs) => {
        //   this.form.many_select_list[index][indexs].sku = [];
        //   this.form.many_select_list[index][indexs].sku[0] = {
        //     product_id: '',
        //     material_stock: '',
        //   };
        //   this.form.many_select_list[index][indexs].sku[0].product_id = items.material_id;
        //   this.form.many_select_list[index][indexs].sku[0].material_stock = items.materialProduct.product_material_stock;
        //   this.form.many_select_list[index][indexs].product_unit_text = items.materialProduct.product_unit_text;
        //   this.form.many_select_list[index][indexs].product_name_text = items.materialProduct.product_name_text;
        // });
        // // 切换规格的时候
        // if (this.form.many_select_list[0].length > 0) {
        //   this.form.single_select_list = JSON.parse(JSON.stringify(this.form.many_select_list[0]));
        // }
      },

      addMaterials(index) {
        this.multiple_selection = this.form.many_select_list[index];
        this.index = index;
        this.open_product = true;
      },

      handleDelete($index, index) {
        this.form.many_select_list[$index].splice(index, 1);
        this.form.model.sku[$index].material.splice(index, 1);
        if ($index == 0) {
          this.form.single_select_list.splice(index, 1);
        }
      },

      closeDialogFunc(e) {
        this.open_product = e.openDialog;
        if (e.type == 'select') {
          let map = new Map();
          if (e.index == 0) {
            [this.form.single_select_list, e.data].flat().forEach((obj) => map.set(obj.product_id, obj));
            this.form.single_select_list = Array.from(map.values());
          }

          [this.form.many_select_list[e.index], e.data].flat().forEach((obj) => map.set(obj.product_id, obj));
          this.form.many_select_list[e.index] = Array.from(map.values());

          let arr = [];
          if (this.form.model.sku[e.index].material.length > 0) {
            this.form.model.sku[e.index].material.map((item) => {
              arr.push(item.product_id);
            });
          }

          this.form.many_select_list[e.index].map((item) => {
            if (!arr.includes(item.product_id)) {
              this.form.model.sku[e.index].material.push({
                product_id: item.product_id,
                material_num: null,
              });
            }
          });
        }
      },

      closeDialog(e) {
        /*关闭弹窗*/
        this.open_add = e.openDialog;
        if (e.type == 'success' && e.data) {
          this.form.spec.unshift(e.data);
        }
      },

      /*翻译*/
      translate(lang, index) {
        this.languageList.map((item) => {
          lang.map((items) => {
            let key = item.name;
            if (key == 'zhtw') {
              key = 'zh-TW';
            }
            if (items[key]) {
              this.form.model.sku[index].spec_name[item.key] = items[key];
            }
          });
        });
      },
    },
  };
</script>

<style scoped>
  .spec-name {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px;
  }

  .spec-name .el-input {
    max-width: calc(50% - 6px);
  }

  .materials-one {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 8px;

    .delete-icon {
      cursor: pointer;
      font-size: 24px;
      margin-top: -16px;
    }
  }

  .max-w230 {
    max-width: 226px;
    width: 100%;
  }
  .mt--16 {
    margin-top: -16px;
  }

  .flex-wrap :deep(.el-form-item__content) {
    flex-wrap: nowrap;
  }
  .mr0 {
    margin-right: 0 !important;
    margin-left: 12px !important;
  }
</style>

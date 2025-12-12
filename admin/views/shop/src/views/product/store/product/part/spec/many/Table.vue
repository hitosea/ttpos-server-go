<template>
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
                      :disabled="erp_is_open == 1"
                    >
                      <template v-for="items in restaurantsObj[item.key]">
                        <el-option :value="items.index" :label="items.value"></el-option>
                      </template>
                    </el-select>
                    <el-button size="small" type="primary" :disabled="erp_is_open == 1" class="el-icon-circle-plus mr0" @click="addSku">{{ $t('添加规格') }}+</el-button>
                  </el-form-item>
                </template>
              </div>
            </template>
          </el-table-column>

          <el-table-column v-if="baseSale == '1'" :label="$t('采购单价')" minWidth="160">
            <template #default="scope">
              <el-form-item for="no_click" label="" style="margin-bottom: 0">
                <numInput :disabled="erp_is_open == 1" :min="0" :max="100000000" :precision="2" :placeholder="$t('请输入采购单价')" v-model="scope.row.purchase_price"></numInput>
              </el-form-item>
            </template>
          </el-table-column>
          <!-- 暂时隐藏库存查询 2025年12月12日13:49:13 任务37468 -->
          <!-- <el-table-column :label="'*' + $t('库存')" minWidth="160">
            <template #default="scope">
              <el-form-item
                for="no_click"
                label=""
                style="margin-bottom: 0"
                :prop="`scope.row.stock_num`"
                :rules="[
                  {
                    validator: (rule, value, callback) => {
                      // 检查库存是否为空或null
                      if (scope.row.stock_num == null || scope.row.stock_num === '' || scope.row.stock_num === undefined) {
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
                  :disabled="scope.row.material.length > 0 || erp_is_open == 1"
                  :min="0"
                  :max="99999999"
                  :precision="2"
                  :placeholder="$t('请输入库存')"
                  v-model="scope.row.stock_num"
                ></numInput>
              </el-form-item>
            </template>
          </el-table-column> -->
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
                  :disabled="erp_is_open == 1"
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
                <numInput :disabled="erp_is_open == 1" :min="0" :max="100000000" :precision="2" :placeholder="$t('请输入商品价格')" v-model="scope.row.product_price"></numInput>
              </el-form-item>
            </template>
          </el-table-column>
          <el-table-column v-if="baseSale == '1'" :label="$t('材料')" minWidth="330">
            <template #default="scope">
              <el-form-item for="no_click" label="" style="margin-bottom: 0">
                <el-button
                  type="primary"
                  :disabled="erp_is_open == 1"
                  :style="form.many_select_list[scope.$index].length > 0 ? 'margin-top: 16px;' : ''"
                  @click="addMaterials(scope.$index)"
                  >{{ $t('添加材料') }}</el-button
                >
              </el-form-item>
              <div class="materials-one" label="" v-for="(item, index) in form.many_select_list[scope.$index]">
                <el-form-item for="no_click" label="" class="max-w230">
                  <el-input v-model="item.product_name_text" disabled></el-input>
                </el-form-item>
                <!-- 暂时隐藏库存查询 2025年12月12日13:49:13 任务37468 -->
                <!-- <el-form-item
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
                  <el-input-number
                    :min="0"
                    :max="999"
                    :controls="false"
                    v-model="form.model.sku[scope.$index].material[index].material_num"
                    :placeholder="$t('请输入数量')"
                    :disabled="erp_is_open == 1"
                  >
                  </el-input-number>
                </el-form-item> -->
                <!-- <span class="mt--16">{{ item.product_unit_text }}</span> -->
                <el-icon class="delete-icon" @click="handleDelete(scope.$index, index)">
                  <Delete />
                </el-icon>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="" fixed="right" width="100">
            <template #default="scope">
              <el-form-item label="" style="margin-bottom: 0">
                <el-button type="primary" :disabled="form.model.sku.length <= 1 || scope.row.is_package_used == 1" link @click="deleteAttr(scope.$index)">{{
                  $t('删除')
                }}</el-button>
              </el-form-item>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-form-item>
    <productList
      v-if="open_product"
      :open_product="open_product"
      :index="selectIndex"
      :multiple_selection="multiple_selection"
      :material_type="20"
      @closeDialogFunc="closeDialogFunc($event)"
    >
    </productList>
    <!-- 新增规格 -->
    <Add v-if="open_add" :open_add="open_add" :addform="model" @closeDialog="closeDialog($event)"></Add>
  </div>
</template>

<script setup>
  import { ref, reactive, inject, watch, nextTick } from 'vue';
  import productList from '@/components/productList/productList.vue';
  import { languageStore } from '@/store/model/language.js';
  import { useUserStore } from '@/store';
  import Type from './Type.vue';
  import Add from '../../../../../expand/spec/add.vue';

  // 获取用户信息和语言数据
  const { computedSupplier, erp_is_open } = useUserStore();
  const supplier = computedSupplier().supplier;
  const baseSale = supplier.value?.sale_stock || 0;
  const languageList = languageStore().getLanguageList().languageList.value;
  const languageKey = languageStore().getLanguageKey().language.value;

  // 注入form
  const form = inject('form', {});

  // 初始化语言对象
  let languageObj = {};
  languageList.forEach((item) => {
    languageObj[item.key] = [];
  });

  // 响应式数据
  const restaurantsObj = reactive(languageObj);
  const open_product = ref(false);
  const multiple_selection = ref([]);
  const selectIndex = ref(0);
  const open_add = ref(false);
  const model = ref({});

  // 监听form变化
  watch(
    () => form,
    (val) => {
      let languageObj = {};
      languageList.forEach((item) => {
        languageObj[item.key] = [];
      });
      Object.assign(restaurantsObj, languageObj);
      val.spec.map((item, index) => {
        let spec_name = JSON.parse(item.spec_name);
        languageList.forEach((items) => {
          if (spec_name[items.key] != null) {
            restaurantsObj[items.key].push({
              value: spec_name[items.key] == '' ? '-' : spec_name[items.key],
              index: index,
              spec_id: item.spec_id,
            });
          }
        });
      });

      // 用材料的库存来计算这个规格的库存
      (val.model.sku || []).map((item, index) => {
        //处理条形码
        nextTick(() => {
          if (form.model.sku[index].barcode) {
            //只能输入纯数字
            form.model.sku[index].barcode = item.barcode.match(/[0-9]*/g).join('');
          }
        });

        let arr = [];
        (item.material || []).map((items, indexs) => {
          //有材料的时候4位小数
          if (items.material_num != null) {
            form.model.sku[index].material[indexs].material_num = String(items.material_num).replace(/(\.\d{1,4})\d*/, '$1');
          }

          let num = 0;
          num = Number(form.many_select_list[index][indexs].sku[0]?.material_stock) / Number(items.material_num);
          num = Math.floor(num);
          arr.push(num);
        });
        if ((item.material || []).length > 0) {
          form.model.sku[index].stock_num = arr.sort((a, b) => a - b)[0] == Infinity ? null : arr.sort((a, b) => a - b)[0] > 99999999 ? 99999999 : arr.sort((a, b) => a - b)[0];
        }
      });
    },
    { deep: true, immediate: true }
  );

  // 方法定义
  const deleteAttr = (i) => {
    if (erp_is_open == 1) {
      return;
    }
    if (form.model.sku.length > 1) {
      form.model.sku.splice(i, 1);
      form.many_select_list.splice(i, 1);
      if (i == 0) {
        form.single_select_list = [];
        form.single_select_list = form.many_select_list[0];
      }
    }
  };

  const addSku = () => {
    open_add.value = true;
  };

  const selectChange = (e, index) => {
    languageList.forEach((item) => {
      form.model.sku[index].spec_name[item.key] = restaurantsObj[item.key][e]?.value || '';
      form.model.sku[index].spec_id = restaurantsObj[item.key][e]?.spec_id || '';
    });
    //多规格的时候填入材料
    form.model.sku[index].material = [];
    form.many_select_list[index] = [];
  };

  const addMaterials = (index) => {
    if (erp_is_open == 1) {
      return;
    }
    multiple_selection.value = form.many_select_list[index];
    selectIndex.value = index;
    open_product.value = true;
  };

  const handleDelete = ($index, index) => {
    if (erp_is_open == 1) {
      return;
    }
    form.many_select_list[$index].splice(index, 1);
    form.model.sku[$index].material.splice(index, 1);
    if ($index == 0) {
      form.single_select_list.splice(index, 1);
    }
  };

  const closeDialogFunc = (e) => {
    open_product.value = e.openDialog;
    if (e.type == 'select') {
      let map = new Map();
      if (e.index == 0) {
        [form.single_select_list, e.data].flat().forEach((obj) => map.set(obj.product_id, obj));
        form.single_select_list = Array.from(map.values());
      }

      [form.many_select_list[e.index], e.data].flat().forEach((obj) => map.set(obj.product_id, obj));
      form.many_select_list[e.index] = Array.from(map.values());

      let arr = [];
      if (form.model.sku[e.index].material.length > 0) {
        form.model.sku[e.index].material.map((item) => {
          arr.push(item.product_id);
        });
      }

      form.many_select_list[e.index].map((item) => {
        if (!arr.includes(item.product_id)) {
          form.model.sku[e.index].material.push({
            product_id: item.product_id,
            material_num: null,
          });
        }
      });
    }
  };

  const closeDialog = (e) => {
    /*关闭弹窗*/
    open_add.value = e.openDialog;
    if (e.type == 'success' && e.data) {
      form.spec.unshift(e.data);
    }
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

<template>
  <div>
    <template v-if="form.model.type == 10" v-for="(item, index) in languageList">
      <el-form-item
        for="no_click"
        :label="index == 0 ? $t('规格名称：') : ''"
        :prop="`model.sku[${0}].spec_name[${item.key}]`"
        :rules="[{ required: true, message: $t('请输入规格名称') }]"
      >
        <mAutocomplete
          v-model:valueData="form.model.sku[0].spec_name[item.key]"
          :value="form.model.sku[0].spec_name[item.key]"
          @select="(e) => selectChange(e, 0)"
          :numKey="item.key"
          :restaurantsObj="restaurantsObj"
          :placeholder="$t('请输入') + `(${item.value})`"
          @translate="(e) => translate(e)"
        ></mAutocomplete>
      </el-form-item>
    </template>

    <el-form-item
      v-if="form.model.type == 10"
      for="no_click"
      :label="$t('商品价格：')"
      width="80"
      :rules="[{ required: true, message: $t('请填写商品价格') }]"
      prop="model.sku[0].product_price"
    >
      <numInput :min="0" :max="100000000" v-model="form.model.sku[0].product_price" :placeholder="$t('请填写商品价格')" class="max-w460"></numInput>
    </el-form-item>

    <el-form-item for="no_click" v-if="baseSale == '1'" :label="$t('采购单价：')" width="80">
      <numInput :min="0" :max="100000000" v-model="form.model.sku[0].purchase_price" :placeholder="$t('请填写采购单价')" class="max-w460"></numInput>
    </el-form-item>

    <el-form-item for="no_click" :label="$t('库存数量：')" v-if="form.model.type == 10" :rules="[{ required: true, message: $t('请填写库存数量') }]" prop="model.sku[0].stock_num">
      <numInput
        :min="0"
        :max="99999999"
        :disabled="form.single_select_list.length > 0"
        :placeholder="$t('请填写库存数量')"
        v-model="form.model.sku[0].stock_num"
        class="max-w460"
      ></numInput>
    </el-form-item>
    <el-form-item
      for="no_click"
      :label="$t('库存数量：')"
      v-if="form.model.type == 20"
      :rules="[{ required: true, message: $t('请填写库存数量') }]"
      prop="model.sku[0].material_stock"
    >
      <numInput
        :min="0"
        :max="99999999"
        :disabled="form.single_select_list.length > 0"
        :placeholder="$t('请填写库存数量')"
        v-model="form.model.sku[0].material_stock"
        class="max-w460"
      ></numInput>
    </el-form-item>

    <el-form-item
      for="no_click"
      prop="model.sku[0].barcodeUniqueness"
      :rules="[
        {
          validator: () => {
            return form.model.sku[0].barcodeUniqueness ? true : false;
          },
          message: $t('商品条码重复'),
          trigger: 'change',
        },
      ]"
      v-if="baseSale == '1'"
      :label="$t('商品条码：')"
    >
      <el-input
        :placeholder="$t('请输入商品条码')"
        @input="
          () => {
            form.model.sku[0].barcodeUniqueness = true;
          }
        "
        v-model="form.model.sku[0].barcode"
        class="max-w460"
      ></el-input>
    </el-form-item>
    <el-form-item for="no_click" class="materials" :label="$t('商品材料：')" v-if="form.model.type == 10 && baseSale == '1'">
      <el-button style="margin-bottom: 16px" type="primary" @click="addMaterials">{{ $t('添加材料') }}+</el-button>

      <div class="materials-one" label="" v-for="(item, index) in form.single_select_list">
        <el-form-item label="" class="max-w230">
          <el-input v-model="item.product_name_text" disabled></el-input>
        </el-form-item>
        <el-form-item
          label=""
          class="max-w230"
          :rules="[
            {
              validator: () => {
                return form.model.sku[0].material[index].material_num && form.model.sku[0].material[index].material_num != 0 ? true : false;
              },
              message: $t('请输入数量'),
            },
          ]"
          :prop="`form.model.sku[0].material[${index}].material_num`"
        >
          <numInput :min="0" :max="999" v-model="form.model.sku[0].material[index].material_num" :placeholder="$t('请输入数量')"> </numInput>
        </el-form-item>
        <span class="mt--16">{{ item.product_unit_text }}</span>
        <el-icon class="delete-icon" @click="handleDelete(index)">
          <Delete />
        </el-icon>
      </div>
    </el-form-item>
    <productList
      v-if="open_product"
      :open_product="open_product"
      :index="0"
      :multiple_selection="multiple_selection"
      :material_type="20"
      @closeDialogFunc="closeDialogFunc($event)"
    >
    </productList>
  </div>
</template>

<script setup>
  import { ref, reactive, inject, watch, nextTick } from 'vue';
  import productList from '@/components/productList/productList.vue';
  import { useUserStore } from '@/store';
  import { languageStore } from '@/store/model/language.js';
  import mAutocomplete from '@/components/m-autocomplete/index.vue';

  // 获取用户信息和语言数据
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const baseSale = supplier.value?.sale_stock || 0;
  const languageList = languageStore().getLanguageList().languageList.value;

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
  const oldStockNum = ref(0);
  const action = ref(0);

  // 监听form.model.sku变化
  watch(
    () => form.model.sku,
    (val) => {
      if (val) {
        let arr = [];
        (val[0]?.material || []).map((item, index) => {
          let num = 0;
          num = Number(form.single_select_list[index].sku[0].material_stock) / Number(item.material_num);
          num = Math.floor(num);
          arr.push(num);
        });
        if ((val[0].material || []).length > 0) {
          form.model.sku[0].stock_num = arr.sort((a, b) => a - b)[0] == Infinity ? null : arr.sort((a, b) => a - b)[0] > 99999999 ? 99999999 : arr.sort((a, b) => a - b)[0];
        }

        //处理条形码
        (val || []).map((item, index) => {
          nextTick(() => {
            if (form.model.sku[index].barcode) {
              form.model.sku[index].barcode = item.barcode.match(/[a-zA-Z0-9]*/g).join('');
            }
          });
        });

        // 单规格成品库存取整
        if (form.single_select_list.length == 0 && form.model.type == 10) {
          nextTick(() => {
            form.model.sku[0].stock_num = Number(String(form.model.sku[0].stock_num).replace(/(\.\d{0,0})\d*/, '$1'));
          });
        }

        // 材料库存4位小数
        if (form.single_select_list.length == 0 && form.model.type == 20) {
          nextTick(() => {
            form.model.sku[0].material_stock = Number(String(form.model.sku[0].material_stock).replace(/(\.\d{1,4})\d*/, '$1'));
          });
        }

        // 单规格材料数4位小数
        if (form.single_select_list.length > 0 && form.model.type == 10) {
          nextTick(() => {
            form.model.sku[0].material.map((item, index) => {
              if (item.material_num != null) {
                form.model.sku[0].material[index].material_num = String(item.material_num).replace(/(\.\d{1,4})\d*/, '$1');
              }
            });
          });
        }
        //记录旧的库存数据
        action.value++;
        if (action.value == 1) {
          oldStockNum.value = JSON.parse(JSON.stringify(form.model.sku[0].stock_num));
        }
      }
    },
    { deep: true, immediate: true }
  );

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
          if (spec_name[items.key]) {
            restaurantsObj[items.key].push({
              value: spec_name[items.key],
              index: index,
            });
          }
        });
      });
    },
    { deep: true, immediate: true }
  );

  // 方法定义
  const addMaterials = () => {
    multiple_selection.value = form.single_select_list;
    open_product.value = true;
  };

  const handleDelete = (index) => {
    form.single_select_list.splice(index, 1);
    form.many_select_list[0].splice(index, 1);
    form.model.sku[0].material.splice(index, 1);
    //当没材料是回复初始库存
    if (form.single_select_list.length == 0 && form.model.type == 10) {
      form.model.sku[0].stock_num = JSON.parse(JSON.stringify(oldStockNum.value));
    }
  };

  const closeDialogFunc = (e) => {
    open_product.value = e.openDialog;
    if (e.type == 'select') {
      let map = new Map();
      [form.single_select_list, e.data].flat().forEach((obj) => map.set(obj.product_id, obj));
      form.single_select_list = Array.from(map.values());

      form.many_select_list[0] = [];
      form.many_select_list[0] = JSON.parse(JSON.stringify(form.single_select_list));

      let arr = [];
      form.model.sku[0].material.map((item) => {
        arr.push(item.product_id);
      });

      form.single_select_list.map((item) => {
        if (!arr.includes(item.product_id)) {
          form.model.sku[0].material.push({
            product_id: item.product_id,
            material_num: null,
          });
        }
      });
    }
  };

  const selectChange = (e, index) => {
    languageList.forEach((item) => {
      form.model.sku[index].spec_name[item.key] = restaurantsObj[item.key][e.index]?.value || '';
    });
  };

  /*翻译*/
  const translate = (lang) => {
    languageList.map((item) => {
      lang.map((items) => {
        let key = item.name;
        if (key == 'zhtw') {
          key = 'zh-TW';
        }
        if (items[key]) {
          form.model.sku[0].spec_name[item.key] = items[key];
        }
      });
    });
  };
</script>

<style lang="scss" scoped>
  :deep(.el-input__wrapper) {
    padding-left: 7px !important;
    padding-right: 7px !important;
  }

  .materials {
    :deep(.el-form-item__content) {
      flex-direction: column;
      align-items: flex-start;
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
  }

  .max-w230 {
    max-width: 226px;
    width: 100%;
  }
  .mt--16 {
    margin-top: -16px !important;
    flex-shrink: 0;
  }
</style>

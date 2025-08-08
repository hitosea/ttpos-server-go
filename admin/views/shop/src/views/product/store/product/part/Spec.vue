<template>
  <div>
    <!--规格设置-->
    <div class="common-form mt50">{{ $t('规格/库存') }}</div>

    <!--减库存方式-->
    <el-form-item for="no_click" :label="$t('库存计算方式：')" v-if="form.model.type == 10">
      <el-radio-group v-model="form.model.deduct_stock_type">
        <el-radio :value="10">{{ $t('下单减库存') }}</el-radio>
        <el-radio :value="20">{{ $t('付款减库存') }}</el-radio>
      </el-radio-group>
    </el-form-item>

    <template v-for="(item, index) in languageList" :key="index">
      <el-form-item
        for="no_click"
        :label="$t('商品单位：')"
        :rules="[{ required: true, message: $t('请填写商品单位') }]"
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
        >
          <template v-for="items in restaurantsObj[item.key]" :key="items.index">
            <el-option :value="items.index" :label="items.value"></el-option>
          </template>
        </el-select>
        <el-button size="small" type="primary" class="el-icon-circle-plus" @click="addUnit">{{ $t('添加单位') }}+</el-button>
      </el-form-item>
    </template>

    <!--单规格-->
    <template v-if="form.model.type == 20">
      <Single></Single>
    </template>

    <!--多规格-->
    <template v-if="form.model.type == 10">
      <Many></Many>
    </template>
    <!--添加-->
    <Add v-if="open_add_feed" :open_add="open_add_feed" @closeDialog="closeDialogFunc($event, 'add')"></Add>
  </div>
</template>

<script setup>
  import { ref, reactive, inject, watch } from 'vue';
  import Single from './spec/Single.vue';
  import Many from './spec/Many.vue';
  import { languageStore } from '@/store/model/language.js';
  import Add from '../../../expand/unit/add.vue';

  // 获取语言数据
  const languageList = languageStore().getLanguageList().languageList.value;
  const languageKey = languageStore().getLanguageKey().language.value;

  // 注入form
  const form = inject('form');

  // 响应式数据
  const restaurants = ref([]);
  const open_add_feed = ref(false);

  // 初始化语言对象
  let languageObj = {};
  languageList.forEach((item) => {
    languageObj[item.key] = [];
  });

  const restaurantsObj = reactive(languageObj);

  // 工具函数定义
  const isValidJSON = (str) => {
    try {
      JSON.parse(str);
      return true; // 如果解析成功，返回 true
    } catch (e) {
      return false; // 如果解析失败，返回 false
    }
  };

  // 监听form变化
  watch(
    () => form,
    (val) => {
      let languageObj = {};
      languageList.forEach((item) => {
        languageObj[item.key] = [];
      });
      Object.assign(restaurantsObj, languageObj);

      val.unit.map((item, index) => {
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

  // 方法定义
  const selectChange = (e) => {
    languageList.forEach((item) => {
      form.model.product_unit[item.key] = restaurantsObj[item.key][e]?.value || '';
      form.model.unit_id = restaurantsObj[item.key][e]?.unit_id || '';
    });
  };

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
</script>

<style scoped lang="scss">
  :deep(.inline-input) {
    max-width: 460px;
    width: 100%;
  }
</style>

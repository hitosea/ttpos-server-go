<template>
  <!--添加规格-->
  <el-button v-if="!form.isSpecLocked" size="small" type="primary" :disabled="erp_is_open == 1" class="el-icon-circle-plus" @click="onToggleAddGroupForm">{{ $t('添加规格') }}+</el-button>
</template>

<script setup>
  import { inject } from 'vue';
  import { languageStore } from '@/store/model/language.js';
  import { useUserStore } from '@/store';
  const { erp_is_open } = useUserStore();
  // 获取语言数据
  const languageData = JSON.stringify(languageStore().getLanguageKeyForm());

  // 注入form
  const form = inject('form', {});

  // 方法定义
  /*显示/隐藏添加规则组 */
  const onToggleAddGroupForm = () => {
    if (erp_is_open == 1) {
      return;
    }
    if (!Array.isArray(form.model.sku)) {
      form.model.sku = [];
    }
    form.model.sku.push({
      spec_name: JSON.parse(languageData),
      product_price: null,
      stock_num: null,
      barcode: '', //条形码
      purchase_price: null, //单价
      material: [], //材料
      spec_id: null, //材料库存
      barcodeUniqueness: true, //条形码是否唯一
    });
    form.many_select_list.push([]);
  };
</script>

<style scoped>
  .spec-many-type {
    margin-top: 16px;
    padding: 20px;
    border: 1px solid #e5ecf4;
    background: #f6f9fc;
  }

  .spec-wrap .spec-hd {
    padding: 10px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: #fff;
    font-weight: bold;
  }

  .spec-wrap .spec-hd .el-icon-delete-solid {
    font-size: 16px;
    color: #999999;
  }

  .spec-wrap .min-spc {
    border: 1px solid #dfecf8;
  }

  .spec-wrap .spec-bd {
    padding: 5px;
    display: flex;
    justify-content: flex-start;
    flex-wrap: wrap;
    border-top: 1px solid #dfecf8;
    background: #ffffff;
  }

  .spec-wrap .spec-bd .el-tag {
    color: #333333;
  }

  .spec-wrap .spec-bd .item {
    position: relative;
    padding: 5px;
  }

  .spec-wrap .spec-bd .item input {
    padding-right: 30px;
  }

  .spec-wrap .spec-hd a,
  .spec-wrap .spec-hd .svg-icon,
  .spec-wrap .spec-bd .item .svg-icon {
    display: block;
    width: 16px;
    height: 16px;
  }

  .spec-wrap .spec-bd .item a {
    position: absolute;
    top: 6px;
    right: 5px;
    width: 30px;
    height: 30px;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .add-spec .from-box {
    display: flex;
    justify-content: flex-start;
  }

  .add-spec .item {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    width: 200px;
    margin-right: 20px;
  }

  .add-spec .item .key {
    display: block;
    white-space: nowrap;
  }
</style>

<template>
  <template v-for="(item, index) in form.productTaxes">
    <el-form-item
      for="no_click"
      :label="returnType(item.product_tax_type)"
      :prop="`form.productTaxes[${index}].tax_category_id`"
      :rules="[
        {
          required: true,
          validator: () => {
            return item.tax_category_id ? true : false;
          },
          message: returnMessage(item.product_tax_type),
        },
      ]"
    >
      <el-select v-model="item.tax_category_id" clearable class="max-w460" size="default" :placeholder="$t('请选择')">
        <template v-for="cat in taxList" :key="cat.id">
          <el-option :value="cat.id" :label="cat.name"></el-option>
        </template>
      </el-select>
    </el-form-item>
  </template>
</template>
<script>
  import PorductApi from '@/api/product.js';
  export default {
    name: 'taxChange',
    data() {
      return {
        taxList: [],
      };
    },
    inject: ['form'],
    created() {
      this.getTaxData();
    },
    methods: {
      /*获取基础数据*/
      getTaxData: function () {
        PorductApi.getTaxList({}, true)
          .then((res) => {
            this.taxList = res.data.list;
            let idArr = [];
            this.taxList.map((item) => {
              idArr.push(item.id);
            });
            this.form.productTaxes.map((item) => {
              if (!idArr.includes(item.tax_category_id)) {
                item.tax_category_id = '';
              }
            });
          })
          .catch((error) => {});
      },

      returnType(type) {
        let result = '';
        if (type == '1') {
          result = $t('堂食税类：');
        } else {
          result = $t('外带税类：');
        }
        return result;
      },
      returnMessage(type) {
        let result = '';
        if (type == '1') {
          result = $t('请选择堂食税类');
        } else {
          result = $t('请选择外带税类');
        }
        return result;
      },

      submit() {
        let self = this;
        self.loading = true;
        self.$emit('loading', true);
        const data = {
          productTaxes: this.form.productTaxes,
          product_ids: this.form.product_ids,
        };
        PorductApi.batchUpdateTax(data, true)
          .then((data) => {
            self.loading = false;
            self.$emit('loading', false);
            this.$ElMessage({
              type: 'success',
              message: this.$t('操作成功'),
            });
            this.$emit('close');
          })
          .catch((error) => {
            self.loading = false;
            self.$emit('loading', false);
          });
      },
    },
  };
</script>
<style lang=""></style>

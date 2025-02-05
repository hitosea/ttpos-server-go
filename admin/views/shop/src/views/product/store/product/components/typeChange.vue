<template>
  <el-form-item
    :label="$t('所属分类')"
    for="no_click"
    prop="category_id"
    :rules="[
      {
        validator: () => {
          return form.category_id ? true : false;
        },
        required: true,
        message: $t('请选择分类'),
      },
    ]"
  >
    <el-cascader class="max-w460 mr8" :options="options" v-model="form.category_id" clearable style="width: 100%" :placeholder="$t('请选择分类')"></el-cascader>
    <el-button type="primary" size="small" :loading="loading" @click="add">{{ $t('添加分类') }} </el-button>
  </el-form-item>

  <!--添加-->
  <Add v-if="open_add" :open_add="open_add" @closeDialog="closeDialogFunc($event, 'add')"> </Add>
</template>
<script>
  import PorductApi from '@/api/product.js';
  import Add from '../../category/Add.vue';
  export default {
    components: {
      Add,
    },
    inject: ['form'],
    data() {
      return {
        loading: false,
        open_add: false,
        options: [],
      };
    },
    created() {
      /*获取列表*/
      this.getData();
    },
    methods: {
      /*获取列表*/
      getData() {
        let self = this;
        self.loading = true;
        PorductApi.storeCatList(
          {
            page: 1,
            list_rows: 1000,
          },
          true
        )
          .then(async (data) => {
            self.loading = false;
            this.options = [];
            await data.data.list.data.map((item) => {
              if (item.category_id != '0') {
                this.options.push({
                  value: item.category_id,
                  label: item.name_text,
                  children: [],
                });
              }
            });
            data.data.list.data.map((item, index) => {
              if (item.child && item.child.length > 0) {
                item.child.map((items) => {
                  this.options[index].children.push({
                    value: items.category_id,
                    label: items.name_text,
                  });
                });
              }
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      submit() {
        let self = this;
        self.loading = true;
        self.$emit('loading', true);
        const data = {
          category_id: this.form.category_id[this.form.category_id.length - 1],
          product_ids: this.form.product_ids,
        };
        PorductApi.batchUpdateCategory(data, true)
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

      add() {
        this.open_add = true;
      },

      /*关闭弹窗*/
      closeDialogFunc(e, f) {
        if (f == 'add') {
          this.open_add = e.openDialog;
          if (e.type == 'success') {
            this.getData();
          }
        }
      },
    },
  };
</script>
<style lang="scss" scoped></style>

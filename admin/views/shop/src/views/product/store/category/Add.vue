<template>
  <el-dialog :title="$t('添加普通分类')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false" append-to-body>
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <el-form-item for="no_click" :label="$t('分类级别')" prop="parent">
        <el-radio-group v-model="parent" @change="radioChange">
          <el-radio :label="1">{{ $t('一级分类') }}</el-radio>
          <el-radio :label="0">{{ $t('二级分类') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item for="no_click" v-if="parent == 0" :label="$t('上级分类')" prop="parent_id" :rules="[{ required: true, message: $t('请选择上级分类') }]">
        <el-select v-model="form.parent_id" :placeholder="$t('请选择上级分类')">
          <template v-for="cat in category" :key="cat.category_id">
            <el-option :value="cat.category_id" :label="cat.name_text"></el-option>
          </template>
        </el-select>
      </el-form-item>

      <UniqueNameForm ref="uniqueNameFormRef" :labelPrefix="$t('分类名称')" apiSource="category" />

      <el-form-item
        for="no_click"
        :label="$t('分类排序')"
        prop="sort"
        :rules="[
          {
            required: true,
            message: $t('分类排序不能为空'),
          },
          {
            type: 'number',
            message: $t('分类排序必须为数字'),
          },
        ]"
      >
        <el-input-number
          :controls="false"
          :precision="0"
          :placeholder="$t('接近0，排序等级越高')"
          :min="0"
          :max="999"
          v-model.number="form.sort"
          autocomplete="off"
        ></el-input-number>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-tooltip effect="dark" placement="top" :content="$t('如长时间无响应，请刷新后重试。')">
          <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
        </el-tooltip>
      </div>
    </template>
  </el-dialog>
</template>

<script>
  import ProductApi from '@/api/product.js';
  import mInput from '@/components/m-input/index.vue';
  import { languageStore } from '@/store/model/language.js';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';

  const languageData = JSON.stringify(languageStore().getLanguageData().languageData.value);

  export default {
    name: 'ProductCategoryAdd',
    components: {
      mInput,
      UniqueNameForm,
    },
    data() {
      return {
        category: [],
        parent: 1,
        form: {
          parent_id: 0,
          category_id: 0,
          name: JSON.parse(languageData),
          sort: null,
        },
        /*左边长度*/
        /*是否显示*/
        dialogVisible: false,
        loading: false,
      };
    },
    props: ['open_add', 'addform'],
    created() {
      this.dialogVisible = this.open_add;
      /*获取父级分类*/
      this.getParentCategory();
    },
    methods: {
      /*获取父级分类*/
      getParentCategory: function () {
        const self = this;
        ProductApi.storeCatParentList({}, true)
          .then((res) => {
            self.loading = false;
            self.category = res.data.list;
          })
          .catch(() => {
            self.loading = false;
          });
      },

      async submit() {
        let self = this;
        self.loading = true;
        try {
          const validUniqueName = await self.$refs.uniqueNameFormRef.validate();
          const validForm = await self.$refs.formRef.validate();

          if (!validUniqueName || !validForm) return;
          const params = JSON.parse(JSON.stringify(self.form));
          if (self.parent === 1) {
            params.parent_id = '';
          }

          const _name = self.$refs.uniqueNameFormRef.data;
          params.name = JSON.stringify(_name);
          const res = await ProductApi.storeCatAdd(params, true);
          self.$ElMessage({
            message: self.$t('保存成功'),
            type: 'success',
          });
          self.dialogFormVisible(true, res.data);
        } catch (error) {
          //
        } finally {
          self.loading = false;
        }
      },

      /*关闭弹窗*/
      dialogFormVisible(e, data) {
        if (e) {
          this.$emit('closeDialog', {
            type: 'success',
            data: data,
            openDialog: false,
          });
        } else {
          this.$emit('closeDialog', {
            type: 'error',
            openDialog: false,
          });
        }
      },

      radioChange() {
        this.form.parent_id = '';
      },
    },
  };
</script>

<style scoped>
  .img {
    margin-top: 10px;
  }
</style>

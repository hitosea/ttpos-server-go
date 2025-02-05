<template>
  <el-dialog :title="$t('添加特色分类')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" ref="formRef">
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
    name: 'ProductFeatureAdd',
    components: {
      mInput,
      UniqueNameForm,
    },
    data() {
      return {
        category: [],
        form: {
          parent_id: 0,
          category_id: 0,
          name: JSON.parse(languageData),
          sort: null,
          is_special: 1,
        },
        /*是否显示*/
        dialogVisible: false,
        loading: false,
      };
    },
    props: ['open_add', 'addform'],
    created() {
      this.dialogVisible = this.open_add;
    },
    methods: {
      async submit() {
        const self = this;
        self.loading = true;
        try {
          const validUniqueName = await self.$refs.uniqueNameFormRef.validate();
          const validForm = await self.$refs.formRef.validate();

          if (!validUniqueName || !validForm) return;

          const params = JSON.parse(JSON.stringify(self.form));

          const _name = self.$refs.uniqueNameFormRef.data;
          params.name = JSON.stringify(_name);
          await ProductApi.storeCatAdd(params, true);
          self.$ElMessage({
            message: self.$t('保存成功'),
            type: 'success',
          });
          self.dialogFormVisible(true);
        } catch (error) {
          //
        } finally {
          self.loading = false;
        }
      },

      /*关闭弹窗*/
      dialogFormVisible(e) {
        if (e) {
          this.$emit('closeDialog', {
            type: 'success',
            openDialog: false,
          });
        } else {
          this.$emit('closeDialog', {
            type: 'error',
            openDialog: false,
          });
        }
      },
    },
  };
</script>

<style scoped>
  .img {
    margin-top: 10px;
  }
</style>

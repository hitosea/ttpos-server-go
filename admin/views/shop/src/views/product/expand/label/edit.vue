<!-- 描述：商品-打印标签-编辑 -->
<template>
  <el-dialog :title="$t('编辑标签')" v-model="dialogVisible" @close="handleClose" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <UniqueNameForm
        ref="uniqueNameFormRef"
        :single-language="true"
        :labelPrefix="$t('标签名称')"
        :placeholder="$t('请输入标签名称')"
        apiSource="label"
        :apiId="form.label_id"
        :overrideLanguages="{ SINGLE: form.label_name }"
      />
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">{{ $t('取消') }}</el-button>
        <el-tooltip effect="dark" placement="top" :content="$t('如长时间无响应，请刷新后重试。')">
          <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
        </el-tooltip>
      </div>
    </template>
  </el-dialog>
</template>

<script>
  import ProductApi from '@/api/product.js';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';

  export default {
    name: 'ProductLabelEditDialog',
    components: {
      UniqueNameForm,
    },
    data() {
      return {
        form: {
          label_id: undefined,
          label_name: '',
        },
        /*是否显示*/
        dialogVisible: false,
        loading: false,
      };
    },
    props: ['open_edit', 'editform'],
    created() {
      this.dialogVisible = this.open_edit;

      this.form.label_id = this.editform.label_id;
      this.form.label_name = this.editform.label_name;
    },
    methods: {
      async submit() {
        const self = this;
        self.loading = true;
        try {
          const validForm = await self.$refs.formRef.validate();
          if (!validForm) return;

          const validUniqueName = await self.$refs.uniqueNameFormRef.validate();
          if (!validUniqueName) return;

          const _name = self.$refs.uniqueNameFormRef.data;
          const params = JSON.parse(JSON.stringify(self.form));
          params.label_name = _name?.SINGLE ?? '';

          const res = await ProductApi.editLabel(params, true);
          self.$ElMessage({
            message: self.$t('保存成功'),
            type: 'success',
          });

          self.handleClose(true, res.data);
        } catch (error) {
          console.error(error);
        } finally {
          self.loading = false;
        }
      },
      /*关闭弹窗*/
      handleClose(isSuccess = false, data) {
        this.$emit('closeDialog', {
          type: isSuccess ? 'success' : 'error',
          openDialog: false,
          data: data,
        });
      },
    },
  };
</script>

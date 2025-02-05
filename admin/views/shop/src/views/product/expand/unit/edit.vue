<!-- 描述：商品-单位库-编辑单位 -->
<template>
  <el-dialog :title="$t('编辑单位')" v-model="dialogVisible" @close="handleClose" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <UniqueNameForm ref="uniqueNameFormRef" :labelPrefix="$t('单位名称')" apiSource="unit" :apiId="form.unit_id" :overrideLanguages="form.unit_name" />
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
    name: 'ProductUnitEditDialog',
    components: {
      UniqueNameForm,
    },
    data() {
      return {
        form: {
          unit_id: undefined,
          unit_name: {},
        },
        /*是否显示*/
        dialogVisible: false,
        loading: false,
      };
    },
    props: ['open_edit', 'editform'],
    created() {
      this.dialogVisible = this.open_edit;

      this.form.unit_id = this.editform.unit_id;
      try {
        const _names = typeof this.editform.unit_name === 'string' ? JSON.parse(this.editform.unit_name) : this.editform.unit_name ?? {};
        this.form.unit_name = _names;
      } catch (error) {
        console.error('parse name faild', error);
      }
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
          params.unit_name = JSON.stringify(_name);

          const res = await ProductApi.editUnit(params, true);
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

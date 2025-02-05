<template>
  <el-dialog :title="$t('添加区域')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" :rules="formRules" ref="formRef">
      <el-form-item
        for="no_click"
        :label="$t('区域名称')"
        prop="area_name"
        :label-width="formLabelWidth"
        :rules="[
          {
            required: true,
            message: $t('请输入区域名称'),
          },
          {
            validator: uniqueNameValidator('table_area', undefined, 'SINGLE'),
            trigger: 'blur',
          },
        ]"
      >
        <el-input :maxlength="50" v-model="form.area_name" autocomplete="off" :placeholder="$t('请输入区域名称')"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="addUser" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script>
  import StoreApi from '@/api/store.js';
  import { uniqueNameValidator } from '@/utils/form.js';

  export default {
    name: 'SupplierTableAreaAdd',
    components: {},
    data() {
      return {
        form: {
          area_name: '',
        },
        /*是否显示*/
        dialogVisible: false,
        loading: false,
        /*是否上传图片*/
        isupload: false,
      };
    },
    props: ['open_add', 'addform'],
    created() {
      this.dialogVisible = this.open_add;
    },
    methods: {
      /*添加用户*/
      addUser() {
        let self = this;
        let params = self.form;
        self.$refs.formRef.validate((valid) => {
          if (valid) {
            self.loading = true;
            StoreApi.addArea(params, true)
              .then(() => {
                self.loading = false;
                self.$ElMessage({
                  message: self.$t('添加成功'),
                  type: 'success',
                });
                self.dialogFormVisible(true);
              })
              .catch(() => {
                self.loading = false;
              });
          }
        });
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
      uniqueNameValidator: uniqueNameValidator,
    },
  };
</script>

<style scoped>
  .img {
    margin-top: 10px;
  }
</style>

<template>
  <el-dialog :title="$t('修改区域')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
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
            validator: uniqueNameValidator('table_area', form.area_id, 'SINGLE'),
            trigger: 'blur',
          },
        ]"
      >
        <el-input :maxlength="50" v-model="form.area_name" autocomplete="off" :placeholder="$t('请输入区域名称')"></el-input>
      </el-form-item>
      <!-- <el-form-item for="no_click" :label="$t('排序')" prop="sort" :label-width="formLabelWidth">
                <el-input-number :controls="false" :min="0" :max="999" :placeholder="$t('接近0，排序等级越高')" v-model.number="form.sort"></el-input-number>
            </el-form-item> -->
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
    name: 'SupplierTableAreaEdit',
    components: {},
    data() {
      return {
        form: {
          area_id: null,
          area_name: '',
        },
        /*是否显示*/
        dialogVisible: false,
        loading: false,
        /*是否上传图片*/
        isupload: false,
      };
    },
    props: ['open_edit', 'editform'],
    created() {
      this.dialogVisible = this.open_edit;
      this.form.area_id = this.editform.model.area_id;
      this.form.area_name = this.editform.model.area_name;
    },
    methods: {
      /*修改用户*/
      addUser() {
        let self = this;
        let params = self.form;
        self.$refs.formRef.validate((valid) => {
          if (valid) {
            self.loading = true;
            StoreApi.editArea(params, true)
              .then(() => {
                self.loading = false;
                self.$ElMessage({
                  message: self.$t('保存成功'),
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

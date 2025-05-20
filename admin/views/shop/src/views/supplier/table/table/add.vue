<template>
  <el-dialog :title="$t('添加桌位')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" :rules="formRules" ref="formRef">
      <el-form-item
        for="no_click"
        :label="$t('桌位名称')"
        prop="table_no"
        :label-width="formLabelWidth"
        :rules="[
          {
            required: true,
            message: this.$t('请输入桌位名称'),
          },
          {
            validator: uniqueNameValidator('table', undefined, 'SINGLE'),
            trigger: 'blur',
          },
        ]"
      >
        <el-input :maxlength="50" v-model="form.table_no" autocomplete="off" :placeholder="$t('请输入桌位名称')"></el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('所属类型')" prop="type_id" :label-width="formLabelWidth">
        <el-select v-model="form.type_id" :placeholder="$t('所属类型')">
          <el-option v-for="item in type" :key="item.type_id" :label="item.type_name" :value="item.type_id"> </el-option>
        </el-select>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('所属区域')" prop="area_id" :label-width="formLabelWidth">
        <el-select v-model="form.area_id" :placeholder="$t('所属区域')">
          <el-option v-for="item in area_list" :key="item.area_id" :label="item.area_name" :value="item.area_id"> </el-option>
        </el-select>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('排序')" prop="sort" :label-width="formLabelWidth">
        <el-input-number :controls="false" :min="0" :max="999" :placeholder="$t('接近0，排序等级越高')" v-model.number="form.sort"></el-input-number>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('默认开桌人数')" :label-width="formLabelWidth">
        <el-radio-group v-model="form.is_open_default_people_num">
          <el-radio :label="1">{{ $t('开启') }}</el-radio>
          <el-radio :label="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
        <template v-if="form.is_open_default_people_num == 1">
          <el-input-number
            class="mt4"
            :controls="false"
            :min="1"
            :max="99"
            :precision="0"
            :placeholder="$t('请输入默认开桌人数')"
            v-model.number="form.default_people_num"
          ></el-input-number>
          <div class="tips">{{ $t('默认桌台人数仅非自助餐类型生效') }}</div>
        </template>
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
    name: 'SupplierTableAdd',
    components: {},
    data() {
      return {
        form: {
          table_no: '',
          area_id: '',
          type_id: '',
          sort: null,
          is_open_default_people_num: 0,
          default_people_num: null,
        },
        formRules: {
          area_id: [
            {
              required: true,
              message: this.$t('请选择所属区域'),
              trigger: 'change',
            },
          ],
          type_id: [
            {
              required: true,
              message: this.$t('请选择类型名称'),
              trigger: 'change',
            },
          ],
          sort: [
            { required: true, message: this.$t('排序不能为空') },
            { type: 'number', message: this.$t('排序必须为数字') },
            { type: 'number', min: 0, message: this.$t('请输入不小于0的数字'), trigger: 'blur' },
          ],
        },
        /*左边长度*/
        formLabelWidth: '120px',
        /*是否显示*/
        dialogVisible: false,
        loading: false,
        /*是否上传图片*/
        isupload: false,
      };
    },
    props: ['open_add', 'addform', 'type', 'area_list'],
    created() {
      this.dialogVisible = this.open_add;
    },
    methods: {
      /*添加用户*/
      addUser() {
        const self = this;
        const params = self.form;
        if (self.form.default_people_num < 1) {
          self.$ElMessage({
            message: self.$t('默认开桌人数必须大于等于1'),
            type: 'error',
          });
          return;
        }
        self.$refs.formRef.validate((valid) => {
          if (valid) {
            self.loading = true;
            StoreApi.addTable(params, true)
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

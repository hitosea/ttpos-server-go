<template>
  <!--
      
      时间：2019-10-25
      描述：会员-用户列表-会员等级
  -->
  <el-dialog :title="$t('扣减')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" ref="formRef" :model="form" label-position="top">
      <el-form-item for="no_click" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.source" @change="changeSource">
          <el-radio label="0">{{ $t('赠送账户余额') }}</el-radio>
          <el-radio label="1">{{ $t('积分') }}</el-radio>
          <el-radio label="2">{{ $t('主账户余额') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item for="no_click" :label="labelText">
        <el-input v-if="form.source == '0'" v-model="editData.gift_balance" autocomplete="off" disabled></el-input>
        <el-input v-else-if="form.source == '1'" v-model="editData.points" autocomplete="off" disabled></el-input>
        <el-input v-else-if="form.source == '2'" v-model="editData.balance" autocomplete="off" disabled></el-input>
      </el-form-item>

      <el-form-item
        for="no_click"
        :label="form.source != '1' ? $t('扣减金额') : $t('扣减数量')"
        prop="value"
        :rules="[{ required: true, message: form.source == '0' ? $t('请输入扣减金额') : $t('请输入扣减数量') }]"
      >
        <el-input-number
          :controls="false"
          :min="0"
          :max="maxValue"
          :placeholder="form.source != '1' ? $t('请输入扣减金额') : $t('请输入扣减数量')"
          :precision="2"
          v-model.number="form.value"
        ></el-input-number>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('备注')">
        <el-input type="text" :maxlength="50" v-model="form.remark" :placeholder="$t('请输入备注')"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script>
  import UserApi from '@/api/user.js';
  export default {
    data() {
      return {
        /*是否显示*/
        loading: false,
        dialogVisible: false,
        form: {
          user_id: '',
          source: '0',
          value: null,
          remark: '',
        },
      };
    },
    props: ['open_edit', 'editData'],
    created() {
      this.dialogVisible = this.open_edit;
      this.form.user_id = this.editData.user_id;
    },
    computed: {
      labelText() {
        if (this.form.source == '0') {
          return $t('当前赠送账户余额');
        } else if (this.form.source == '1') {
          return $t('当前积分');
        } else if (this.form.source == '2') {
          return $t('当前主账户余额');
        }
      },
      maxValue() {
        if (this.form.source == '0') {
          return Number(this.editData.gift_balance);
        } else if (this.form.source == '1') {
          return Number(this.editData.points);
        } else if (this.form.source == '2') {
          return Number(this.editData.balance);
        }
      },
    },
    methods: {
      /*修改用户*/
      submit() {
        let self = this;
        self.$refs.formRef.validate((valid) => {
          if (valid) {
            self.loading = true;
            let params = self.form;
            UserApi.editUserDeduct(params, true)
              .then((data) => {
                self.loading = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: $t('操作成功'),
                    type: 'success',
                  });
                  self.dialogFormVisible(true);
                }
              })
              .catch((error) => {
                self.loading = false;
              });
          }
        });
      },



      changeSource() {
        this.form.value = null;
        this.$refs.formRef.clearValidate('value');
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

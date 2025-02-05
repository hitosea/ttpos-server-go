<template>
  <!--
      
      时间：2019-10-25
      描述：会员-用户列表-会员充值
  -->
  <el-dialog :title="$t('充值')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="recharge" ref="form2" label-position="top">
      <el-form-item for="no_click" :label="$t('当前余额')" :label-width="formLabelWidth">
        <el-input v-model="nowMoney" autocomplete="off" disabled="disabled"></el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('会员等级')" :label-width="formLabelWidth">
        <el-input v-model="form.grade.name" autocomplete="off" disabled="disabled"></el-input>
      </el-form-item>

      <el-form-item
        for="no_click"
        :label="$t('充值金额')"
        :label-width="formLabelWidth"
        prop="balance.money"
        :rules="[
          {
            required: true,
            message: $t('请输入充值金额'),
          },
          {
            validator: () => {
              return recharge.balance.money || recharge.balance.money === 0 ? true : false;
            },
            message: $t('请输入充值金额'),
          },
        ]"
      >
        <el-input-number :controls="false" :min="1" :max="100000000" :precision="2" :placeholder="$t('请输入充值金额')" v-model.number="recharge.balance.money"></el-input-number>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('赠送金额')" :label-width="formLabelWidth" prop="balance.gift_balance">
        <el-input-number
          :controls="false"
          :min="0"
          :max="100000000"
          :precision="2"
          :placeholder="$t('请输入赠送金额')"
          v-model.number="recharge.balance.gift_balance"
        ></el-input-number>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('赠送积分')" :label-width="formLabelWidth" prop="points.value">
        <el-input-number :controls="false" :min="0" :max="100000000" :precision="2" :placeholder="$t('请输入赠送积分')" v-model.number="recharge.points.value"></el-input-number>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('备注')" :label-width="formLabelWidth">
        <el-input type="textarea" v-model="recharge.balance.remark" :placeholder="$t('请输入备注')" :maxlength="50"></el-input>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="addUser(form.user_id)" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script>
  import UserApi from '@/api/user.js';
  export default {
    data() {
      return {
        loading: false,
        /*左边长度*/
        formLabelWidth: '120px',
        /*是否显示*/
        dialogVisible: false,
        /*默认选中*/
        activeName: 'first',
        recharge: {
          balance: {
            mode: 'inc',
            remark: '',
            money: null,
            gift_balance: null,
          },
          points: {
            mode: 'inc',
            remark: '',
            value: null,
          },
        },
        source: 2,
      };
    },
    props: ['open_add', 'form'],
    created() {
      this.dialogVisible = this.open_add;
    },
    computed: {
      nowMoney() {
        return Number(this.form.balance) + Number(this.form.gift_balance);
      },
    },
    methods: {
      /*选中*/
      handleClick(tab, event) {
        this.source = tab.index;
      },
      /*充值*/
      addUser(e) {
        let self = this;
        self.$refs[`form${this.source}`].validate((valid) => {
          if (valid) {
            let params = self.recharge;
            let user_id = e;
            let source = self.source;
            params.points.remark = params.balance.remark ?? '';
            self.loading = true;
            UserApi.userRecharge(
              {
                params: params,
                user_id: user_id,
                source: source,
              },
              true
            )
              .then((data) => {
                self.loading = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: $t('保存成功'),
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

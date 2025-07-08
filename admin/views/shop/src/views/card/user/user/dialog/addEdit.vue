Ï<template>
  <el-dialog :title="title" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" ref="form" :model="form" label-position="top">
      <!--店员修改-->
      <el-form-item for="no_click" :label="$t('昵称')" prop="nick_name" :rules="[{ required: true, message: $t('请输入昵称') }]">
        <el-input class="percent-w100" v-model="form.nick_name" :maxlength="50" :placeholder="$t('请输入昵称')"></el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('性别')" prop="gender" :rules="[{ required: true, message: $t('请选择性别') }]">
        <el-radio-group v-model="form.gender">
          <el-radio :label="2">{{ $t('保密') }}</el-radio>
          <el-radio :label="1">{{ $t('男') }}</el-radio>
          <el-radio :label="0">{{ $t('女') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('会员等级')" prop="grade_id" :rules="[{ required: true, message: $t('请选择等级') }]">
        <el-select class="percent-w100" v-model="form.grade_id" :placeholder="$t('请选择等级')">
          <el-option v-for="(item, index) in gradeSelectList" :key="index" :label="item.name" :value="item.grade_id"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('会员卡')">
        <el-select class="percent-w100" v-model="form.card_uuid" :placeholder="$t('请选择会员卡')" clearable>
          <el-option v-for="(item, index) in cardList" :key="index" :label="item.price > 0 ? item.name + ' ' + `(${item.price})` : item.name" :value="item.uuid"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item v-if="form.card_uuid" for="no_click" :label="$t('会员卡号')">
        <el-input class="percent-w100" v-model="form.card_number" @input="inputCardNumber" :maxlength="48" :placeholder="$t('请输入会员卡号')"></el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('推荐人')">
        <el-button size="small" type="primary" @click="selectReferrer" v-if="(form.referrer == null && editData) || !editData">{{ $t('选择') }}</el-button>
        <template v-if="selectMenber.length > 0">
          <el-tag v-for="tag in selectMenber" size="large" :key="tag.user_id" closable @close="handleClose(tag)">
            {{ `${tag.phone || '-'} (${tag.nickname})` }}
          </el-tag>
        </template>

        <template v-if="form.referrer">
          <span>{{ `${form.referrer.phone || '-'} (${form.referrer.nickname || '-'})` }}</span>
        </template>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('手机号')" prop="mobile" :rules="[{ required: true, message: $t('请输入手机号') }]">
        <el-input class="percent-w100" :maxlength="20" v-model="form.mobile" :placeholder="$t('请输入手机号')"></el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('密码')" prop="password">
        <el-input class="percent-w100" type="password" v-model="form.password" :placeholder="$t('请输入密码')"></el-input>
        <div class="tips">{{ $t('密码必须是4-16位的数字') }}</div>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('生日')">
        <el-date-picker class="date-picker" :clearable="true" v-model="form.birthday" type="date" value-format="YYYY-MM-DD"></el-date-picker>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button size="small" @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button size="small" type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
  <!--选择用户-->
  <GetUser :is_open="open_getuser" :exclude_user_id="form.user_id" @close="closeGetuserFunc" :is_single="true"></GetUser>
</template>
<script>
  import UserApi from '@/api/user.js';
  import { useUserStore } from '@/store';
  import GetUser from '@/components/user/GetUser.vue';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;
  export default {
    components: {
      GetUser,
    },
    data() {
      return {
        dialogVisible: false,
        form: {
          nick_name: '',
          gender: 2,
          mobile: '',
          grade_id: 1,
          password: '',
          birthday: '',
          card_uuid: '',
          card_number: '',
          referrer_uuid: '',
        },
        loading: false,
        gradeSelectList: [],
        app_id,
        cardList: [],
        open_getuser: false,
        selectMenber: [],
      };
    },
    props: ['open', 'editform', 'title', 'gradeList', 'editData'],
    created() {
      this.dialogVisible = this.open;
      if (this.editData) {
        this.form = JSON.parse(JSON.stringify(this.editData));
        this.form.nick_name = this.editData.nickName;
        this.form.card_uuid = this.editData.memberCard?.card?.card_id || '';
        this.form.card_number = this.editData.member_card_no;
        this.form.referrer_uuid = this.editData.referrer_uuid;
      }
      if (this.gradeList && this.gradeList.length > 0) {
        this.gradeList.map((item) => {
          this.gradeSelectList.push({
            grade_id: item.grade_id,
            name: item.name,
          });
        });
        if (!this.editData) {
          this.form.grade_id = this.gradeSelectList[0].grade_id;
        }
      }
      this.getCardList();
    },
    methods: {
      handleClose(tag) {
        this.selectMenber = this.selectMenber.filter((item) => item.user_id !== tag.user_id);
        this.form.referrer_uuid = '';
      },
      selectReferrer() {
        this.open_getuser = true;
      },
      /*关闭获取用户*/
      closeGetuserFunc(e) {
        if (e && e.type != 'error') {
          this.selectMenber = e.params;
          this.form.referrer_uuid = this.selectMenber[0].user_id;
        }
        this.open_getuser = false;
      },
      inputCardNumber(e) {
        //1~48位字符，允许输入字母和数字，不允许输入特殊字符
        this.$nextTick(() => {
          this.form.card_number = e.replace(/[^a-zA-Z0-9]/g, '');
        });
      },
      onSubmit() {
        let self = this;
        if (self.editData) {
          let params = {};
          params.user_id = self.form.user_id;
          params.nick_name = self.form.nick_name;
          params.gender = self.form.gender;
          params.grade_id = self.form.grade_id;
          params.mobile = self.form.mobile;
          params.password = self.form.password;
          params.birthday = self.form.birthday;
          params.card_uuid = self.form.card_uuid;
          params.card_number = self.form.card_number;
          params.referrer_uuid = self.form.referrer_uuid;
          self.$refs.form.validate((valid) => {
            if (valid) {
              self.loading = true;
              UserApi.edituser(params, true)
                .then((data) => {
                  self.loading = false;
                  this.$ElMessage({
                    message: $t('保存成功'),
                    type: 'success',
                  });
                  self.dialogFormVisible(1);
                })
                .catch((error) => {
                  self.loading = false;
                });
            }
          });
        } else {
          let params = self.form;
          self.$refs.form.validate((valid) => {
            if (valid) {
              self.loading = true;
              UserApi.adduser(params, true)
                .then((data) => {
                  self.loading = false;
                  this.$ElMessage({
                    message: $t('添加成功'),
                    type: 'success',
                  });
                  self.dialogFormVisible(1);
                })
                .catch((error) => {
                  self.loading = false;
                });
            }
          });
        }
      },

      getCardList() {
        UserApi.getCardList()
          .then((data) => {
            this.cardList = data.data.list;
          })
          .catch((error) => {
            this.$ElMessage({
              message: $t('获取失败'),
              type: 'error',
            });
          });
      },

      /*关闭弹窗*/
      dialogFormVisible(e) {
        this.$emit('closeDialog', e);
      },
    },
  };
</script>
<style scoped lang="scss">
  :deep(.date-picker) {
    width: 180px !important;
    max-width: 180px !important;
  }
</style>

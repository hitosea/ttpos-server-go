<template>
  <el-dialog
    :title="$t('发卡')"
    v-model="dialogVisible"
    @close="dialogFormVisible"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :modal-append-to-body="false"
    width="600px"
  >
    <el-form size="small" :model="form" label-position="top">
      <el-form-item>
        <div class="d-s-s">
          <div class="d-b-s">
            <div class="fb mr10"></div>
            <div class="text item">
              <div>{{ $t('卡名称: ') }}{{ form.card_name }}</div>
              <div>{{ $t('卡ID:') }} {{ form.card_id }}</div>
              <div
                >{{ $t('有效期:') }}
                <span v-if="form.expire > 0">{{ form.expire }}{{ $t('月') }}</span>
                <span v-else>{{ $t('永久有效') }}</span>
              </div>
              <div
                >{{ $t('折扣: ') }} <span v-if="form.is_discount > 0">{{ Number(form.discount) }}{{ $t('折') }}</span>
                <span v-else>{{ $t('无') }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-form-item>
      <el-form-item label="" :label-width="formLabelWidth">
        <div class="d-s-s d-c w-100">
          <el-button @click="openGetuser" icon="Plus">{{ $t('选择会员') }}</el-button>
          <el-table class="mt10" v-if="select_list.length > 0" size="small" max-height="300" :data="select_list" border style="width: 100%">
            <el-table-column prop="nickName" :label="$t('会员')"></el-table-column>
            <el-table-column prop="member_card_no" :label="$t('会员卡号')">
              <template #default="scope">
                <el-input
                  class="percent-w100"
                  v-model="scope.row.member_card_no"
                  @input="(e) => inputCardNumber(e, scope.$index)"
                  :maxlength="48"
                  :placeholder="$t('请输入会员卡号')"
                ></el-input>
              </template>
            </el-table-column>
            <el-table-column :label="$t('操作')" width="100">
              <template #default="scope">
                <el-button type="danger" link size="small" @click="deleteOne(scope.$index)">{{ $t('移除') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="editUser" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
    <!--选择用户-->
    <GetUser :is_open="open_getuser" :detailSelection="selectListID" @close="closeGetuserFunc"></GetUser>
  </el-dialog>
</template>

<script>
  import CardApi from '@/api/card.js';
  import GetUser from '@/components/user/GetUser.vue';
  export default {
    components: {
      GetUser,
    },
    data() {
      return {
        /*左边长度*/
        formLabelWidth: '120px',
        /*是否显示*/
        loading: false,
        dialogVisible: false,
        /*获取用户是否显示*/
        open_getuser: false,
        user_ids: '',
        /*选择的用户列表*/
        select_list: [],
      };
    },
    props: ['open_edit', 'form'],
    created() {
      this.dialogVisible = this.open_edit;
    },
    computed: {
      selectListID () {
        if(this.select_list.length>0) {
          return this.select_list.map(item => item.id);
        }
        return [];
      },
    },
    methods: {
      inputCardNumber(e, index) {
        //1~48位字符，允许输入字母和数字，不允许输入特殊字符
        this.$nextTick(() => {
          this.select_list[index].member_card_no = e.replace(/[^a-zA-Z0-9]/g, '');
        });
      },

      /*修改用户*/
      editUser() {
        let self = this;
        let params = {};
        params.card_id = self.form.card_id;
        params.user_ids = [];
        self.select_list.map((item) => {
          params.user_ids.push({
            uuid: item.uuid,
            card_number: item.member_card_no,
          });
        });
        self.loading = true;
        CardApi.putcard(params, true)
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
      },
      /*打开获取用户*/
      openGetuser() {
        this.open_getuser = true;
      },

      deleteOne(index) {
        this.select_list.splice(index, 1);
      },

      /*关闭获取用户*/
      closeGetuserFunc(e) {
        if (e && e.type != 'error') {
          this.select_list = [...e.params];
        }
        this.open_getuser = false;
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
  .d-c-s {
    display: flex;
    justify-content: center;
    align-items: flex-start;
  }

  .w-100 {
    width: 100%;
  }
</style>

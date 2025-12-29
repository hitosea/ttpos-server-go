<template>
  <el-dialog class="" @close="handleClose" v-model="dialogVisible" :close-on-click-modal="false" :close-on-press-escape="false" :title="title">
    <el-form size="small" ref="form" :model="form" label-position="top">
      <el-form-item v-if="!editItem" for="no_click" :label="$t('添加方式')" prop="add_method" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.add_method" @change="handleChangeType">
          <el-radio :label="1">{{ $t('直接添加') }}</el-radio>
          <el-radio :label="2">{{ $t('快捷添加') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item for="no_click" class="h-auto" v-if="form.add_method == 2" prop="add_pay_type" :rules="[{ required: true, message: $t('请选择') }]">
        <el-select v-model="form.add_pay_type" multiple @change="handleChange">
          <template v-for="(item, index) in payList" :key="index">
            <el-option :value="item.value" :label="item.name" :disabled="item.can_add == false">{{ item.name }}</el-option>
          </template>
        </el-select>
      </el-form-item>
      <template v-if="form.add_method == 1">
        <el-form-item
          for="no_click"
          :label="$t('名称')"
          prop="remark"
          :rules="[
            { required: true, message: $t('请输入名称') },
            { validator: uniqueNameValidator('pay_type', editItem?.id, 'SINGLE'), trigger: 'blur' },
          ]"
        >
          <el-input v-model="form.remark" :placeholder="$t('请输入名称')" :maxlength="100" :disabled="editItem.id && (editItem.source == '2' || editItem.source == '3')"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('支付方式')" prop="name" :rules="[{ required: true, message: $t('请输入支付方式') }]">
          <el-input
            v-model="form.name"
            :placeholder="$t('请输入支付方式')"
            :disabled="(editItem.id && (editItem.source == '2' || editItem.source == '3')) || editItem.id || erp_is_open == 1"
          ></el-input>
          <div class="tips">{{ $t('注：需为正确的支付方式，用于显示给用户查看') }}</div>
        </el-form-item>
        <el-form-item
          for="no_click"
          :label="$t('图片')"
          prop="img"
          :rules="[
            {
              validator: () => {
                return (!editItem.id && form.img && form.img.length > 0) ||
                  (editItem.id && editItem.source != '2' && form.img && form.img.length > 0) ||
                  (editItem.id && editItem.source == '2')
                  ? true
                  : false;
              },
              required: true,
              trigger: 'change',
              message: $t('请选择图片'),
            },
          ]"
        >
          <div class="draggable-list">
            <template v-if="!editItem.id || (editItem.id && editItem.source != '2')">
              <draggable class="wrapper" v-model="form.img" :disabled="editItem.source == '3'">
                <transition-group>
                  <div class="item" v-for="(item, index) in form.img" :key="item.file_path">
                    <img v-img-url="item.file_path" />
                    <a href="javascript:void(0);" v-if="!editItem.id || (editItem.id && editItem.source != '2' && editItem.source != '3')" class="delete-btn" @click.stop="deleteImg(index)"
                      ><el-icon> <Close /> </el-icon
                    ></a>
                  </div>
                </transition-group>
              </draggable>
              <div class="item img-select" v-if="form.img?.length == 0" @click="openProductUpload(1, -1)">
                <el-icon>
                  <Plus />
                </el-icon>
              </div>
            </template>
            <template v-if="editItem.id && editItem.source == '2'">
              <draggable class="wrapper" v-model="form.img" v-if="form.img.length > 0">
                <transition-group>
                  <div class="item" v-for="item in form.img" :key="item.file_path">
                    <img v-img-url="item.file_path" />
                  </div>
                </transition-group>
              </draggable>
              <img v-if="form.img?.length == 0" src="@/assets/img/ja_pay.png" style="width: 108px" />
            </template>
            <div class="tips">{{ $t('支持JPG、JPEG、PNG、WEBP格式，小于15MB，尺寸：48*48px') }}</div>
          </div>
        </el-form-item>

        <el-form-item v-if="!editItem.id || editItem.source == '1'" for="no_click" :label="$t('二维码')">
          <div class="draggable-list">
            <draggable class="wrapper" v-model="form.qrcode">
              <transition-group>
                <div class="item" v-for="(item, index) in form.qrcode" :key="item.file_path">
                  <img v-img-url="item.file_path" />
                  <a href="javascript:void(0);" class="delete-btn" @click.stop="deleteQrCode(index)"
                    ><el-icon>
                      <Close />
                    </el-icon>
                  </a>
                </div>
              </transition-group>
            </draggable>
            <div class="item img-select" v-if="form.qrcode?.length == 0" @click="openProductUpload(2, -1)">
              <el-icon>
                <Plus />
              </el-icon>
            </div>
            <div class="tips">{{ $t('支持JPG、JPEG、PNG、WEBP格式，小于15MB，尺寸：148*148px') }}</div>
          </div>
        </el-form-item>
        <el-form-item v-if="!editItem.id || editItem.source != '0'" for="no_click" :label="$t('手续费')" prop="fee" :rules="[{ required: true, message: $t('默认0') }]">
          <el-input v-model="form.fee" :min="0" :max="100" type="number">
            <template #append>%</template>
          </el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('排序')" prop="sort" :rules="[{ required: true, message: $t('接近0，排序等级越高') }]">
          <el-input-number :controls="false" :min="0" :max="999" :precision="0" :placeholder="$t('接近0，排序等級越高')" v-model.number="form.sort"></el-input-number>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('结账显示')">
          <el-checkbox-group v-model="form.is_show_checkout">
            <el-checkbox label="10" size="small">
              {{ $t('收银机') }}
            </el-checkbox>
            <el-checkbox label="20" size="small">
              {{ $t('点餐助手') }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item v-if="!editItem.id || editItem.value != '10'" for="no_click" :label="$t('会员充值')">
          <el-checkbox-group v-model="form.is_show_recharge">
            <el-checkbox label="10" size="small">
              {{ $t('收银机') }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('状态')">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0"></el-switch>
        </el-form-item>
      </template>

      <template v-if="form.add_method == 2">
        <div class="multiple-box" v-for="(item, index) in form.params" :key="item.id">
          <el-form-item
            for="no_click"
            :label="$t('名称')"
            :prop="`params[${index}].remark`"
            :rules="[
              {
                validator: uniqueNameValidator('pay_type', undefined, 'SINGLE'),
                trigger: 'blur',
              },
            ]"
          >
            <template #label> <span class="must">*</span>{{ $t('名称') }} </template>
            <el-input v-model="item.remark" :placeholder="$t('请输入名称')" :maxlength="100" :disabled="item.source == '3'"></el-input>
          </el-form-item>
          <el-form-item
            for="no_click"
            :label="$t('支付方式')"
            :prop="`params[${index}].name`"
            :rules="[
              {
                validator: () => {
                  return !!item.name;
                },
                required: true,
                message: $t('请输入支付方式'),
              },
            ]"
          >
            <el-input v-model="item.name" :disabled="editItem.id || item.source == '3'" :placeholder="$t('请输入支付方式')"></el-input>
          </el-form-item>
          <el-form-item
            for="no_click"
            :label="$t('图片')"
            :prop="`params[${index}].img[0].file_path`"
            :rules="[
              {
                validator: () => {
                  return item.img && item.img.length > 0 ? true : false;
                },
                required: true,
                trigger: 'change',
                message: $t('请选择图片'),
              },
            ]"
          >
            <div class="draggable-list">
              <draggable class="wrapper" v-model="item.img" :disabled="item.source == '3'">
                <transition-group>
                  <div class="item" v-for="(items, indexs) in item.img" :key="items.file_path">
                    <img v-img-url="items.file_path ? items.file_path : ''" />
                    <a href="javascript:void(0);" v-if="item.source != '3'" class="delete-btn" @click.stop="deleteQuicklyImg(index, indexs)"
                      ><el-icon> <Close /> </el-icon
                    ></a>
                  </div>
                </transition-group>
              </draggable>
              <div class="item img-select" v-if="item.img?.length == 0" @click="openProductUpload(1, index)">
                <el-icon>
                  <Plus />
                </el-icon>
              </div>
              <div class="tips">{{ $t('支持JPG、JPEG、PNG、WEBP格式，小于15MB，尺寸：48*48px') }}</div>
            </div>
          </el-form-item>
          <el-form-item for="no_click" :label="$t('二维码')">
            <div class="draggable-list">
              <draggable class="wrapper" v-model="item.qrcode">
                <transition-group>
                  <div class="item" v-for="(items, indexs) in item.qrcode" :key="items.file_path">
                    <img v-img-url="items.file_path" />
                    <a href="javascript:void(0);" class="delete-btn" @click.stop="deleteQuicklyQrCode(index, indexs)"
                      ><el-icon>
                        <Close />
                      </el-icon>
                    </a>
                  </div>
                </transition-group>
              </draggable>
              <div class="item img-select" v-if="item.qrcode?.length == 0" @click="openProductUpload(2, index)">
                <el-icon>
                  <Plus />
                </el-icon>
              </div>
              <div class="tips">{{ $t('支持JPG、JPEG、PNG、WEBP格式，小于15MB，尺寸：148*148px') }}</div>
            </div>
          </el-form-item>
          <el-form-item
            v-if="!editItem.id || editItem.source != '0'"
            for="no_click"
            :label="$t('手续费')"
            :prop="`params[${index}].fee`"
            :rules="[
              {
                validator: () => {
                  return typeof item.fee == 'number' && item.fee >= 0 ? true : false;
                },
                required: true,
                message: $t('默认0'),
              },
            ]"
          >
            <el-input v-model.number="item.fee" type="number">
              <template #append>%</template>
            </el-input>
          </el-form-item>
          <el-form-item
            for="no_click"
            :label="$t('排序')"
            :prop="`params[${index}].sort`"
            :rules="[
              {
                validator: () => {
                  return item.sort != null ? true : false;
                },
                trigger: 'change',
                required: true,
                message: $t('接近0，排序等级越高'),
              },
            ]"
          >
            <el-input-number :controls="false" :min="0" :max="999" :precision="0" :placeholder="$t('接近0，排序等級越高')" v-model.number="item.sort"></el-input-number>
          </el-form-item>
          <el-form-item for="no_click" :label="$t('结账显示')">
            <el-checkbox-group v-model="item.is_show_checkout">
              <el-checkbox label="10" size="small">
                {{ $t('收银机') }}
              </el-checkbox>
              <el-checkbox label="20" size="small">
                {{ $t('点餐助手') }}
              </el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item for="no_click" :label="$t('会员充值')">
            <el-checkbox-group v-model="item.is_show_recharge">
              <el-checkbox label="10" size="small">
                {{ $t('收银机') }}
              </el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item for="no_click" :label="$t('状态')">
            <el-switch v-model="item.status" :active-value="1" :inactive-value="0"></el-switch>
          </el-form-item>
        </div>
      </template>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </span>
    </template>
    <Upload v-if="isProductUpload" :config="{ total: 1 }" :isupload="isProductUpload" @returnImgs="returnProductImgsFunc"> </Upload>
  </el-dialog>
</template>
<script>
  import Upload from '@/components/file/Upload.vue';
  import SettingApi from '@/api/setting.js';
  import { uniqueNameValidator } from '@/utils/form.js';

  export default {
    components: { Upload },
    props: ['title', 'open', 'editItem', 'list', 'erp_is_open', 'erp_site_code'],
    created() {
      this.dialogVisible = this.open;
      this.getDefaultPay();
      if (this.editItem) {
        this.form.id = this.editItem.id;
        this.form.name = this.editItem.name;
        this.form.remark = this.editItem.remark;
        this.form.fee = this.editItem.fee;
        this.form.sort = this.editItem.sort;
        this.form.status = this.editItem.status;
        this.form.is_show_checkout = this.editItem.is_show_checkout;
        this.form.is_show_recharge = this.editItem.is_show_recharge;
        if (this.editItem.img) {
          this.form.img.push({ file_id: this.editItem.logo_file_uuid, file_path: this.editItem.img });
        }
        if (this.editItem.qrcode) {
          this.form.qrcode.push({ file_id: this.editItem.qrcode_file_uuid, file_path: this.editItem.qrcode });
        }
      }
    },
    data() {
      return {
        isProductUpload: false,
        loading: false,
        payList: [],

        form: {
          add_method: 1,
          name: '',
          img: [],
          qrcode: [],
          remark: '',
          fee: 0,
          sort: 0,
          status: 1,
          is_show_checkout: [],
          is_show_recharge: [],
        },
        formIncludes: [],
        imgType: 1,
        imgIndex: -1,
        quiver: '',
      };
    },
    methods: {
      /*打开上传图片*/
      openProductUpload: function (e, index) {
        if (this.editItem.id && (this.editItem.source == '2' || this.editItem.source == '3')) return;
        this.isProductUpload = true;
        this.imgType = e;
        this.imgIndex = index;
      },

      /*原始支付列表*/
      getDefaultPay() {
        self.loading = true;
        SettingApi.defaultPay({}, true)
          .then((data) => {
            self.loading = false;
            this.payList = data.data.list;
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      /*上传商品图片*/
      returnProductImgsFunc(e) {
        if (this.imgIndex == -1) {
          if (e != null && this.imgType == 1) {
            let imgs = this.form.img.concat(e);
            this.form['img'] = imgs;
            this.$refs.form.validateField('img');
          }
          if (e != null && this.imgType == 2) {
            let imgs = this.form.qrcode.concat(e);
            this.form['qrcode'] = imgs;
          }
        } else {
          if (e != null && this.imgType == 1) {
            let imgs = this.form.params[this.imgIndex].img.concat(e);
            this.form.params[this.imgIndex]['img'] = imgs;
            this.$refs.form.validateField(`params[${this.imgIndex}].img[0].file_path`);
          }
          if (e != null && this.imgType == 2) {
            let imgs = this.form.params[this.imgIndex].qrcode.concat(e);
            this.form.params[this.imgIndex]['qrcode'] = imgs;
          }
        }
        this.isProductUpload = false;
      },

      /*删除商品图片*/
      deleteImg(index) {
        if (this.editItem.source == '3') return;
        this.form.img.splice(index, 1);
        this.$refs.form.validateField('img');
      },

      /*快捷添加删除商品图片*/
      deleteQuicklyImg(index, indexs) {
        if (this.form.params[index].source == '3') return;
        this.form.params[index].img.splice(indexs, 1);
        this.$refs.form.validateField(`params[${index}].img[${indexs}].file_path`);
      },

      /*删除QCODE图片*/
      deleteQrCode(index) {
        this.form.qrcode.splice(index, 1);
      },
      /*删除QCODE图片*/
      deleteQuicklyQrCode(index, indexs) {
        this.form.params[index].qrcode.splice(indexs, 1);
      },

      //提交表单
      onSubmit() {
        let self = this;
        if (this.form.add_method == 1) {
          self.$refs.form.validate((valid) => {
            if (valid) {
              this.addPay();
            }
          });
        } else {
          self.$refs.form.validate((valid) => {});
          // 设置一个定时器，在200毫秒后，获取所有带有el-form-item__error类的元素，并将第一个元素滚动到视图中
          self.form.params.map((item, index) => {
            self.$refs.form.validateField(`params[${index}].remark`);
          });
          setTimeout(() => {
            const errorItems = document.querySelectorAll('.el-form-item__error');
            if (errorItems.length > 0) {
              const firstErrorItem = errorItems[0];
              firstErrorItem.scrollIntoView({
                behavior: 'smooth',
                block: 'center',
              });
            } else {
              this.addPay();
            }
          }, 200);
        }
      },

      addPay() {
        let self = this;
        let data = {};
        data.add_method = JSON.parse(JSON.stringify(self.form.add_method));
        if (data.add_method == 1) {
          data = JSON.parse(JSON.stringify(self.form));
          data.img = JSON.parse(JSON.stringify(self.form?.img[0] || '{}'));
          data.qrcode = JSON.parse(JSON.stringify(self.form?.qrcode[0] || '{}'));
          data.is_show_checkout.length == 0 ? (data.is_show_checkout = '') : '';
          data.is_show_recharge.length == 0 ? (data.is_show_recharge = '') : '';
        } else {
          data.params = JSON.parse(JSON.stringify(self.form.params));
          data.params.map((item) => {
            delete item.id;
            item.img = JSON.parse(JSON.stringify(item?.img[0] || '{}'));
            item.qrcode = JSON.parse(JSON.stringify(item?.qrcode[0] || '{}'));
            item.is_show_checkout.length == 0 ? (item.is_show_checkout = '') : '';
            item.is_show_recharge.length == 0 ? (item.is_show_recharge = '') : '';
          });
        }
        self.loading = true;
        if (this.editItem) {
          SettingApi.setPaytype(data, true)
            .then((data) => {
              self.loading = false;
              this.$ElMessage({
                message: $t('操作成功'),
                type: 'success',
              });
              this.$emit('close', 1);
            })
            .catch((error) => {
              self.loading = false;
            });
        } else {
          SettingApi.addPayType(data, true)
            .then((data) => {
              self.loading = false;
              this.$ElMessage({
                message: $t('添加成功'),
                type: 'success',
              });
              this.$emit('close', 1);
            })
            .catch((error) => {
              self.loading = false;
            });
        }
      },

      handleClose() {
        this.$emit('close');
      },

      uniqueNameValidator: uniqueNameValidator,

      //切换添加方式
      handleChangeType() {
        this.formIncludes = [];
        if (this.form.add_method == 1) {
          this.form = {
            add_method: 1,
            name: '',
            img: [],
            qrcode: [],
            remark: '',
            fee: 0,
            sort: 0,
            status: 1,
            is_show_checkout: [],
            is_show_recharge: [],
          };
        } else {
          this.form = {
            add_method: 2,
            add_pay_type: [],
            params: [],
          };
        }
      },

      //多选的时候
      handleChange(e) {
        const result = this.compareArrays(e, this.formIncludes);
        //新增
        if (result.missingInB[0]) {
          this.payList.map((item) => {
            if (item.value == result.missingInB[0]) {
              this.form.params.push({
                id: result.missingInB[0],
                name: item.name,
                img: item.img ? [{ file_path: item.img || '' }] : [],
                qrcode: [],
                remark: item.source == '3' ? item.name + '（Kbank）' : item.name,
                fee: 0,
                sort: 0,
                status: 1,
                is_show_checkout: [],
                is_show_recharge: [],
                source: item.source,
              });
            }
          });
        }
        //去掉
        if (result.extraInB[0]) {
          this.form.params = this.form.params.filter((item) => item.id != result.extraInB[0]);
        }
        this.formIncludes = JSON.parse(JSON.stringify(e));
      },
      //计算两数组确实或者少了
      compareArrays(arr1, arr2) {
        const missingInB = arr1.filter((item) => !arr2.includes(item));
        const extraInB = arr2.filter((item) => !arr1.includes(item));

        return {
          missingInB,
          extraInB,
        };
      },
    },
  };
</script>
<style scoped lang="scss">
  .draggable-list {
    .item {
      margin-top: 0;
    }
  }

  .multiple-box {
    border-bottom: dashed 1px #ccc;
    margin-bottom: 16px;
  }

  .multiple-box:nth-last-child(1) {
    border-bottom: none;
    margin-bottom: 0;
  }
  .must {
    color: var(--el-color-danger);
    margin-right: 4px;
  }
  :deep(.el-select--small .el-select__wrapper) {
    min-height: 32px !important;
    height: auto;
    padding: 4px 8px !important;
  }
</style>

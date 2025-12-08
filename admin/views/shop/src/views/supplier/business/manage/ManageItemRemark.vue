<template>
  <el-dialog class="" @close="handleClose" v-model="dialogVisible" :close-on-click-modal="false" :close-on-press-escape="false" :title="$t('单品备注')">
    <el-form size="small" ref="formRef" :model="form" label-position="top" :rules="rules">
      <template v-for="(item, itemIndex) in form.remark" :key="item.id">
        <el-card class="item mb-2" shadow="never" :style="{ display: item.action === 'delete' ? 'none' : 'block' }">
          <div class="item-input">
            <template v-for="(language, languageIndex) in item.data" :key="languageIndex">
              <el-form-item
                for="no_click"
                :label="language.value"
                :prop="`remark.${itemIndex}.data.${languageIndex}.remark`"
                :rules="[{ required: true, message: $t('请输入备注') }]"
              >
                <mInput
                  v-model:valueData="language.remark"
                  :value="language.remark"
                  :placeholder="$t('请输入备注')"
                  :maxlength="50"
                  :langKey="language.name"
                  width="w-full"
                  @translate="(response) => translate(item)(response)"
                />
              </el-form-item>
            </template>
          </div>
          <template #footer>
            <div class="item-remove">
              <el-button @click="handleRemove(item.id)" type="danger" size="small">{{ $t('删除') }}</el-button>
            </div>
          </template>
        </el-card>
      </template>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <div>
          <el-button @click="handleAdd" type="primary">{{ $t('添加') }}</el-button>
        </div>
        <div>
          <el-button @click="handleClose">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>
<script>
  import SettingApi from '@/api/setting.js';
  import mInput from '@/components/m-input/index.vue';
  import { languageStore } from '@/store/model/language.js';
  const languageList = languageStore().getLanguageList().languageList.value;

  export default {
    components: {
      mInput,
    },
    props: ['open'],
    created() {
      this.dialogVisible = this.open;
      this.getData();
    },
    data() {
      return {
        dialogVisible: false,
        loading: false,
        form: {
          remark: [],
        },
        rules: {},
      };
    },
    methods: {
      // 获取单品备注列表
      getData() {
        let self = this;
        self.loading = true;
        SettingApi.getItemRemark({}, true)
          .then((res) => {
            const data = res.data;
            if (!Array.isArray(data)) return;

            self.form.remark = data.map((item) => {
              const getItemRemark = (value) => {
                try {
                  const original = JSON.parse(value);
                  return languageList.map((language) => ({
                    ...language,
                    remark: original[language.key] || '???',
                  }));
                } catch (error) {
                  return languageList.map((language) => ({
                    ...language,
                    remark: '???',
                  }));
                }
              };
              return {
                id: item.id,
                action: 'edit',
                data: getItemRemark(item.order_item_remark),
              };
            });
          })
          .catch((err) => {})
          .finally(() => {
            self.loading = false;
          });
      },

      // 提交保存
      onSubmit() {
        let self = this;
        self.$refs.formRef.validate((valid) => {
          if (!valid) return;
          self.loading = true;
          const remark = self.form.remark.map((item) => {
            const remarkData = Array.isArray(item.data) ? item.data.reduce((remarkData, language) => ({ ...remarkData, [language.key]: language.remark }), {}) : {};
            return {
              id: item.id,
              remark: JSON.stringify(remarkData),
              action: item.action,
            };
          });
          SettingApi.setItemRemark({ order_item_remark: remark }, true)
            .then((res) => {
              this.$ElMessage({
                type: 'success',
                message: $t('更新成功'),
              });
              self.handleClose();
            })
            .catch(() => {})
            .finally(() => {
              self.loading = false;
            });
        });
      },

      // 关闭弹窗
      handleClose() {
        this.$emit('close', true);
      },

      // 添加备注项
      handleAdd() {
        const item = {
          id: 0,
          action: 'add',
          data: languageList.map((item) => ({
            ...item,
            remark: '',
          })),
        };
        this.form.remark.push(item);
      },

      // 删除备注项
      handleRemove(id) {
        const item = this.form.remark.find((item) => item.id === id);
        if (!item) return;
        if (item.action === 'add') {
          this.form.remark = this.form.remark.filter((item) => item.id !== id);
        } else {
          item.action = 'delete';
        }
      },

      // 翻译功能
      translate(item) {
        return (response) => {
          if (!Array.isArray(item?.data) || !Array.isArray(response)) return;
          const res = response[0];
          if (!res) return;

          for (const language of item.data) {
            language.remark = res[language.name === 'zhtw' ? 'zh-TW' : language.name] || '';
          }
        };
      },
    },
  };
</script>

<style lang="scss" scoped>
  .dialog-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .item {
    &.mb-2 {
      margin-bottom: 20px;
    }

    .item-input {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 10px;
    }
  }
</style>

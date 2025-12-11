<template>
  <el-dialog class="" @close="handleClose" v-model="dialogVisible" :close-on-click-modal="false" :close-on-press-escape="false" :title="$t('免单/赠菜原因')">
    <el-form size="small" ref="form" :model="form" label-position="top" :rules="rules">
      <template v-for="(item, itemIndex) in form.free_tag" :key="item.id">
        <el-card class="item mb-2" shadow="never" :style="{ display: item.action === 'delete' ? 'none' : 'block' }">
          <div class="item-input">
            <template v-for="(language, languageIndex) in item.data" :key="languageIndex">
              <el-form-item
                for="no_click"
                :label="language.value"
                :prop="`free_tag.${itemIndex}.data.${languageIndex}.reason`"
                :rules="[{ required: true, message: $t('请输入原因') }]"
              >
                <mInput
                  v-model:valueData="language.reason"
                  :value="language.reason"
                  :placeholder="$t('请输入原因')"
                  :maxlength="100"
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
        loading: false,
        form: {
          free_tag: [],
        },
        rules: {},
      };
    },
    methods: {
      getData() {
        let self = this;
        self.loading = true;
        SettingApi.getFreeTag({}, true)
          .then((res) => {
            const data = res.data;
            if (!Array.isArray(data)) return;

            self.form.free_tag = data.map((item) => {
              const getFreeTag = (value) => {
                try {
                  const original = JSON.parse(value);
                  return languageList.map((language) => ({
                    ...language,
                    reason: original[language.key] || '???',
                  }));
                } catch (error) {
                  return languageList.map((language) => ({
                    ...language,
                    reason: '???',
                  }));
                }
              };
              return {
                id: item.id,
                action: 'edit',
                data: getFreeTag(item.free_tag),
              };
            });
          })
          .catch((err) => {})
          .finally(() => {
            self.loading = false;
          });
      },
      onSubmit() {
        let self = this;
        self.$refs.form.validate((valid) => {
          if (!valid) return;
          self.loading = true;
          const free_tag = self.form.free_tag.map((item) => {
            const tag = Array.isArray(item.data) ? item.data.reduce((tag, language) => ({ ...tag, [language.key]: language.reason }), {}) : {};
            return {
              id: item.id,
              free_tag: JSON.stringify(tag),
              action: item.action,
            };
          });
          SettingApi.setFreeTag({ free_tag }, true)
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

      handleClose() {
        this.$emit('close', true);
      },

      handleAdd() {
        const item = {
          id: 0,
          action: 'add',
          data: languageList.map((item) => ({
            ...item,
            reason: '',
          })),
        };
        this.form.free_tag.push(item);
      },
      handleRemove(id) {
        const item = this.form.free_tag.find((item) => item.id === id);
        if (!item) return;
        if (item.action === 'add') {
          this.form.free_tag = this.form.free_tag.filter((item) => item.id !== id);
        } else {
          item.action = 'delete';
        }
      },

      translate(item) {
        return (response) => {
          if (!Array.isArray(item?.data) || !Array.isArray(response)) return;
          const res = response[0];
          if (!res) return;

          for (const language of item.data) {
            language.reason = res[language.name === 'zhtw' ? 'zh-TW' : language.name] || '';
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

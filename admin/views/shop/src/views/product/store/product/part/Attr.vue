<template>
  <!--描述：商品管理-商品编辑-商品属性-->
  <div>
    <!--商品属性-->
    <div class="common-form mt50">{{ $t('商品属性') }}</div>
    <!--商品属性-->
    <div>
      <div class="mt16">
        <!-- <el-form-item for="no_click"  label="属性明细："> -->
        <el-form-item for="no_click" :label="$t('商品属性：')">
          <el-button type="primary" @click="addAttr">{{ $t('添加属性') }}+</el-button>
        </el-form-item>

        <!--商品属性-->
        <div class="tableb-body">
          <div v-for="(item, index) in form.model.product_attr" :key="index">
            <div class="table-checkbox">
              <div>
                {{ index + 1 }}:{{
                  typeof item.attribute_name === 'string' ? JSON.parse(item.attribute_name || '{}')[languageKey] : item.attribute_name ? item.attribute_name[languageKey] : '-'
                }}
              </div>
              <div> <el-checkbox v-model="item.attribute_required" size="large" :true-label="1" :false-label="0" :label="$t('必选')" /></div>
              <div class="table-c-item">
                <el-checkbox
                  v-model="item.attribute_open_max_select"
                  size="large"
                  :true-label="1"
                  :false-label="0"
                  :label="$t('最多可选')"
                  @change="checkDefaultSelect(index, item.attribute_value)"
                />
                <el-input-number
                  v-if="item.attribute_open_max_select == '1'"
                  @input="checkDefaultSelect(index, item.attribute_value)"
                  @blur="onBlur"
                  :controls="false"
                  :min="1"
                  :max="item.attribute_value.length"
                  :precision="0"
                  v-model="item.attribute_max_select"
                  class="max-w460"
                ></el-input-number>
              </div>
              <el-icon class="delete-icon" @click="handleDelete(index)">
                <Delete />
              </el-icon>
            </div>
            <el-table :data="item.attribute_value" border style="width: 100%" class="table-no-padding">
              <el-table-column width="45" :label="$t('序号')" header-align="center" align="center">
                <template #default="scope">{{ scope.$index + 1 }} </template>
              </el-table-column>
              <el-table-column prop="attribute_name_text" :label="$t('属性值')" minWidth="140">
                <template #default="scope">
                  {{ JSON.parse(scope.row)[languageKey] || '-' }}
                </template>
              </el-table-column>
              <el-table-column :label="$t('默认勾选')" width="400">
                <template #default="scope">
                  <el-form-item
                    for="no_click"
                    class="mt16"
                    :prop="`model.product_attr[${index}].default_select[${scope.$index}]`"
                    :rules="[
                      {
                        validator: () => {
                          return defaultSelect(item.attribute_open_max_select, item.attribute_max_select, item.default_select) ? true : false;
                        },
                        message: $t('不能超过最多可选数量') + ' ' + item.attribute_max_select,
                      },
                    ]"
                  >
                    <el-checkbox
                      @change="checkDefaultSelect(index, item.attribute_value)"
                      v-model="item.default_select[scope.$index]"
                      size="large"
                      :true-label="1"
                      :false-label="0"
                    />
                  </el-form-item>
                </template>
              </el-table-column>
              <el-table-column prop="name" :label="$t('操作')" width="120">
                <template #default="scope">
                  <el-button @click="deleteClick(index, scope.$index, item.attribute_value)" type="primary" link size="small"> {{ $t('删除') }}</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>
    </div>

    <AttributeSelector v-if="open_selector" :open="open_selector" :selectedAttributeIds="select_arr" @close="handleCloseSelector" />
  </div>
</template>

<script>
  import { languageStore } from '@/store/model/language.js';
  import AttributeSelector from '@/components/product/AttributeSelector.vue';
  const languageKey = languageStore().getLanguageKey().language.value;

  export default {
    name: 'ProductStorePartAttr',
    components: { AttributeSelector },
    data() {
      return {
        formData: {},
        languageKey: languageKey,
        open_add_attr: false,
        open_selector: false,
        select_arr: [],
        parent_ids: [],
      };
    },
    inject: ['form'],
    watch: {
      'form.model.product_attr': {
        handler(val) {
          this.parent_ids = [];
          val.map((item) => {
            this.parent_ids.push(item.parent_id);
          });
        },
        immediate: true,
        deep: true,
      },
    },
    methods: {
      addAttr() {
        this.select_arr = [];
        this.form.model.product_attr.map((item) => {
          item.attribute_ids.map((items) => {
            this.select_arr.push(items);
          });
        });
        // this.open_add_attr = true;
        this.open_selector = true;
      },
      handleDelete(index) {
        // 删除属性
        this.form.model.product_attr.splice(index, 1);
      },

      deleteClick(index, indexs, attribute_value) {
        // 删除属性值
        this.form.model.product_attr[index].attribute_value.splice(indexs, 1);
        this.form.model.product_attr[index].default_select.splice(indexs, 1);
        this.form.model.product_attr[index].attribute_ids.splice(indexs, 1);
        if (this.form.model.product_attr[index].attribute_ids.length == 0) {
          this.form.model.product_attr.splice(index, 1);
        }
        this.$nextTick(() => {
          attribute_value.map((item, i) => {
            this.$emit('validateField', `model.product_attr[${index}].default_select[${i}]`);
          });
        });
      },

      defaultSelect(attribute_open_max_select, attribute_max_select, default_select) {
        let count = 0;
        default_select.map((item) => {
          if (item === 1) {
            count++;
          }
        });
        if (count > attribute_max_select && attribute_open_max_select == '1' && attribute_max_select != null) {
          return false;
        } else {
          return true;
        }
      },

      checkDefaultSelect(index, attribute_value) {
        this.$nextTick(() => {
          attribute_value.map((item, indexs) => {
            this.$emit('validateField', `model.product_attr[${index}].default_select[${indexs}]`);
          });
        });
      },

      onBlur() {
        this.form.model.product_attr.map((item, index) => {
          if (item.attribute_max_select == null) {
            item.attribute_max_select = 1;
            this.checkDefaultSelect(index, item.attribute_value);
          }
        });
      },

      handleCloseSelector(val) {
        this.open_selector = false;

        if (val == undefined) return;

        const _newParentIds = val.map((item) => item.parent_id).filter((item, index, self) => self.indexOf(item) === index);
        const _originalParentIds = this.parent_ids;
        const _originalData = this.form.model.product_attr;
        const needToDelete = _originalParentIds.filter((item) => !_newParentIds.includes(item));
        needToDelete.map((item) => {
          _originalData.splice(
            _originalData.findIndex((originalItem) => originalItem.parent_id === item),
            1
          );
        });

        const newData = [];
        for (const newItem of val) {
          const hasGroup = newData.find((item) => item.parent_id === newItem.parent_id);
          if (typeof hasGroup !== 'undefined') {
            hasGroup.attribute_value.push(newItem.attribute_name);
            hasGroup.default_select.push(0);
            hasGroup.attribute_ids.push(newItem.attribute_id);
            newData.splice(
              newData.findIndex((item) => item.parent_id === newItem.parent_id),
              1,
              hasGroup
            );
          } else {
            newData.push({
              parent_id: newItem.parent_id,
              parent_name: newItem.parent_attribute_name,
              attribute_name: newItem.parent_attribute_name,
              attribute_value: [newItem.attribute_name],
              default_select: [0],
              attribute_ids: [newItem.attribute_id],
              attribute_max_select: 0,
              attribute_open_max_select: 0,
              attribute_required: 0,
            });
          }
        }

        // compare newData and _originalData
        for (const newItem of newData) {
          const originalGroupItem = _originalData.find((item) => item.parent_id === newItem.parent_id);
          if (typeof originalGroupItem === 'undefined') continue;

          newItem.attribute_required = Number(originalGroupItem.attribute_required || 0);
          newItem.attribute_open_max_select = Number(originalGroupItem.attribute_open_max_select || 0);
          newItem.attribute_max_select = Number(originalGroupItem.attribute_max_select || 0);

          if (newItem.attribute_required === 1 && newItem.attribute_value.length === 0) {
            newItem.attribute_required = 0;
          }

          if (newItem.attribute_open_max_select === 1) {
            if (newItem.attribute_max_select > newItem.attribute_value.length) {
              newItem.attribute_max_select = newItem.attribute_value.length;
            }
          } else {
            newItem.attribute_max_select = 0;
          }

          if (originalGroupItem.default_select.some((item) => item !== 0)) {
            const _ids = originalGroupItem.attribute_ids.filter((item, index) => originalGroupItem.default_select[index] === 1);

            for (const id of _ids) {
              const _index = newItem.attribute_ids.findIndex((item) => item === id);
              if (_index !== -1) {
                newItem.default_select[_index] = 1;
              }
            }
          }

          newData.splice(
            newData.findIndex((item) => item.parent_id === newItem.parent_id),
            1,
            newItem
          );
        }

        this.form.model.product_attr = newData;
      },
    },
  };
</script>

<style scoped lang="scss">
  .product-attr {
    width: 100%;
    box-shadow: 0 0 0 1px var(--el-input-border-color, var(--el-border-color)) inset;
    padding: 16px 16px 0 16px;
    border-radius: var(--el-input-border-radius, var(--el-border-radius-base));
    margin-bottom: 12px;
  }

  .add-button {
    cursor: pointer;
    font-size: 24px;
    margin-right: 16px;
    margin-top: 4px;
  }

  :deep(.inline-input) {
    max-width: 460px;
    width: 100%;
  }

  .delete-icon {
    cursor: pointer;
    font-size: 24px;
    margin-left: auto;
  }

  .product-tips {
    font-size: 12px;
    color: var(--el-color-tips);
  }

  .product-box {
    display: flex;
  }

  :deep(.product-attr-item) {
    .el-form-item__content {
      align-items: flex-start !important;
    }
  }

  .tableb-body {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  .table-checkbox {
    display: flex;
    align-items: center;
    gap: 16px;
    background-color: var(--el-table-header-bg-color);
  }
  .table-c-item {
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .table-no-padding {
    :deep(td.el-table__cell) {
      padding: 0 !important;
    }
  }
</style>

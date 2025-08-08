<template>
  <div>
    <!--商品加料-->
    <div class="common-form mt50">{{ $t('商品加料') }}</div>

    <div>
      <div class="mt16">
        <el-form-item for="no_click" :label="$t('商品加料：')">
          <el-button type="primary" :disabled="form.model.product_feed.length >= 10" @click="addIngredients">{{ $t('添加加料') }}+</el-button>
        </el-form-item>
        <!--加料-->
        <template v-if="form.model.product_feed.length > 0">
          <div class="table-checkbox">
            <div> <el-checkbox v-model="form.model.feed_required" size="large" :true-value="1" :false-value="0" :label="$t('必选')" /></div>
            <div class="table-c-item">
              <el-checkbox v-model="form.model.feed_open_max_select" size="large" :true-value="1" :false-value="0" :label="$t('最多可选')" @change="checkDefaultSelect" />
              <el-input-number
                v-if="form.model.feed_open_max_select == '1'"
                @input="checkDefaultSelect()"
                @blur="onBlur"
                :controls="false"
                :min="1"
                :max="form.model.product_feed.length"
                v-model="form.model.feed_max_select"
                class="max-w460"
              ></el-input-number>
            </div>
            <el-icon class="delete-icon" @click="handleDelete()">
              <Delete />
            </el-icon>
          </div>
          <el-table :data="form.model.product_feed" border style="width: 100%">
            <el-table-column :label="$t('序号')" width="80">
              <template #default="scope">{{ scope.$index + 1 }} </template>
            </el-table-column>
            <el-table-column prop="feed_name" :label="$t('加料')" minWidth="300">
              <template #default="scope">
                {{ JSON.parse(scope.row.feed_name)[languageKey] || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="price" :label="$t('价格')" width="180">
              <template #default="scope">
                <el-form-item
                  for="no_click"
                  label=""
                  class="mt16"
                  :prop="`model.product_feed[${scope.$index}].price`"
                  :rules="[
                    {
                      validator: () => {
                        return scope.row.price != null ? true : false;
                      },
                      message: $t('请输入价格'),
                    },
                  ]"
                >
                  <numInput :controls="false" :min="0" :max="1000000" :precision="2" :placeholder="$t('请输入价格')" v-model="scope.row.price"></numInput>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="stock_num" :label="$t('库存')" width="180">
              <template #default="scope">
                <el-form-item
                  for="no_click"
                  label=""
                  class="mt16"
                  :prop="`model.product_feed[${scope.$index}].stock_num`"
                  :rules="[
                    {
                      validator: () => {
                        return scope.row.stock_num != null ? true : false;
                      },
                      message: $t('请输入库存'),
                    },
                  ]"
                >
                  <numInput
                    :controls="false"
                    :disabled="scope.row.material.length > 0"
                    :min="0"
                    :max="99999999"
                    :precision="0"
                    :placeholder="$t('请填写库存数量')"
                    v-model="scope.row.stock_num"
                  ></numInput>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="name" :label="$t('默认勾选')" width="400">
              <template #default="scope">
                <el-form-item
                  for="no_click"
                  class="mt16"
                  :prop="`model.product_feed[${scope.$index}].default_select`"
                  :rules="[
                    {
                      validator: () => {
                        return defaultSelect(form.model.feed_open_max_select, form.model.feed_max_select, form.model.product_feed) ? true : false;
                      },
                      message: $t('不能超过最多可选数量') + ' ' + form.model.feed_max_select,
                    },
                  ]"
                >
                  <el-checkbox @change="checkDefaultSelect" v-model="scope.row.default_select" size="large" :true-value="1" :false-value="0" />
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="name" :label="$t('关联材料')" width="120">
              <template #default="scope">
                {{ scope.row.material.length > 0 ? $t('是') : $t('否') }}
              </template>
            </el-table-column>
            <el-table-column prop="name" :label="$t('操作')" width="120">
              <template #default="scope">
                <el-button @click="deleteClick(scope.$index)" type="primary" link size="small"> {{ $t('删除') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </template>
        <!-- </el-form-item> -->
      </div>
    </div>
    <addFeed v-if="open_add_feed" :open="open_add_feed" :feed_ids="feed_ids" @close="handleClose" />
  </div>
</template>

<script setup>
  import { ref, inject, watch, nextTick, getCurrentInstance } from 'vue';
  import { useUserStore } from '@/store';
  import addFeed from '../components/addFeed.vue';
  import { languageStore } from '@/store/model/language.js';

  // 获取当前实例
  const { proxy } = getCurrentInstance();

  // 获取用户信息和语言数据
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const languageKey = languageStore().getLanguageKey().language.value;

  // 注入form
  const form = inject('form', {});

  // 响应式数据
  const open_add_feed = ref(false);
  const feed_ids = ref([]);

  // 定义emit
  const emit = defineEmits(['validateField']);

  // 监听form变化
  watch(
    () => form,
    (val) => {
      // 遍历val.model.product_feed
      (val.model.product_feed || []).map((item, index) => {
        //已有的加料
        feed_ids.value = [];
        feed_ids.value.push(item.feed_id);

        // 定义一个空数组arr
        let arr = [];
        // 遍历item.material
        (item.material || []).map((items, indexs) => {
          // 在下次DOM更新循环结束之后执行延迟回调
          nextTick(() => {
            // 将items.material_num转换为数字，并赋值给this.form.model.product_feed[index].material[indexs].material_num
            form.model.product_feed[index].material[indexs].material_num = Number(String(items.material_num).replace(/(\.\d{1,4})\d*/, '$1'));
          });
          // 定义一个变量num，初始值为0
          let num = 0;
          // 将this.form.ing_select_list[index][indexs].sku[0].material_stock除以items.material_num，并向下取整，赋值给num
          if (form.ing_select_list[index][indexs]?.sku) {
            num = Number(form.ing_select_list[index][indexs]?.sku[0]?.material_stock) / Number(items.material_num);
          }
          num = Math.floor(num);
          // 将num添加到arr中
          arr.push(num);
        });
        // 如果item.material的长度大于0
        if ((item.material || []).length > 0) {
          // 将arr排序，取最小值，赋值给this.form.model.product_feed[index].stock_num
          form.model.product_feed[index].stock_num =
            arr.sort((a, b) => a - b)[0] == Infinity ? null : arr.sort((a, b) => a - b)[0] > 99999999 ? 99999999 : arr.sort((a, b) => a - b)[0];
        }
      });
      if (val.model.product_feed.length == 0) {
        form.model.feed_required = 0;
        form.model.feed_open_max_select = 0;
        form.model.feed_max_select = 0;
      }
    },
    // 深度监听
    { deep: true, immediate: true }
  );

  // 方法定义
  const addIngredients = () => {
    feed_ids.value = [];
    form.model.product_feed.map((item) => {
      feed_ids.value.push(item.feed_id);
    });
    open_add_feed.value = true;
  };

  const handleDelete = () => {
    form.model.product_feed = [];
  };

  const handleClose = async (val) => {
    if (val === false) {
      open_add_feed.value = false;
      return;
    }

    // 第一步：构建新的加料id数组
    const newIds = await Promise.resolve().then(() => {
      const ids = [];
      (val || []).map((item) => {
        ids.push(item.feed_id);
      });
      return ids;
    });

    // 第二步：等上一步完成，计算需要新增的id
    const _pushIds = await Promise.resolve().then(() => {
      return newIds.filter((item) => {
        return form.model.product_feed.findIndex((items) => items.feed_id == item) == -1;
      });
    });

    // 第三步：等上一步完成，计算需要删除的id
    const _needToDeleteIds = await Promise.resolve().then(() => {
      return form.model.product_feed.filter((item) => {
        return newIds.findIndex((items) => items == item.feed_id) == -1;
      });
    });

    // 第四步：等上一步完成，执行删除操作
    await Promise.resolve().then(() => {
      _needToDeleteIds.map((item) => {
        form.model.product_feed.map((items, indexs) => {
          if (item.feed_id == items.feed_id) {
            form.model.product_feed.splice(indexs, 1);
            form.ing_select_list.splice(indexs, 1);
          }
        });
      });
    });

    // 第五步：等上一步完成，添加新的加料数据
    await Promise.resolve().then(() => {
      (val || []).map((item, index) => {
        if (_pushIds.indexOf(item.feed_id) != -1) {
          form.ing_select_list.push([]);
          form.model.product_feed.push({
            feed_id: item.feed_id,
            feed_name: item.feed_name,
            material: item.material,
            stock_num: null,
            price: Number(item.price) || null,
            default_select: '0',
          });
          form.ing_select_list[index] = item.material;
        }
      });
    });

    // 第六步：等上一步完成，处理材料选择列表
    await Promise.resolve().then(() => {
      (form.model.product_feed || []).map((item, index) => {
        if (form.ing_select_list[index].length > 0) {
          form.ing_select_list[index].map((items, indexs) => {
            form.ing_select_list[index][indexs].sku = [];
            form.ing_select_list[index][indexs].sku[0] = {
              product_id: '',
              material_stock: '',
            };
            form.ing_select_list[index][indexs].sku[0].product_id = items.material_id;
            form.ing_select_list[index][indexs].sku[0].material_stock = items.materialProduct.product_material_stock;
            form.ing_select_list[index][indexs].product_unit_text = items.materialProduct.product_unit_text;
            form.ing_select_list[index][indexs].product_name_text = items.materialProduct.product_name_text;
          });
        }
      });
    });

    // 第七步：等上一步完成，关闭弹窗
    await Promise.resolve().then(() => {
      open_add_feed.value = false;
    });
  };

  const defaultSelect = (feed_open_max_select, feed_max_select, product_feed) => {
    let count = 0;
    product_feed.map((item) => {
      if (item.default_select === 1) {
        count++;
      }
    });
    if (count > feed_max_select && feed_open_max_select == '1' && feed_max_select != null) {
      return false;
    } else {
      return true;
    }
  };

  const checkDefaultSelect = () => {
    nextTick(() => {
      form.model.product_feed.map((item, index) => {
        emit('validateField', `model.product_feed[${index}].default_select`);
      });
    });
  };

  const onBlur = () => {
    form.model.feed_max_select = form.model.feed_max_select == null ? 1 : form.model.feed_max_select;
    nextTick(() => {
      form.model.product_feed.map((item, index) => {
        emit('validateField', `model.product_feed[${index}].default_select`);
      });
    });
  };

  const deleteClick = (index) => {
    form.model.product_feed.splice(index, 1);
    form.ing_select_list.splice(index, 1);
    nextTick(() => {
      form.model.product_feed.map((item, indexs) => {
        emit('validateField', `model.product_feed[${indexs}].default_select`);
      });
    });
  };
</script>

<style scoped>
  .delete-icon {
    cursor: pointer;
    font-size: 24px;
    margin-left: auto;
  }

  .product-attr {
    width: 100%;
    box-shadow: 0 0 0 1px var(--el-input-border-color, var(--el-border-color)) inset;
    padding: 16px 16px 0 16px;
    border-radius: var(--el-input-border-radius, var(--el-border-radius-base));
    margin-bottom: 12px;
  }

  :deep(.inline-input) {
    width: 100%;
  }

  .product-tips {
    color: var(--el-color-tips);
  }

  .product-box {
    display: flex;
  }

  .materials-one {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 8px;

    .delete-icon {
      cursor: pointer;
      font-size: 24px;
      margin-top: -16px;
      margin-right: 0;
    }
  }

  .max-w230 {
    width: 100%;
  }
  .mt--16 {
    margin-top: -16px;
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
</style>

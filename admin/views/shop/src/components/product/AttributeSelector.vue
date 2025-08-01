<template>
  <el-dialog
    class="attribute-selector"
    @close="handleClose"
    v-model="dialogVisible"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :title="$t('选择属性')"
    append-to-body
  >
    <!-- {{ props.selectedAttributeIds.length }}
    {{ selectedAttributesTmp.length }}
    {{ loading }} -->
    <!-- {{ attributeGroupsTreeCurrentKey }} -->
    <!-- <pre style="font-size: 12px">{{ selectedAttributesTmp }}</pre> -->
    <div class="attribute-selector-content" v-loading="loading">
      <div class="attribute-selector-tree">
        <el-tree-v2
          ref="attributeGroupsTreeRef"
          :height="480"
          :data="attributeGroupsTree"
          node-key="id"
          highlight-current
          :current-node-key="attributeGroupsTreeCurrentKey"
          @current-change="handleAttributeGroupsTreeCurrentChange"
          auto-expand-parent
          :expand-on-click-node="false"
          :default-expanded-keys="attributeGroupsTreeExpandedKeys"
          @node-expand="handleAttributeGroupsTreeExpand"
          @node-collapse="handleAttributeGroupsTreeExpand"
          :props="{ children: 'children', label: 'label', disabled: 'disabled' }"
        >
          <template #default="{ node }">
            <div v-show="false" :style="{ marginRight: '4px' }" @click.stop>
              <el-checkbox
                v-model="attributeGroupsIsAllSelected[`attr-${node.key}`]"
                @change="(v) => handleAttributeGroupsCheck(node.key, v)"
                :disabled="attributeGroupsCount[`attr-${node.key}`] === 0"
              />
            </div>
            <span>{{ node.label }}</span>
            <template v-if="attributeGroupsSelectedCount[`attr-${node.key}`] > 0">
              <span style="margin-left: 2px">({{ attributeGroupsSelectedCount[`attr-${node.key}`] }})</span>
            </template>
          </template>
        </el-tree-v2>
      </div>
      <!-- <div class="attribute-selector-divider" /> -->
      <div class="attribute-selector-main">
        <div class="attribute-selector-form">
          <el-form size="small" ref="formRef" :model="form" :inline="true">
            <el-form-item :label="$t('属性组属性值')" :placeholder="$t('请输入属性组属性值')">
              <el-input size="small" v-model="form.attribute_name" @input="onDebounceSearch" />
            </el-form-item>
            <el-form-item>
              <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
                {{ $t('查询') }}
              </el-button>
            </el-form-item>
          </el-form>
        </div>
        <div class="attribute-selector-table">
          <el-auto-resizer>
            <template #default="{ height, width }">
              <el-table
                ref="attributesTableRef"
                :height="height"
                :style="{ width }"
                fixed
                :data="attributesTableData"
                size="small"
                border
                @select="handleSelect"
                @select-all="handleSelectAll"
              >
                <el-table-column type="selection" width="40" />
                <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center"></el-table-column>
                <el-table-column prop="parent_attribute_name_text" :label="$t('属性组')" width="140">
                  <template #default="{ row }">
                    {{ attributeGroups.find((item) => item.attribute_id === row.parent_id)?.attribute_name_text }}
                  </template>
                </el-table-column>
                <el-table-column prop="attribute_name_text" :label="$t('属性值')" />
              </el-table>
            </template>
          </el-auto-resizer>
        </div>
      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button
          class="dialog-add"
          type="primary"
          size="small"
          icon="Plus"
          @click="
            () => {
              open_add = true;
            }
          "
          >{{ $t('新增属性') }}</el-button
        >
        <el-button @click="handleClose">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </span>
    </template>
  </el-dialog>
  <!--添加-->
  <Add v-if="open_add" :open_add="open_add" @closeDialog="closeDialogFunc($event, 'add')"></Add>
</template>

<script setup>
  import { ref, reactive, computed, nextTick, getCurrentInstance } from 'vue';
  import ProductApi from '@/api/product.js';
  import Add from '@/views/product/expand/attr/add.vue';

  defineOptions({
    name: 'ProductAttributeSelector',
  });

  const { proxy } = getCurrentInstance();

  const emit = defineEmits(['close']);

  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
    selectedAttributeIds: {
      type: Array,
      default: () => [],
    },
  });

  const dialogVisible = ref(props.open);

  const loading = ref(false);

  const isFirstRender = ref(true);

  const open_add = ref(false);

  const form = reactive({
    attribute_name: '',
  });

  const formRef = ref(null);

  const attributeGroups = ref([]);
  const attributes = ref([]);

  const attributeGroupsCount = computed(() => {
    const _map = {};

    let count = 0;
    for (const item of attributeGroups.value) {
      const _count = attributes.value.filter((attr) => attr.parent_id === item.attribute_id).length;
      _map[`attr-${item.attribute_id}`] = _count;
      count += _count;
    }

    _map[`attr-0`] = count;
    return _map;
  });

  const attributeGroupsTree = computed(() => {
    return [
      {
        id: 0,
        label: $t('全部'),
        children: loading.value
          ? undefined
          : Array.isArray(attributeGroups.value)
          ? attributeGroups.value.map((item) => {
              return {
                id: item.attribute_id,
                pid: 0,
                label: item.attribute_name_text,
              };
            })
          : undefined,
      },
    ];
  });

  const attributeGroupsTreeRef = ref(null);
  const attributeGroupsTreeCurrentKey = ref(0);
  const attributeGroupsTreeExpandedKeys = ref([0]);

  const attributesTableData = computed(() => {
    return (
      attributes.value
        .filter(
          (item) => attributeGroupsTreeCurrentKey.value === 0 || attributeGroupsTreeCurrentKey.value === item.parent_id || attributeGroupsTreeCurrentKey.value === item.attribute_id
        )
        // FIXME: 属性组名称和属性名称都支持搜索
        .filter((item) => !searchValue.value || item.parent_attribute_name_text.includes(searchValue.value) || item.attribute_name_text.includes(searchValue.value))
    );
  });

  const selectedAttributesTmp = ref([]);

  const attributeGroupsSelectedCount = computed(() => {
    const _map = {};

    let count = 0;
    for (const item of attributeGroups.value) {
      const _count = selectedAttributesTmp.value.filter((attr) => attr.parent_id === item.attribute_id).length;
      _map[`attr-${item.attribute_id}`] = _count;
      count += _count;
    }

    _map[`attr-0`] = count;
    return _map;
  });

  const attributeGroupsIsAllSelected = computed({
    get() {
      const _map = {};

      for (const item of attributeGroups.value) {
        _map[`attr-${item.attribute_id}`] =
          attributeGroupsCount.value[`attr-${item.attribute_id}`] > 0 &&
          attributeGroupsSelectedCount.value[`attr-${item.attribute_id}`] === attributeGroupsCount.value[`attr-${item.attribute_id}`];
      }

      return _map;
    },
    set(val) {
      console.log(val);
    },
  });

  const pagination = ref({
    page: 1,
    pageSize: 10000,
    total: 0,
    totalPage: 0,
  });

  const getAttributes = async (f) => {
    loading.value = true;
    try {
      const res = await ProductApi.AttributeList(
        {
          page: pagination.value.page,
          list_rows: pagination.value.pageSize,
          type: 2,
        },
        true
      );

      attributes.value = res.data.list.data;

      pagination.value.total = res.data.list.total;
      pagination.value.totalPage = res.data.list.total_page;
      pagination.value.page = res.data.list.current_page;
      pagination.value.pageSize = res.data.list.per_page;
      await toggleRowSelection(f);
    } catch (error) {
      //
    } finally {
      loading.value = false;
    }
  };

  const getAttributeGroups = async () => {
    try {
      const res = await ProductApi.AttributeList(
        {
          page: pagination.value.page,
          list_rows: pagination.value.pageSize,
          type: 1,
        },
        true
      );

      attributeGroups.value = res.data.list.data;
    } catch (error) {
      //
    } finally {
      //
    }
  };

  getAttributeGroups();
  getAttributes();

  const onSubmit = () => {
    emit('close', selectedAttributesTmp.value, attributes.value);
    reset();
  };

  const reset = () => {
    selectedAttributesTmp.value = [];
  };

  const handleClose = () => {
    emit('close');
    reset();
  };

  const attributesTableRef = ref(null);

  const handleSelect = (data, node) => {
    const isChecked = data.some((item) => item.attribute_id === node.attribute_id);
    if (isChecked) {
      if (selectedAttributesTmp.value.every((item) => item.attribute_id !== node.attribute_id)) {
        selectedAttributesTmp.value.push(node);
      }
    } else {
      selectedAttributesTmp.value = selectedAttributesTmp.value.filter((item) => item.attribute_id !== node.attribute_id);
    }
    toggleRowSelection(false);
  };

  const handleSelectAll = (data) => {
    handleAttributesTableSelectAll(data);
  };

  const toggleRowSelection = async (f) => {
    try {
      if (isFirstRender.value) {
        selectedAttributesTmp.value = attributesTableData.value.filter((item) => props.selectedAttributeIds.includes(item.attribute_id));
      }
      if (f == 'add') {
        let oldIds = selectedAttributesTmp.value.map((item) => {
          return item.attribute_id;
        });
        let arr = attributes.value.filter((item) => oldIds.includes(item.attribute_id));
        selectedAttributesTmp.value = arr;
      }
      await nextTick();
      isFirstRender.value = false;
      await new Promise((resolve) => {
        selectedAttributesTmp.value.forEach((item) => {
          attributesTableRef.value.toggleRowSelection(item, true, true);
        });
        resolve();
      });
    } catch (err) {
      console.log(err);
    } finally {
      //
    }
  };

  let searchTimer;
  const onDebounceSearch = () => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
      onSearch();
    }, 200);
  };

  const searchValue = ref('');

  const onSearch = () => {
    searchValue.value = form.attribute_name;

    nextTick(() => {
      toggleRowSelection(false);
    });
  };

  const handleAttributeGroupsTreeCurrentChange = async ({ id }) => {
    if (id === attributeGroupsTreeCurrentKey.value) return;
    attributeGroupsTreeCurrentKey.value = id;
    attributeGroupsTreeRef.value.setCurrentKey(id);

    toggleRowSelection(false);
  };

  const handleAttributeGroupsCheck = (id, checked) => {
    const _attributes = attributes.value.filter((item) => item.attribute_id === id);

    if (id === attributeGroupsTreeCurrentKey.value) {
      _attributes.forEach((item) => {
        attributesTableRef.value.toggleRowSelection(item, checked, true);
      });
    }

    if (checked) {
      _attributes.forEach((item) => {
        if (selectedAttributesTmp.value.some((attribute) => attribute.attribute_id === item.attribute_id)) return;
        selectedAttributesTmp.value.push(item);
      });
    } else {
      selectedAttributesTmp.value = selectedAttributesTmp.value.filter((item) => !_attributes.some((attribute) => attribute.attribute_id === item.attribute_id));
    }
  };

  const handleAttributeGroupsTreeExpand = ({ id }, { expanded }) => {
    if (expanded) {
      attributeGroupsTreeExpandedKeys.value = [...attributeGroupsTreeExpandedKeys.value, id];
    } else {
      if (id === 0) {
        attributeGroupsTreeExpandedKeys.value = [];
        return;
      }
      attributeGroupsTreeExpandedKeys.value = attributeGroupsTreeExpandedKeys.value.filter((item) => item !== id);
    }
  };

  const handleAttributesTableSelectAll = (val) => {
    if (attributeGroupsTreeCurrentKey.value === 0) {
      selectedAttributesTmp.value = val;
      toggleRowSelection(false);
      return;
    }

    const filterCondition = (item) => item.parent_id !== attributeGroupsTreeCurrentKey.value;
    const selectedAttributesWithoutCurrent = selectedAttributesTmp.value.filter(filterCondition);
    selectedAttributesTmp.value = [...new Set([...selectedAttributesWithoutCurrent, ...val].map((item) => item.attribute_id))].map((id) =>
      [...selectedAttributesWithoutCurrent, ...val].find((item) => item.attribute_id === id)
    );

    toggleRowSelection(false);
  };

  /*关闭弹窗*/
  const closeDialogFunc = async (e, f) => {
    if (f == 'add') {
      open_add.value = e.openDialog;
      if (e.type == 'success') {
        await getAttributeGroups();
        await getAttributes('add');
      }
    }
  };

  // const handleAttributesTreeExpandByRowSelection = async (val) => {
  //   if (!val.length) {
  //     attributesTreeExpandedKeys.value = [0];
  //     return;
  //   }
  //   for (const item of val) {
  //     const pid = getAttributesPid(item.category_id);
  //     if (typeof pid !== 'number') continue;
  //     if (attributesTreeExpandedKeys.value.includes(pid)) continue;
  //     attributesTreeExpandedKeys.value.push(pid);
  //   }
  // };
</script>

<style lang="scss" scoped>
  .attribute-selector-content {
    display: flex;
    justify-content: flex-start;
    align-items: stretch;
    gap: 8px;

    .attribute-selector-tree {
      width: 240px;
      flex-shrink: 0;
      overflow-x: hidden;
    }

    .attribute-selector-divider {
      margin: 0 4px;
      width: 2px;
      flex-shrink: 0;
      background-color: #f0f2f5;
    }

    .attribute-selector-main {
      flex-grow: 1;
      display: flex;
      flex-direction: column;
      gap: 8px;

      .attribute-selector-form {
        flex-shrink: 0;
      }

      .attribute-selector-table {
        flex-grow: 1;
      }
    }
  }
  .dialog-add {
    float: left;
  }
</style>

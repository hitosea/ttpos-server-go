<template>
  <div class="pt-5 flex justify-end">
    <el-pagination
      :disabled="props.disabled"
      :currentPage="currentPage"
      :pageSize="pageSize"
      :page-sizes="props.pageSizes"
      :layout="props.layout"
      :total="props.total"
      :pager-count="5"
      :hide-on-single-page="false"
      background
      @size-change="pageSizeChange"
      @current-change="currentPageChange"
    ></el-pagination>
  </div>
</template>

<script lang="ts" setup>
  const props = withDefaults(
    defineProps<{
      disabled?: boolean;
      currentPage?: number;
      pageSize?: number;
      total?: number;
      pageSizes?: number[];
      layout?: string;
    }>(),
    {
      disabled: false,
      currentPage: 1,
      pageSize: 15,
      total: 0,
      pageSizes: () => [15, 20, 30, 40],
      layout: 'total, prev, pager, next, jumper',
    },
  );
  const emits = defineEmits<{
    (e: 'change', page?: number, pageSize?: number): void;
  }>();

  const pageSizeChange = (value: number) => {
    emits('change', 1, value);
  };
  const currentPageChange = (value: number) => {
    emits('change', value, props.pageSize);
  };
</script>

<style lang="scss" scoped></style>

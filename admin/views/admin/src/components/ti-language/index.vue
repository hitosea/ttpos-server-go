<template>
  <el-dropdown trigger="hover" @command="setLanguage">
    <div class="flex items-center gap-1 text-[#100a05] text-sm cursor-pointer">
      <ti-icon name="language" :size="props.iconSize" />
      <template v-if="props.showText">
        <span>{{ languageText }}</span>
        <el-icon><arrow-down /></el-icon>
      </template>
    </div>
    <template #dropdown>
      <el-dropdown-menu class="ti-language-dropdown">
        <el-dropdown-item v-for="item in languageData()" :key="item.lang" :command="item.lang" :disabled="getLanguage() == item.lang">
          <div class="flex items-center justify-between gap-4">
            <span>{{ item.text }}</span>
            <ti-icon v-if="getLanguage() == item.lang" name="check" />
          </div>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { languageData, getLanguage, setLanguage } from '@/i18n';

  const props = withDefaults(
    defineProps<{
      showText?: boolean;
      iconSize?: number;
    }>(),
    {
      showText: true,
      iconSize: 16,
    },
  );

  const languageText = computed(() => {
    return languageData().find((item) => item.lang == getLanguage())?.text;
  });
</script>

<style lang="scss" scoped>
  .ti-language-dropdown {
    :deep(.is-disabled) {
      color: #ffbe00;
      background-color: #fff6de;
    }
  }
</style>

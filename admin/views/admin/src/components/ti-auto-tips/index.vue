<template>
  <el-tooltip :placement="props.placement" :effect="props.tooltipTheme" :delay="props.delay" :disabled="!showTooltip || props.showTooltips" transfer>
    <template #content>
      <div :style="{ maxWidth: `${props.tooltipMaxWidth}px` }">{{ tipText }}</div>
    </template>
    <div ref="contentRef" class="truncate" @mouseenter="handleTooltipIn" @click="handleClick">
      <template v-if="existSlot && getTexts(1)">
        <div style="visibility: hidden; position: absolute"><slot /></div>
        <div class="w-full flex leading-[1.5]">
          <span class="truncate">{{ getTexts(2) }}</span>
          <span class="whitespace-nowrap" style="text-overflow: initial">{{ getTexts(3) }}</span>
        </div>
      </template>
      <template v-else-if="existSlot"><slot /></template>
      <template v-else>{{ content }}</template>
    </div>
  </el-tooltip>
</template>

<script setup lang="ts">
  import { ref, computed, getCurrentInstance } from 'vue';
  //
  const emits = defineEmits<{
    (e: 'click'): void;
  }>();
  const props = withDefaults(
    defineProps<{
      content?: string | number;
      placement?: string;
      tooltipTheme?: string;
      tooltipMaxWidth?: string | number;
      delay?: number;
      showTooltips?: boolean;
      forcedShow?: boolean;
    }>(),
    {
      content: '',
      placement: 'top',
      tooltipTheme: 'dark',
      tooltipMaxWidth: 350,
      delay: 100,
      showTooltips: false,
      forcedShow: false,
    },
  );
  //
  const instance = getCurrentInstance();
  const contentRef = ref();
  const tooltipContent = ref<string>('');
  const showTooltip = ref<boolean>(false);
  const tipText = computed(() => props.content || tooltipContent.value || '');
  const existSlot = computed(() => !(instance?.slots?.default === undefined || instance?.slots?.default().length < 1));

  const getTexts = (type: number = 1) => {
    let text = instance?.slots?.default ? (instance?.slots?.default() as any)[0]?.text : '';
    let finallyText = '';
    let texts = '';
    if (text) {
      let finallyNum = 3;
      if (text.length < 6) {
        finallyNum = 2;
      }
      if (text.length > 12) {
        finallyNum = 6;
      }
      texts = text.substring(0, text.length - finallyNum);
      finallyText = text.substring(text.length - finallyNum);
    }
    if (type == 1) {
      return text;
    } else if (type == 2) {
      return texts;
    } else if (type == 3) {
      return finallyText;
    }
  };

  const handleTooltipIn = () => {
    let range: any = document.createRange();
    range.setStart(contentRef.value, 0);
    range.setEnd(contentRef.value, contentRef.value?.childNodes.length);
    const rangeWidth = range.getBoundingClientRect().width;
    showTooltip.value = props.forcedShow || Math.floor(rangeWidth) > Math.floor(contentRef.value?.offsetWidth);
    if (showTooltip.value && existSlot.value) {
      const tmpArray = instance?.slots?.default
        ? (instance?.slots?.default() as any)?.map((e: any) => {
            if (e?.text) return e.text;
            if (e?.elm?.innerText) return e.elm.innerText;
            return '';
          })
        : [];
      tooltipContent.value = tmpArray.join('');
    }
    range = null;
  };

  const handleClick = () => {
    emits('click');
  };
</script>

<style lang="scss" scoped></style>

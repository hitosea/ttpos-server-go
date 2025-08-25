<template>
  <el-tooltip 
    :content="content" 
    :placement="placement" 
    :effect="tooltipTheme" 
    :delay="delay" 
    :disabled="!shouldShowTooltip" 
    :max-width="tooltipMaxWidth" 
    transfer
  >
    <span 
      ref="contentRef" 
      class="common-auto-tip" 
      @click="onClick"
    >
      <slot>{{ content }}</slot>
    </span>
  </el-tooltip>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, watch } from 'vue';

// 定义props
const props = defineProps({
  content: {
    type: [String, Number],
    default: '',
  },
  placement: {
    type: String,
    default: 'top',
  },
  tooltipTheme: {
    type: String,
    default: 'dark',
  },
  tooltipMaxWidth: {
    type: [String, Number],
    default: 350,
  },
  delay: {
    type: Number,
    default: 100,
  },
  forcedShow: {
    type: Boolean,
    default: false,
  },
});

// 定义emits
const emit = defineEmits(['on-click']);

// 响应式数据
const shouldShowTooltip = ref(false);
const contentRef = ref(null);

// 检测文本是否超出
const checkTextOverflow = async () => {
  if (!contentRef.value) return;
  
  await nextTick();
  
  try {
    const element = contentRef.value;
    const isOverflow = element.scrollWidth > element.offsetWidth;
    
    shouldShowTooltip.value = props.forcedShow || isOverflow;
    
    // 调试信息
    console.log('文本溢出检测:', {
      scrollWidth: element.scrollWidth,
      offsetWidth: element.offsetWidth,
      isOverflow,
      shouldShow: shouldShowTooltip.value
    });
    
  } catch (error) {
    console.error('检测文本溢出时出错:', error);
    shouldShowTooltip.value = true;
  }
};

// 点击事件
const onClick = (e) => {
  emit('on-click', e);
};

// 组件挂载后初始化
onMounted(() => {
  // 延迟检查，确保DOM完全渲染
  setTimeout(() => {
    checkTextOverflow();
  }, 100);
});

// 监听内容变化
watch(() => props.content, () => {
  nextTick(() => {
    checkTextOverflow();
  });
});
</script>

<style lang="scss" scoped>
.common-auto-tip {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 310px; // 与你的CSS保持一致
}

:deep(.el-popper) {
  max-width: 450px !important;
}
</style>

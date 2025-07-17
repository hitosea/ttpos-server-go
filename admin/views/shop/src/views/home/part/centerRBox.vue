<template>
  <div class="center-r-box">
    <div class="center-r-box-t">
      <h4>{{ title }}</h4>
      <p @click="handleClick">{{ description }}</p>
    </div>
    <p class="fb">{{ value }}</p>
  </div>
</template>
<script setup>
  const props = defineProps({
    title: {
      type: String,
      default: '',
    },
    description: {
      type: String,
      default: '',
    },
    value: {
      type: [String, Number],
      default: '',
    },
  });
  const emit = defineEmits(['click']);
  const handleClick = () => {
    emit('click');
  };
</script>
<style scoped lang="scss">
  .center-r-box {
    padding: 16px;
    border-radius: 4px;
    background: #fff1ee;
    margin-bottom: 12px;
    &:nth-last-child(1) {
      margin-bottom: 0;
    }
  }

  .center-r-box-t {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    margin-bottom: 16px;
  }

  .center-r-box-t h4 {
    color: var(--el-color-black);
    font-size: 16px;
    font-weight: 400;
  }

  .center-r-box-t p {
    color: var(--el-color-primary);
    font-size: 14px;
    font-weight: 400;
    cursor: pointer;
  }

  .fb {
    color: #ff6241;
    font-size: 20px;
    font-style: normal;
    font-weight: 700;
    line-height: normal;
  }
</style>

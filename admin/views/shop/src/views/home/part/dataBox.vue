<template>
  <div class="data-box">
    <div class="data-box-title">
      <h3>{{ title }}</h3>
      <el-tooltip class="item" effect="dark" :content="content" placement="bottom">
        <template #content>
          <div v-html="content"></div>
        </template>
        <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
      </el-tooltip>
    </div>
    <h4>{{ value }}</h4>
  </div>
</template>
<script setup>
  const props = defineProps({
    title: {
      type: String,
      default: '',
    },
    content: {
      type: [String, Number],
      default: '',
    },
    value: {
      type: [String, Number],
      default: '',
    },
  });
</script>
<style scoped lang="scss">
  .data-box {
    flex: 1;
    padding: 16px;
    background: #fff6de;
    border-radius: 4px;
    display: flex;
    flex-direction: column;
  }

  .data-box-title {
    display: flex;
    justify-content: space-between;
  }

  .data-box-title h3 {
    font-size: 16px;
    font-weight: 400;
    color: var(--el-color-black);
    margin-bottom: 16px;
  }

  .data-box h4 {
    font-size: 20px;
    font-weight: 700;
    margin-top: auto;
    color: var(--el-color-black);
  }

  .data-box:hover {
    background: #ffbe00;
  }

  .data-box:hover .data-box-icon {
    color: #fff;
  }

  .data-box-icon {
    width: 24px;
    height: 24px;
    color: #ffbe00;
  }
</style>

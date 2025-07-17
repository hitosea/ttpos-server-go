<template>
  <div class="grid-content">
    <p class="des">
      {{ title }}
      <el-tooltip v-if="tooltipContent" class="item" effect="dark" :content="tooltipContent" placement="bottom">
        <SvgIcon class="data-icon" name="icon6"></SvgIcon>
      </el-tooltip>
    </p>
    <h3>
      {{ value }}
    </h3>
    <p class="yesterday"> {{ yesterdayData }} </p>
  </div>
</template>
<script setup>
  const props = defineProps({
    title: {
      type: String,
      default: '',
    },
    yesterdayData: {
      type: [String, Number],
      default: '',
    },
    value: {
      type: [String, Number],
      default: '',
    },
    tooltipContent: {
      type: [String, Number],
      default: '',
    },
  });
</script>
<style scoped lang="scss">
  .grid-content {
    border-radius: 4px;
    padding: 16px;
    background: #f7f9fb;
    flex: 1;
    min-width: calc(33.3% - 8px);
    max-width: 30%;
    display: flex;
    flex-direction: column;
  }

  .grid-content .des {
    color: var(--el-color-black);
    font-size: 16px;
    font-style: normal;
    font-weight: 400;
    margin-bottom: 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .grid-content .des .data-icon {
    width: 20px;
    height: 20px;
    color: #ffbe00;
  }

  .grid-content h3 {
    color: var(--el-color-black);
    font-size: 20px;
    font-style: normal;
    font-weight: 700;
    margin-bottom: 8px;
    margin-top: auto;
  }

  .grid-content .yesterday {
    color: var(--el-color-tips);
    font-size: 12px;
    font-style: normal;
    font-weight: 400;
  }
</style>

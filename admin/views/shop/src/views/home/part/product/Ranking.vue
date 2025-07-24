<template>
  <div class="right-box d-s-s d-c">
    <div class="ww100 right-box-t">
      <div :class="activeName == 'sale' ? 'active' : ''" @click="handleClick('sale')">{{ $t('销量TOP10') }}</div>
      <div :class="activeName == 'view' ? 'active' : ''" @click="handleClick('view')">{{ $t('销售额TOP10') }}</div>
    </div>
    <div class="list ww100">
      <el-table v-if="listData.length > 0" :data="listData" style="width: 100%" size="small">
        <el-table-column prop="product_name_text" :label="$t('商品名称')">
          <template #default="scope">
            <div class="product-name">
              <span :class="scope.$index < 3 ? 'key-box' : 'key-box2'">{{ scope.$index + 1 }}</span>
              <span class="">{{ scope.row.product_name_text }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="total_num" :label="$t('销量')"> </el-table-column>
        <el-table-column prop="total_price" :label="$t('销售额')">
          <template #default="scope">
            {{ $formatPrice(scope.row.total_price) }}
          </template>
        </el-table-column>
      </el-table>
      <div v-else class="tc pt30">{{ $t('暂无上榜记录') }}</div>
    </div>
  </div>
</template>

<script setup>
  import { ref, computed } from 'vue';

  const props = defineProps({
    product_data: {
      type: Object,
      default: () => ({}),
    },
  });

  // 当前激活tab
  const activeName = ref('sale');

  // 列表数据根据tab动态切换，保证响应式
  const listData = computed(() => {
    return activeName.value == 'sale' ? props.product_data?.salesNumRank || [] : props.product_data?.salesMoneyRank || [];
  });

  // 切换tab
  const handleClick = (type) => {
    activeName.value = type;
  };
</script>

<style scoped>
  .right-box {
    width: 100%;
    box-sizing: border-box;
  }

  .Echarts > div {
    height: 400px;
  }
  .product-name {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .key-box {
    display: block;
    width: 20px;
    height: 20px;
    font-size: 12px;
    line-height: 20px;
    border-radius: 50%;
    font-weight: bold;
    text-align: center;
    color: var(--el-color-black);
    background: var(--el-color-primary);
    font-weight: 700;
    flex-shrink: 0;
  }
  .key-box2 {
    display: block;
    width: 20px;
    height: 20px;
    font-size: 12px;
    line-height: 20px;
    border-radius: 50%;
    font-weight: bold;
    text-align: center;
    color: var(--el-color-primary);
    background: #fff6de;
    font-weight: 700;
    flex-shrink: 0;
  }
  .right-box .list img {
    width: 30px;
    height: 30px;
  }

  .right-box-t {
    display: flex;
    align-items: center;
    gap: 24px;
    margin-bottom: 16px;
  }

  .right-box-t > div {
    color: var(--el-color-tips);
    font-size: 16px;
    font-style: normal;
    font-weight: 400;
    cursor: pointer;
    text-transform: capitalize;
  }

  .right-box-t > div.active {
    color: var(--el-color-black);
    font-size: 20px;
    font-weight: 600;
  }
</style>

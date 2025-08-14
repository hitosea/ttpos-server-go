<template>
  <div class="product-bottom">
    <div class="flex-1">
      <div class="right-box d-s-s d-c pr16">
        <div class="common-form">{{ $t('销量TOP10') }}</div>
        <div class="list ww100">
          <el-table v-if="salesNumRank.length > 0" :data="salesNumRank" style="width: 100%" size="small">
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
                {{ proxy.$formatPrice(scope.row.total_price) }}
              </template>
            </el-table-column>
          </el-table>
          <div v-else class="tc pt30">{{ $t('暂无上榜记录') }}</div>
        </div>
      </div>
    </div>
    <div class="flex-1">
      <div class="right-box d-s-s d-c pr16">
        <div class="common-form">{{ $t('销售额TOP10') }}</div>
        <div class="list ww100">
          <el-table v-if="salesMoneyRank.length > 0" :data="salesMoneyRank" style="width: 100%" size="small">
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
                {{ proxy.$formatPrice(scope.row.total_price) }}
              </template>
            </el-table-column>
          </el-table>
          <div v-else class="tc pt30">{{ $t('暂无上榜记录') }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
  import { getCurrentInstance } from 'vue';
  const { proxy } = getCurrentInstance();
  const props = defineProps({
    salesNumRank: {
      type: Array,
      default: () => [],
    },
    salesMoneyRank: {
      type: Array,
      default: () => [],
    },
  });
</script>
<style lang="scss" scoped>
  .product-bottom {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 16px;
  }


  .product-name {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .key-box {
    flex-shrink: 0;
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
  }

  .key-box2 {
    flex-shrink: 0;
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
  }
</style>

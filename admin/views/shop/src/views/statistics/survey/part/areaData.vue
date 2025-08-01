<template>
  <div class="area-data">
    <div class="area-data-item">
      <div class="common-form">{{ $t('区域数据') }}</div>
      <template v-if="regionData.length > 0">
        <div class="area-data-item-content">
          <div class="area-data-item-content-item" v-for="(item, index) in regionData" :key="index">
            <div class="area-data-item-content-item-title">
              <p>{{ item.area_name }}</p>
            </div>
            <div class="area-data-item-content-item-content">
              <div class="area-data-item-content-item-content-wrap">
                <p>
                  {{ $t('总销售额') }}
                </p>
                <p>
                  {{ proxy.$formatPrice(item.sales_price || 0) }}
                </p>
              </div>
              <div class="area-data-item-content-item-content-wrap">
                <p>
                  {{ $t('营业收入') }}
                </p>
                <p>
                  {{ proxy.$formatPrice(item.business_price || 0) }}
                </p>
              </div>
              <div class="area-data-item-content-item-content-wrap">
                <p>
                  {{ $t('商品数量') }}
                </p>
                <p>
                  {{ proxy.$formatPrice(item.product_num || 0) }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </template>
      <div v-else>{{ $t('暂无数据') }}</div>
    </div>
  </div>
</template>
<script setup>
  import { getCurrentInstance } from 'vue';
  const { proxy } = getCurrentInstance();
  const props = defineProps({
    regionData: {
      type: Array,
      default: () => [],
    },
  });
</script>
<style lang="scss" scoped>
  .area-data {
    margin-bottom: 16px;

    .area-data-item-content {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 12px;

      .area-data-item-content-item {
        border-radius: 4px;
        background: #fff6de;
        padding: 12px;
        display: flex;

        .area-data-item-content-item-title {
          flex-shrink: 0;
          min-width: 80px;
          height: 80px;
          background: var(--el-color-primary);
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 18px;
          border-radius: 4px;
        }

        .area-data-item-content-item-content {
          flex-grow: 1;
          padding-left: 12px;
          display: flex;
          flex-direction: column;
          justify-content: space-between;

          .area-data-item-content-item-content-wrap {
            display: flex;
            align-items: baseline;
            justify-content: space-between;
            font-size: 14px;
            font-weight: bold;
          }
        }
      }
    }
  }
</style>

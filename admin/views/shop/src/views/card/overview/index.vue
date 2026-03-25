<template>
  <div class="overview-content" v-loading="loading">
    <div class="common-form">{{ $t('账户数据') }}</div>
    <div class="operation-data">
      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('主账户余额') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('所有会员主账户中剩余的余额') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ proxy.$formatPrice(detail.balance || 0) }}</h4>
      </div>

      <div class="data-box">
        <div class="data-box-title">
          <h3>{{ $t('赠送账户余额') }}</h3>
          <el-tooltip class="item" effect="dark" placement="bottom">
            <template #content>
              <span>{{ $t('所有会员赠送账户中剩余的余额') }}</span>
            </template>
            <SvgIcon class="data-box-icon" name="icon6"></SvgIcon>
          </el-tooltip>
        </div>
        <h4>{{ proxy.$formatPrice(detail.gift_balance || 0) }}</h4>
      </div>
    </div>

    <!--内容-->
    <div class="product-bottom">
      <div class="flex-1">
        <div class="right-box d-s-s d-c">
          <div class="common-form">{{ $t('充值排行榜') }}</div>
          <template v-if="rechargeRank.length > 0">
            <div class="data-list">
              <div class="item" v-for="(item, index) in rechargeRank">
                <div class="p-relative p-relative-l">
                  <span class="num" :class="indexMethod(index, rechargeRankPage, rechargeRankPageSize) < 4 ? 'key-box' : 'key-box2'">{{
                    indexMethod(index, rechargeRankPage, rechargeRankPageSize)
                  }}</span>
                  <div class="l-autoTips">
                    <autoTips :content="item.nickname + '(' + item.user_id + ')'"> {{ item.nickname }}({{ item.user_id }})</autoTips>
                  </div>
                </div>
                <div class="p-relative">
                  <autoTips :content="$t('充值：') + $formatPrice(item.recharge_amount || 0) + ' (' + $t('额外赠送：') + $formatPrice(item.gift_amount || 0) + ')'" :textRight="true">
                    {{ $t('充值：') + $formatPrice(item.recharge_amount || 0) + ' (' + $t('额外赠送：') + $formatPrice(item.gift_amount || 0) + ')' }}
                  </autoTips>
                </div>
              </div>
            </div>
            <!--分页-->
            <div class="pagination">
              <el-pagination
                @current-change="handleCurrentChange"
                background
                :current-page="rechargeRankPage"
                :page-size="rechargeRankPageSize"
                layout="total, prev, pager, next, jumper"
                :total="rechargeRankTotal"
              >
              </el-pagination>
            </div>
          </template>
          <div v-else class="tc pt30">{{ $t('暂无上榜记录') }}</div>
        </div>
      </div>
      <div class="flex-1">
        <div class="right-box d-s-s d-c">
          <div class="common-form">
            {{ $t('消费排行榜') }}
            <el-radio-group v-model="consumerRankSort" class="radio-search" @change="sortChange">
              <el-radio-button :label="1">{{ $t('按消费次数') }}</el-radio-button>
              <el-radio-button :label="2">{{ $t('按消费金额') }}</el-radio-button>
            </el-radio-group>
          </div>
          <div class="list ww100">
            <template v-if="consumerRank.length > 0">
              <div class="data-list">
                <div class="item" v-for="(item, index) in consumerRank">
                  <div class="p-relative p-relative-l">
                    <span class="num" :class="indexMethod(index, consumerRankPage, consumerRankPageSize) < 4 ? 'key-box' : 'key-box2'">
                      {{ indexMethod(index, consumerRankPage, consumerRankPageSize) }}
                    </span>
                    <div class="l-autoTips">
                      <autoTips :content="item.nickname + '(' + item.user_id + ')'"> {{ item.nickname }}({{ item.user_id }})</autoTips>
                    </div>
                  </div>
                  <div class="p-relative">
                    <autoTips :content="$t('消费次数：') + item.consumption_num + ' ' + $t('消费金额：') + $formatPrice(item.consumption_amount || 0)" :textRight="true">
                      {{ $t('消费次数：') + item.consumption_num }}
                      {{ $t('消费金额：') + $formatPrice(item.consumption_amount || 0) }}
                    </autoTips>
                  </div>
                </div>
              </div>
              <!--分页-->
              <div class="pagination">
                <el-pagination
                  @current-change="handleCurrentChangeConsumer"
                  background
                  :current-page="consumerRankPage"
                  :page-size="consumerRankPageSize"
                  layout="total, prev, pager, next, jumper"
                  :total="consumerRankTotal"
                >
                </el-pagination>
              </div>
            </template>
            <div v-else class="tc pt30">{{ $t('暂无上榜记录') }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
  import { ref, getCurrentInstance, onMounted } from 'vue';
  import CardApi from '@/api/card.js';
  import autoTips from '@/components/autoTips/autoTips.vue';

  const { proxy } = getCurrentInstance();

  const loading = ref(false);
  const detail = ref({});
  const consumerRank = ref([]);
  const consumerRankSort = ref(1);
  const consumerRankPage = ref(1);
  const consumerRankPageSize = ref(10);
  const consumerRankTotal = ref(0);

  const rechargeRank = ref([]);
  const rechargeRankPage = ref(1);
  const rechargeRankPageSize = ref(10);
  const rechargeRankTotal = ref(0);

  const getData = async () => {
    loading.value = true;
    try {
      const res = await CardApi.getSurvey();
      detail.value = res.data.detail;
    } catch (error) {
      console.log(error);
    } finally {
      loading.value = false;
    }
  };

  const handleCurrentChange = (val) => {
    rechargeRankPage.value = val;
    getRechargeRank();
  };

  const getRechargeRank = async () => {
    loading.value = true;
    try {
      let Params = {
        page: rechargeRankPage.value,
        list_rows: rechargeRankPageSize.value,
      };
      const res = await CardApi.getRechargeRank(Params, true);
      rechargeRank.value = res.data.list.data;
      rechargeRankPage.value = res.data.list.current_page;
      rechargeRankPageSize.value = res.data.list.per_page;
      rechargeRankTotal.value = res.data.list.total;
    } catch (error) {
      console.log(error);
    } finally {
      loading.value = false;
    }
  };

  const handleCurrentChangeConsumer = (val) => {
    consumerRankPage.value = val;
    getConsumerRank();
  };

  const sortChange = () => {
    consumerRankPage.value = 1;
    getConsumerRank();
  };

  const getConsumerRank = async () => {
    loading.value = true;
    try {
      let Params = {
        sort: consumerRankSort.value,
        page: consumerRankPage.value,
        list_rows: consumerRankPageSize.value,
      };
      const res = await CardApi.getConsumerRank(Params, true);
      consumerRank.value = res.data.list.data;
      consumerRankPage.value = res.data.list.current_page;
      consumerRankPageSize.value = res.data.list.per_page;
      consumerRankTotal.value = res.data.list.total;
    } catch (error) {
      console.log(error);
    } finally {
      loading.value = false;
    }
  };

  const indexMethod = (index, Page, PageSize) => {
    return (Page - 1) * PageSize + index + 1;
  };

  onMounted(() => {
    getData();
    getRechargeRank();
    getConsumerRank();
  });
</script>
<style lang="scss" scoped>
  .operation-data {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 16px;
  }
  .data-box {
    display: flex;
    flex-direction: column;
    flex: 1;
    padding: 16px;
    background: #fff6de;
    border-radius: 4px;
    min-width: calc(25% - 12px);
    max-width: calc(25% - 8px);
  }

  .data-box-title {
    display: flex;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .data-box-title h3 {
    font-size: 16px;
    font-weight: 400;
    color: var(--el-color-black);
  }

  .data-box h4 {
    font-size: 20px;
    font-weight: 700;
    margin-top: auto;
    color: var(--el-color-black);
  }

  .data-box h5 {
    color: var(--el-color-tips);
    font-size: 12px;
    font-style: normal;
    font-weight: 400;
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

  .product-bottom {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 16px;
  }

  .common-form {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    margin-bottom: 16px;
    .radio-search {
      :deep(.el-radio-button) {
        margin-right: -3px;
        margin-bottom: 0;

        .el-radio-button__inner {
          padding: 8px 11px !important;
          font-size: 12px;
        }
      }
    }
  }
  .product-bottom {
    .flex-1 {
      flex: 1 1 auto;
      flex-shrink: 1;
      min-width: calc(50% - 8px);
    }
  }
  .product-name {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .data-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 300px;
    padding: 16px;
    background: #fff6de;
    flex: 1 1 auto;
    width: 100%;
    border-radius: 8px;
    font-size: 14px;
    color: var(--el-color-black);
    .item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 8px;
      overflow: hidden;
      .num {
        display: block;
        width: 20px;
        height: 20px;
        display: flex;
        justify-content: center;
        align-items: center;
        flex-shrink: 1;
        border-radius: 50%;
        font-size: 12px;
      }
      .key-box {
        background: #ffbe00;
      }
      .key-box2 {
        background: #fff;
      }
      .p-relative {
        display: block;
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        position: relative;
      }
      .p-relative:nth-child(2) {
        justify-content: flex-end;
      }
      .p-relative-l {
        display: flex;
        flex: 1;
        align-items: center;
        gap: 8px;
        overflow: hidden;
        .l-autoTips {
          display: block;
          flex: 1;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }
    }
  }

  .pagination {
    display: flex;
    justify-content: flex-end;
    width: 100%;
    margin-top: 12px;
  }
</style>

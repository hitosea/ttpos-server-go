<template>
  <div class="p-4 bg-white rounded min-h-full">
    <div class="flex justify-between items-center mb-4">
      <el-form :inline="true">
        <el-form-item :label="$t('站点编码')">
          <el-input v-model="siteCode" :placeholder="$t('留空查询全部')" clearable style="min-width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="ti-search-btn" :loading="loading" @click="fetchStats">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 按 Topic 网格展示 -->
    <div class="topic-grid">
      <div v-for="topic in topics" :key="topic.key" class="topic-card">
        <div class="topic-card-header">
          <span class="font-medium">{{ topic.label }}</span>
          <span class="text-xs text-gray-400 ml-1">{{ topic.topicName }}</span>
        </div>
        <div class="topic-card-body">
          <div class="stat-row">
            <div class="stat-item">
              <div class="stat-value" :class="topic.color">{{ topic.count }}</div>
              <div class="stat-label">{{ $t('失败数') }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-value text-blue-500">{{ topic.pendingCount }}</div>
              <div class="stat-label">{{ $t('进行中') }}</div>
            </div>
          </div>
        </div>
        <div class="topic-card-footer" v-permission="['admin_erpnext_siRetry']">
          <el-button
            type="danger"
            size="small"
            :loading="processing"
            :disabled="topic.count === 0"
            @click="handleAction(topic, 'retry')"
          >
            {{ $t('重试') }} ({{ topic.count }})
          </el-button>
          <el-button
            type="info"
            size="small"
            :loading="processing"
            :disabled="topic.count === 0"
            @click="handleAction(topic, 'skip')"
          >
            {{ $t('跳过') }} ({{ topic.count }})
          </el-button>
        </div>
      </div>
    </div>

    <!-- 汇总 -->
    <div class="border-t pt-4 mt-4 flex justify-between items-center">
      <div class="flex gap-8">
        <div class="text-gray-500">{{ $t('总失败数') }}: <span class="text-lg font-semibold text-gray-700">{{ totalFailed }}</span></div>
        <div class="text-gray-500">{{ $t('总进行中') }}: <span class="text-lg font-semibold text-blue-600">{{ totalPending }}</span></div>
      </div>
      <div class="flex gap-2" v-permission="['admin_erpnext_siRetry']">
        <el-button
          type="danger"
          :loading="processing"
          :disabled="totalFailed === 0"
          @click="handleAllAction('retry')"
        >
          {{ $t('全部重试') }} ({{ totalFailed }})
        </el-button>
        <el-button
          type="info"
          :loading="processing"
          :disabled="totalFailed === 0"
          @click="handleAllAction('skip')"
        >
          {{ $t('全部跳过') }} ({{ totalFailed }})
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSIFailedStats, retrySI } from '@/api/erp-dlq';
import type { FailedSIStats } from '@/api/erp-dlq';
import { ElMessageBox } from 'element-plus';
import { message } from '@/utils/feedback';
import { $t } from '@/i18n';

type Action = 'retry' | 'skip';

const siteCode = ref('');
const loading = ref(false);
const processing = ref(false);
const stats = ref<FailedSIStats>({
  save_failed_count: 0,
  cancel_failed_count: 0,
  return_failed_count: 0,
  total_failed_count: 0,
  save_pending_count: 0,
  cancel_pending_count: 0,
  return_pending_count: 0,
  total_pending_count: 0,
});

interface TopicItem {
  key: string;
  label: string;
  topicName: string;
  msgType: string;
  color: string;
  count: number;
  pendingCount: number;
}

const topics = computed<TopicItem[]>(() => [
  {
    key: 'save',
    label: $t('Save Sales Invoice'),
    topicName: 'save-sales-invoice',
    msgType: 'save-sales-invoice',
    color: 'text-orange-500',
    count: stats.value.save_failed_count,
    pendingCount: stats.value.save_pending_count,
  },
  {
    key: 'cancel',
    label: $t('Cancel Sales Invoice'),
    topicName: 'cancel-sales-invoice',
    msgType: 'cancel-sales-invoice',
    color: 'text-red-500',
    count: stats.value.cancel_failed_count,
    pendingCount: stats.value.cancel_pending_count,
  },
  {
    key: 'return',
    label: $t('Return Sales Invoice'),
    topicName: 'return-sales-invoice',
    msgType: 'return-sales-invoice',
    color: 'text-purple-500',
    count: stats.value.return_failed_count,
    pendingCount: stats.value.return_pending_count,
  },
]);

const totalFailed = computed(() => topics.value.reduce((sum, t) => sum + t.count, 0));
const totalPending = computed(() => topics.value.reduce((sum, t) => sum + t.pendingCount, 0));

const actionLabels: Record<Action, { verb: string; confirmTitle: string }> = {
  retry: { verb: '重试', confirmTitle: '确认重试' },
  skip: { verb: '跳过', confirmTitle: '确认跳过' },
};

const fetchStats = async () => {
  try {
    loading.value = true;
    const params = siteCode.value ? { site_code: siteCode.value } : undefined;
    const res = await getSIFailedStats(params);
    if (res.data) {
      stats.value = res.data;
    }
  } finally {
    loading.value = false;
  }
};

const handleAction = async (topic: TopicItem, action: Action) => {
  if (topic.count === 0) return;
  const labels = actionLabels[action];
  try {
    await ElMessageBox.confirm(
      `确认${labels.verb} ${topic.topicName} 的 ${topic.count} 条失败记录？`,
      labels.confirmTitle,
      { type: 'warning' },
    );
  } catch {
    return;
  }

  try {
    processing.value = true;
    const res = await retrySI({
      msg_type: topic.msgType,
      site_code: siteCode.value || undefined,
      action,
    });
    message.success(`已${labels.verb} ${res.data?.retry_count || 0} 条记录`);
    await fetchStats();
  } finally {
    processing.value = false;
  }
};

const handleAllAction = async (action: Action) => {
  const labels = actionLabels[action];
  try {
    await ElMessageBox.confirm(
      `确认${labels.verb}全部 ${totalFailed.value} 条失败记录？`,
      labels.confirmTitle,
      { type: 'warning' },
    );
  } catch {
    return;
  }

  try {
    processing.value = true;
    const res = await retrySI({
      site_code: siteCode.value || undefined,
      action,
    });
    message.success(`已${labels.verb} ${res.data?.retry_count || 0} 条记录`);
    await fetchStats();
  } finally {
    processing.value = false;
  }
};

onMounted(() => {
  fetchStats();
});
</script>

<style lang="scss" scoped>
.topic-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.topic-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
}

.topic-card-header {
  border-bottom: 1px solid #f0f0f0;
  padding-bottom: 8px;
  margin-bottom: 12px;
}

.topic-card-body {
  flex: 1;
  padding: 8px 0;
}

.stat-row {
  display: flex;
  justify-content: space-around;
  gap: 16px;
}

.stat-item {
  flex: 1;
  text-align: center;
}

.topic-card-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}

.stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

.stat-value {
  font-size: 32px;
  font-weight: 600;
}
</style>

<template>
  <div class="right-content pr">
    <div class="right-content-top" v-if="business_status == 1">
      {{ $t('当前门店营业状态为：“测试营业”，若确认正式营业请到“商家档案”板块进行开启。') }}
    </div>
    <!--内容区域-->
    <div class="right-content-box">
      <div class="subject-wrap">
        <div :class="route.path.includes('/home') ? 'home-div' : 'main-div'">
          <ChildTabs></ChildTabs>
          <router-view @selectMenu="selectMenu" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import ChildTabs from '@/views/layout/childTabs.vue';
  import { useRoute } from 'vue-router';
  import { useUserStore } from '@/store';

  const route = useRoute();
  const emit = defineEmits(['selectMenu']);
  const { computedSettings } = useUserStore();
  const settings = computedSettings().settings;
  const business_status = settings?.value?.business_status || 1;

  const selectMenu = (e) => {
    emit('selectMenu', e);
  };
</script>
<style scoped lang="scss">
  .right-content-top {
    display: flex;
    padding: 12px 16px;
    justify-content: center;
    align-items: center;
    gap: 24px;
    align-self: stretch;
    background: #ffeee0;
    font-size: 14px;
    color: #ff6600;
  }
</style>

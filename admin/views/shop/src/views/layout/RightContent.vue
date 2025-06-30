<template>
  <div class="right-content pr">
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
  const route = useRoute();
  const emit = defineEmits(['selectMenu']);

  const selectMenu = (e) => {
    emit('selectMenu', e);
  };
</script>

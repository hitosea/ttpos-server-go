<template>
  <div :class="hasChild != null ? 'main' : 'main right-big'">
    <!--头部-->

    <Head></Head>
    <!--left menu-->

    <LeftMenu ref="leftMenuRef" @selectMenu="selectMenuFunc"></LeftMenu>

    <!--right content-->
    <RightContent @selectMenu="selectMenu"></RightContent>
  </div>
</template>

<script setup>
  // 使用 Vue3 <script setup> 语法改写
  // 引入依赖组件
  import { ref } from 'vue';
  import LeftMenu from '@/views/layout/LeftMenu.vue';
  import RightContent from '@/views/layout/RightContent.vue';
  import Head from '@/views/layout/Head.vue';

  // 左侧菜单组件引用
  const leftMenuRef = ref(null);

  // 是否有子菜单
  const hasChild = ref(null);


  // 左边子组件传来的参数：设置是否存在子菜单
  const selectMenuFunc = (param) => {
    hasChild.value = param;
  };

  // 右侧内容触发菜单选择，转发到左侧菜单组件的方法
  const selectMenu = (data) => {
    leftMenuRef.value?.choseMenu(data.type, data.item, data.index, data.query);
  };
</script>

<template>
  <div>
    <div class="">
      <div class="Echarts">
        <div id="TransactionChart"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted, onBeforeUnmount } from 'vue';
  import { ElMessage } from 'element-plus';
  import * as echarts from 'echarts';
  import StatisticsApi from '@/api/statistics.js';
  import { formatDate } from '@/utils/DateTime.js';

  // 图表实例
  let myChart = null;

  // 初始化日期
  const endDate = new Date();
  const startDate = new Date();
  startDate.setTime(startDate.getTime() - 3600 * 1000 * 24 * 7);

  // 响应式数据
  const loading = ref(true);
  const activeName = ref('order');
  const datePicker = ref([formatDate(startDate, 'yyyy-MM-dd'), formatDate(endDate, 'yyyy-MM-dd')]);
  const dataList = ref(null);

  // 图表配置
  const option = ref({
    title: {
      text: $t('近七日数据'),
      textStyle: {
        fontWeight: 'bold', // 设置文字粗细为粗体
        color: '#100A05',
        fontSize: '20px',
      },
    },
    grid: {
      left: '40px',
      right: '40px',
      bottom: '0px',
      top: '60px',
      containLabel: true,
    },
    tooltip: {
      trigger: 'axis',
    },
    yAxis: {
      type: 'value',
    },
  });

  // 切换
  const handleClick = (e) => {
    getData();
  };

  // 选择时间
  const changeDate = () => {
    getData();
  };

  // 处理窗口大小变化
  const handleResize = () => {
    if (myChart) {
      myChart.resize();
    }
  };

  // 创建图表对象
  const myEcharts = () => {
    // 基于准备好的dom，初始化echarts实例
    const chartDom = document.getElementById('TransactionChart');
    if (!chartDom) return;

    // Dispose existing chart if it exists
    if (myChart) {
      myChart.dispose();
    }

    myChart = echarts.init(chartDom, null, {
      renderer: 'canvas',
      useDirtyRect: false,
    });

    // Set wheel event to be passive
    myChart.getZr().on('mousewheel', function () {}, { passive: true });

    // 获取列表
    getData();
  };

  // 格式化数据
  const createOption = () => {
    if (!loading.value && myChart) {
      let names = [];
      let xAxis = dataList.value.days;
      let series1 = [];
      let series2 = [];
      try {
        (dataList.value.data || []).forEach((item) => {
          series1.push(item.total_money);
          series2.push(item.total_num);
        });
      } catch (error) {
        console.log(error);
      }
      if (activeName.value == 'order') {
        names = [$t('实收'), $t('订单数')];
      } else if (activeName.value == 'refund') {
        names = ['退单额', '退单量'];
      }

      // 指定图表的配置项和数据
      option.value.xAxis = {
        type: 'category',
        boundaryGap: false,
        data: xAxis,
      };
      option.value.color = ['#FFBE00', '#FE852E'];

      option.value.legend = {
        data: [
          {
            name: names[0],
            color: '#ccc',
          },
          {
            name: names[1],
          },
        ],
        right: 20,
      };
      option.value.series = [
        {
          name: names[0],
          type: 'line',
          data: series1,
          lineStyle: {
            color: '#FFBE00',
          },
        },
        {
          name: names[1],
          type: 'line',
          data: series2,
          lineStyle: {
            color: '#FE852E',
          },
        },
      ];

      try {
        myChart.setOption(option.value);
        myChart.resize();
      } catch (error) {
        console.error('设置图表选项失败:', error);
      }
    }
  };

  // 获取数据
  const getData = async () => {
    loading.value = true;
    try {
      const res = await StatisticsApi.getOrderByDate(
        {
          search_time: datePicker.value,
          type: activeName.value,
        },
        true
      );
      if (res.code != 1) return;
      dataList.value = res.data;
      loading.value = false;
      createOption();
    } catch (error) {
      ElMessage.error(error?.message || $t('获取数据失败'));
      loading.value = false;
    }
  };

  // 组件挂载时初始化图表
  onMounted(() => {
    myEcharts();
    window.addEventListener('resize', handleResize, { passive: true });
  });

  // 组件卸载前清理资源
  onBeforeUnmount(() => {
    // Clean up the chart instance when component is unmounted
    if (myChart) {
      myChart.dispose();
      myChart = null;
    }
    window.removeEventListener('resize', handleResize, { passive: true });
  });
</script>

<style scoped>
  .Echarts {
    box-sizing: border-box;
  }

  .Echarts > div {
    width: 100%;
    height: 400px;
    box-sizing: border-box;
  }
</style>

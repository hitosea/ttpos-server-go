<template>
    <div class="echarts-box">
        <div id="TransactionChart">
        </div>
    </div>
</template>
<script>
import * as echarts from "echarts";
let myChart = null;
export default {
    props: {
        chart_list: '',
    },
    watch: {
        'chart_list': {
            handler(val) {

            },
            deep: true,
            immediate: true,
        },
    },
    data() {
        return {
            xdata: [],
            sData: [],
            option: {},
        }
    },
    methods: {
        /*创建图表对象*/
        myEcharts() {
            // 基于准备好的dom，初始化echarts实例
            myChart = echarts.init(document.getElementById('TransactionChart'));
            myChart.setOption(this.option);
            myChart.resize();
        },
    },
    mounted() {
        (this.chart_list || []).map(item => {
            this.xdata.push(item.name);
            this.sData.push({
                value: item.damage_count,
                ratio: item.damage_ratio,
            })
        })
        this.option = {
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'shadow'
                },
                formatter: function (params) {
                    return [
                        params[0]?.name || '-', // 显示柱的值
                        $t('损耗数量：') + (params[0]?.data?.value || 0), // 显示柱的值
                        $t('损耗比例：') + (params[0]?.data?.ratio || 0), // 显示柱的类别
                    ].join('<br>');
                }
            },
            grid: {
                top: '20px',
                left: '16px',
                right: '16px',
                bottom: '0px',
                containLabel: true
            },
            xAxis: [
                {
                    type: 'category',
                    data: this.xdata,
                    axisTick: {
                        alignWithLabel: true
                    }
                }
            ],
            yAxis: [
                {
                    type: 'value',
                }
            ],
            dataZoom: [
                {
                    type: 'slider', // 滑动条类型 dataZoom
                    show: true,
                    startValue: 0,
                    endValue: 10 // 默认显示10个柱子，可通过滑动改变显示范围
                },
                {
                    type: 'inside' // 鼠标滚轮控制 dataZoom
                }
            ],
            series: [
                {
                    type: 'bar',
                    data: this.sData,
                }
            ]
        }
        this.myEcharts();
    },
}
</script>
<style lang="scss" scoped>
.echarts-box {
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    bottom: 0;

    #TransactionChart {
        width: 100%;
        height: 100%;
    }
}
</style>
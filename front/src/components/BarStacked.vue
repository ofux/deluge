<template>
  <div :id="id"></div>
</template>

<script>
  import Highcharts from 'highcharts';

  export default {
    watch: {
      data(data) {
        Highcharts.chart(this.id, {
          chart: {
            type: 'bar',
          },
          title: {
            text: this.xLabel,
          },
          xAxis: {
            categories: ['Ok', 'Ko'],
          },
          yAxis: {
            min: 0,
            title: {
              text: this.yLabel,
            },
          },
          legend: {
            reversed: true,
          },
          plotOptions: {
            series: {
              stacking: 'normal',
            },
          },
          series: data,
        });
      },
    },
    data() {
      return {
        id: `chart-${this._uid}`,
      };
    },
    props: ['data', 'xLabel', 'yLabel'],
  };
</script>

<style scoped>

</style>

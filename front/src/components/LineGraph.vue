<template>
  <div :id="id"></div>
</template>

<script>
  import Highcharts from 'highcharts';

  export default {
    watch: {
      data(data) {
        Highcharts.chart(this.id, {
          title: {
            text: this.title,
          },

          xAxis: {
            title: {
              text: this.xLabel,
            },
          },

          yAxis: {
            title: {
              text: this.yLabel,
            },
          },

          legend: {
            layout: 'vertical',
            align: 'right',
            verticalAlign: 'middle',
          },

          plotOptions: {
            series: {
              pointStart: 0,
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
    props: ['data', 'title', 'xLabel', 'yLabel'],
  };
</script>

<style scoped>

</style>

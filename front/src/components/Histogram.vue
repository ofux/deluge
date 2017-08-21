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
            type: 'column',
          },
          title: {
            text: this.xLabel,
          },
          yAxis: {
            min: 0,
            title: {
              text: this.yLabel,
            },
          },
          plotOptions: {
            column: {
              pointPadding: 0.2,
              borderWidth: 0,
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

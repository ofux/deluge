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
            type: 'area',
          },
          title: {
            text: this.xLabel,
          },
          yAxis: {
            title: {
              text: this.yLabel,
            },
          },
          plotOptions: {
            area: {
              pointStart: 0,
              marker: {
                enabled: false,
                symbol: 'circle',
                radius: 2,
                states: {
                  hover: {
                    enabled: true,
                  },
                },
              },
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

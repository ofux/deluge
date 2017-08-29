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
            type: 'scatter',
            zoomType: 'xy',
          },
          title: {
            text: this.title,
          },
          xAxis: {
            title: {
              enabled: true,
              text: this.xLabel,
            },
            startOnTick: true,
            endOnTick: true,
            showLastLabel: true,
          },
          yAxis: {
            title: {
              text: this.yLabel,
            },
          },
          legend: {
            layout: 'vertical',
            align: 'left',
            verticalAlign: 'top',
            x: 100,
            y: 70,
            floating: true,
            backgroundColor: (Highcharts.theme && Highcharts.theme.legendBackgroundColor) || '#FFFFFF',
            borderWidth: 1,
          },
          plotOptions: {
            scatter: {
              marker: {
                radius: 5,
                states: {
                  hover: {
                    enabled: true,
                    lineColor: 'rgb(100,100,100)',
                  },
                },
              },
              states: {
                hover: {
                  marker: {
                    enabled: false,
                  },
                },
              },
              tooltip: {
                headerFormat: '<b>{series.name}</b><br>',
                pointFormat: '{point.x}, {point.y}',
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
    props: ['data', 'title', 'xLabel', 'yLabel'],
  };
</script>

<style scoped>

</style>

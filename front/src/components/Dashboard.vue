<template>
  <div>
    <h1>Welcome to Deluge !</h1>
    <div class="data">
      Output {{name}} with {{status}} status
    </div>
    <div class="graph">
      <line-graph :data="line.data" :title="line.title" :x-label="line.xLabel" :y-label="line.yLabel"></line-graph>
    </div>
    <div class="graph">
      <area-graph :data="area.data" :x-label="area.xLabel" :y-label="area.yLabel"></area-graph>
    </div>
    <div class="graph">
      <histogram :data="histo.data" :x-label="histo.xLabel" :y-label="histo.yLabel"></histogram>
    </div>
    <div class="graph">
      <bar-stacked :data="barStacked.data" :x-label="barStacked.xLabel" :y-label="barStacked.yLabel"></bar-stacked>
    </div>
    <div class="graph">
      <scatter-plot :data="scatter.data" :title="scatter.title" :x-label="scatter.xLabel" :y-label="scatter.yLabel"></scatter-plot>
    </div>
  </div>
</template>

<script>
  import AreaGraph from '@/components/AreaGraph';
  import BarStacked from '@/components/BarStacked';
  import Histogram from '@/components/Histogram';
  import LineGraph from '@/components/LineGraph';
  import ScatterPlot from '@/components/ScatterPlot';

  export default {
    components: {
      AreaGraph,
      BarStacked,
      Histogram,
      LineGraph,
      ScatterPlot,
    },
    data: function data() {
      this.$http.get('http://localhost:8000/ex1_output.json').then((response) => {
        const scenarioIterations = response.data.Scenarios.sc1.Report.Stats.PerIteration;
        const scenarioGlobal = response.data.Scenarios.sc1.Report.Stats.Global;
        const requestsName = Object.keys(scenarioGlobal.PerRequests);
        const colors = ['rgba(223, 83, 83, .5)', 'rgba(119, 152, 191, .5)'];

        this.name = response.data.Name;

        const lines = [
          { name: 'Mean time', data: [] },
          { name: 'Max time', data: [] },
          { name: 'Min time', data: [] },
        ];

        this.line.data = scenarioIterations.reduce((old, iteration) => {
          old[0].data.push(iteration.Global.MeanTime);
          old[1].data.push(iteration.Global.MaxTime);
          old[2].data.push(iteration.Global.MinTime);
          return old;
        }, lines);

        this.area.data = scenarioIterations.reduce((old, iteration) => {
          const quantiles = iteration.Global.ValueAtQuantiles;
          if (old.length === 0) {
            Object.keys(quantiles).forEach(key => old.push({ name: key, data: [] }));
          } else {
            Object.keys(quantiles).forEach((key, index) => old[index].data.push(quantiles[key]));
          }
          return old;
        }, []).reverse();

        this.histo.data = scenarioIterations.reduce((old, iteration) => {
          if (iteration.PerOkKo.Ok) {
            old[0].data.push(iteration.PerOkKo.Ok.CallCount);
          }

          if (iteration.PerOkKo.Ko) {
            old[1].data.push(iteration.PerOkKo.Ko.CallCount);
          }

          return old;
        }, [{ name: 'Ok', data: [] }, { name: 'Ko', data: [] }]);

        this.barStacked.data = requestsName.reduce((acc, requestName) => {
          const requestOk = scenarioGlobal.PerRequests[requestName].PerOkKo.Ok;
          const requestKo = scenarioGlobal.PerRequests[requestName].PerOkKo.Ko;

          acc.push({
            name: requestName,
            data: [requestOk ? requestOk.CallCount : 0, requestKo ? requestKo.CallCount : 0],
          });
          return acc;
        }, []);

        this.scatter.data = requestsName.reduce((acc, requestName, i) => {
          acc.push({
            name: requestName,
            color: colors[i],
            data: scenarioIterations
              .map((iter, index) => [index, iter.PerRequests[requestName].Global.MeanTime]),
          });

          return acc;
        }, []);
      });

      return {
        name: 'loading',
        status: 'loading',
        area: {
          data: [],
          xLabel: 'Value at quantile over iterations',
          ylabel: 'Value at quantile',
        },
        barStacked: {
          data: [],
          xLabel: 'Nbr of request',
          ylabel: 'Name of request',
        },
        histo: {
          data: [],
          xLabel: 'Success & Fail requests over iterations',
          ylabel: 'Nbr of request',
        },
        line: {
          data: [],
          title: 'Global Time of requests over iterations',
          xLabel: 'N° of iteration',
          yLabel: 'Time',
        },
        scatter: {
          data: [],
          title: 'Mean time over iterations',
          xLabel: 'N° of iteration',
          yLabel: 'Mean time',
        },
      };
    },
  };
</script>

<style scoped>
  ul {
    list-style-type: none;
    padding: 0;
  }

  li {
    display: inline-block;
    margin: 0 10px;
  }

  a {
    color: #42b983;
  }

  .data {
    margin-bottom: 30px;
  }

  .graph {
    margin: 20px 0;
  }
</style>

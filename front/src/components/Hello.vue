<template>
  <div>
    <h1>Welcome to Deluge !</h1>
    <div class="data">
      Output {{outputData.name}} with {{outputData.status}} status
    </div>
    <div class="graph">
      <line-graph :data="outputData.line.data" :title="outputData.line.title" :x-label="outputData.line.xLabel" :y-label="outputData.line.yLabel"></line-graph>
    </div>
    <div class="graph">
      <area-graph :data="outputData.area.data" :x-label="outputData.area.xLabel" :y-label="outputData.area.yLabel"></area-graph>
    </div>
    <div class="graph">
        <histogram :data="outputData.histo.data" :x-label="outputData.histo.xLabel" :y-label="outputData.histo.yLabel"></histogram>
    </div>
    <div class="graph">
      <bar-stacked :data="outputData.barStacked.data" :x-label="outputData.barStacked.xLabel" :y-label="outputData.barStacked.yLabel"></bar-stacked>
    </div>
  </div>
</template>

<script>
  import AreaGraph from '@/components/AreaGraph';
  import BarStacked from '@/components/BarStacked';
  import Histogram from '@/components/Histogram';
  import LineGraph from '@/components/LineGraph';

  export default {
    components: {
      AreaGraph,
      BarStacked,
      Histogram,
      LineGraph,
    },
    asyncComputed: {
      outputData: {
        get() {
          return this.$http.get('http://localhost:8000/ex1_output.json').then((response) => {
            const scenarioIterations = response.data.Scenarios.sc1.Report.Stats.PerIteration;
            const scenarioGlobal = response.data.Scenarios.sc1.Report.Stats.Global;

            const dataForHighcharts = scenarioIterations.reduce((old, iteration) => {
              old[0].push(iteration.Global.MeanTime);
              old[1].push(iteration.Global.MaxTime);
              old[2].push(iteration.Global.MinTime);
              return old;
            }, [[], [], []]);

            const dataForArea = scenarioIterations.reduce((old, iteration) => {
              const quantiles = iteration.Global.ValueAtQuantiles;
              if (old.length === 0) {
                Object.keys(quantiles).forEach(key => old.push({ name: key, data: [] }));
              } else {
                Object.keys(quantiles).forEach((key, index) => old[index].data.push(quantiles[key]));
              }
              return old;
            }, []);

            const dataHisto = scenarioIterations.reduce((old, iteration) => {
              if (iteration.PerOkKo.Ok) {
                old[0].data.push(iteration.PerOkKo.Ok.CallCount);
              }

              if (iteration.PerOkKo.Ko) {
                old[1].data.push(iteration.PerOkKo.Ko.CallCount);
              }

              return old;
            }, [{ name: 'Ok', data: [] }, { name: 'Ko', data: [] }]);

            const dataBarStacked = Object.keys(scenarioGlobal.PerRequests)
              .reduce((acc, requestName) => {
                const requestOk = scenarioGlobal.PerRequests[requestName].PerOkKo.Ok;
                const requestKo = scenarioGlobal.PerRequests[requestName].PerOkKo.Ko;

                acc.push({
                  name: requestName,
                  data: [requestOk ? requestOk.CallCount : 0, requestKo ? requestKo.CallCount : 0],
                });
                return acc;
              }, []);

            return {
              name: response.data.Name,
              status: response.data.Status,
              area: {
                data: dataForArea.reverse(),
              },
              barStacked: {
                data: dataBarStacked,
              },
              histo: {
                data: dataHisto,
              },
              line: {
                data: [{
                  name: 'Mean time',
                  data: dataForHighcharts[0],
                }, {
                  name: 'Max time',
                  data: dataForHighcharts[1],
                }, {
                  name: 'Min time',
                  data: dataForHighcharts[2],
                }],
              },
            };
          });
        },

        default: {
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
        },
      },
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

  .graph {
    margin: 0px 0;
  }
</style>

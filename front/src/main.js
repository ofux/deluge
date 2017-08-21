import Vue from 'vue';
import AsyncComputed from 'vue-async-computed';
import VueResource from 'vue-resource';

import App from './App';
import router from './router';

Vue.use(VueResource);
Vue.use(AsyncComputed);
Vue.config.productionTip = false;

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  template: '<App/>',
  components: { App },
});

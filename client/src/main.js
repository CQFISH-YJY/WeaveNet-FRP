import { createApp } from 'vue';
import { createRouter, createWebHashHistory } from 'vue-router';
import App from './App.vue';
import Home from './views/Home.vue';
import Tunnels from './views/Tunnels.vue';
import Logs from './views/Logs.vue';
import Settings from './views/Settings.vue';
import './assets/global.css';

const routes = [
  { path: '/', name: 'home', component: Home, meta: { title: '首页' } },
  { path: '/tunnels', name: 'tunnels', component: Tunnels, meta: { title: '隧道' } },
  { path: '/logs', name: 'logs', component: Logs, meta: { title: '日志' } },
  { path: '/settings', name: 'settings', component: Settings, meta: { title: '设置' } },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

const app = createApp(App);
app.use(router);
app.mount('#app');

// FILE: frontend/src/entries/admins.js
import { createApp } from 'vue';
import antd from 'ant-design-vue';
import 'ant-design-vue/dist/reset.css';
import AdminsPage from '@/pages/admins/AdminsPage.vue';
import { readyI18n } from '@/i18n';
 
readyI18n().then((i18n) => {
  createApp(AdminsPage).use(antd).use(i18n).mount('#app');
});
 

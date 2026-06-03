<script setup>
  import {inject} from 'vue'
  import Theme from "@/frameworks/Theme.vue";
  import Cookies from "js-cookie";
  import {useRoute, useRouter} from "vue-router";
  import {Logout} from "@/js/util.js";

  const title = inject("title");
  const sidebarStatus = inject('sidebarStatus')
  const login = inject('login')
  const emitter = inject('emitter')
  const router = useRouter();
  const route = useRoute();
  const page = inject('page')
  const maxPage = inject('maxPage')

  function SidebarStatusChanged() {
    sidebarStatus.value = !sidebarStatus.value;
  }
  async function logout() {
    if (await Logout(Cookies.get("token"))) {
      Cookies.remove("token");
      router.push("/dash/login");
    }
  }
</script>

<template>
  <header class="dashboard-header">
    <mdui-button-icon class="header-icon" icon="menu"
      @click="SidebarStatusChanged()"></mdui-button-icon>
    <div class="header-title">{{ title }}</div>
    <div class="header-spacer"></div>
<!--    1 for refresh, 2 for backwards, 3 for forwards-->
    <mdui-segmented-button-group class="pager-control" v-if="login && route.path === '/dash/task'">
      <mdui-segmented-button @click="(emitter=2)">←</mdui-segmented-button>
      <mdui-segmented-button>{{ page }} / {{ maxPage }}</mdui-segmented-button>
      <mdui-segmented-button @click="(emitter=3)">→</mdui-segmented-button>
    </mdui-segmented-button-group>
    <mdui-button-icon @click="(emitter=1)" v-if="login && route.path === '/dash/task'" icon="refresh"></mdui-button-icon>

    <Theme v-if="route.path !== '/dash/task'"></Theme>
    <mdui-button-icon class="header-icon" @click="logout()" v-if="login"
      icon="exit_to_app"></mdui-button-icon>
  </header>
</template>

<style scoped>
.dashboard-header {
  align-items: center;
  background: rgb(var(--mdui-color-surface));
  border-bottom: 1px solid rgba(148, 163, 184, 0.28);
  box-sizing: border-box;
  box-shadow: 0 1px 8px rgba(15, 23, 42, 0.18);
  display: flex;
  gap: 0.5rem;
  height: 64px;
  left: 0;
  padding: 0 0.5rem;
  position: fixed;
  right: 0;
  top: 0;
  z-index: 2000;
}

.header-icon {
  flex: 0 0 auto;
}

.header-title {
  color: rgb(var(--mdui-color-on-surface));
  font-size: 1.375rem;
  font-weight: 500;
  line-height: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header-spacer {
  flex: 1 1 auto;
}

@media (max-width: 620px) {
  .pager-control {
    max-width: 10rem;
  }
}

</style>

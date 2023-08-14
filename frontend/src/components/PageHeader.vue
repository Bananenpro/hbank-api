<template>
  <div class="header">
    <div class="left">
      <img v-if="canGoBack" class="bars clickable" @click="goBack" src="@/assets/back_arrow.svg"/>
      <router-link v-else to="/"><img class="bars clickable" src="@/assets/bars.svg"/></router-link>
      <h3 class="branding clickable"><router-link to="/">H-Bank</router-link></h3>
      <h3 v-if="title" class="title"> | {{title}}</h3>
    </div>
    <div class="right">
      <a v-if="!authenticated" class="clickable auth-btn login-btn" :href="loginURL">{{$t("login")}}</a>
      <HamburgerMenu v-if="authenticated"/>
    </div>
  </div>
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import {authenticated} from '@/api'
import {pageTitle} from '@/router'
import HamburgerMenu from '@/components/HamburgerMenu.vue'
import api from '@/api'
export default defineComponent({
  name: "PageHeader",
  components: {
    HamburgerMenu
  },
  data() {
    return {
      authenticated: authenticated,
      title: pageTitle,
      loginURL: api.defaults.baseURL+"auth/login?redirect="+encodeURIComponent(window.location.pathname),
    }
  },
  computed: {
    canGoBack() {
      return this.$route.path !== "/" && this.$route.path !== "/dashboard"
    }
  },
  methods: {
    goBack() {
      window.history.back()
    }
  }
})
</script>

<style scoped>
.header {
  background-color: var(--header-bg-color);
  color: var(--header-fg-color);
  display: flex;
  justify-content: space-between;
  height: 50px;
}

.left {
  display: flex;
  margin-left: 7px;
  width: 85%;
}

.right {
  display: flex;
  margin-right: 7px;
}

.branding {
  margin-left: 10px;
  font-size: 20px;
  margin-top: auto;
  margin-bottom: auto;
  text-decoration: none;
  color: var(--header-fg-color);
  min-width: 68px;
}

.branding > a {
  text-decoration: none;
  color: var(--header-fg-color);
}

.title {
  margin-left: 7px;
  font-size: 20px;
  font-weight: normal;
  margin-top: auto;
  margin-bottom: auto;
  text-decoration: none;
  color: var(--header-fg-color);
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bars {
  margin-top: 6px;
  margin-bottom: 6px;
  width: 28px;
}

.auth-btn {
  text-decoration: none;
  color: var(--fg-color-header);
  font-size: 18px;
  display: block;
  margin-top: auto;
  margin-bottom: auto;
}

.login-btn {
  padding: 6px 2vw;
  border: 1px solid var(--border-color);
  border-radius: 3px;
}

@media screen and (min-width: 800px) {
  .left {
    margin-left: 15px;
  }

  .right {
    margin-right: 15px;
  }
}

@media screen and (max-width: 350px) {
  .left {
    margin-left: 5px;
  }

  .right {
    margin-right: 5px;
  }

  .title {
    margin-left: 7px;
  }

  .login-btn {
    margin-right: 5px;
  }
}
</style>

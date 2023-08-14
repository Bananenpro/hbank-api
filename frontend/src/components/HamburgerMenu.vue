<template>
  <img @click="show = !show" class="icon clickable" :src="darkTheme ? require('@/assets/hamburger-light.svg') : require('@/assets/hamburger-dark.svg')"/>
  <transition name="fade">
  <div @click="show = false" class="background" v-show="show"></div>
  </transition>
  <div class="dropdown" :class="show ? 'dropdown-shown' : 'dropdown-hidden'">
    <div class="separator" id="top-separator"></div>
    <a v-bind:href="idProvider+'/user/profile'" class="dropdown-item clickable">{{$t("hamburger.account")}}</a>
    <p class="dropdown-item clickable" id="settings-btn" @click="$router.push('/settings')">{{$t("hamburger.settings")}}</p>
    <div class="separator"></div>
    <p class="dropdown-item clickable" id="logout-btn" @click="signOut()">{{$t("hamburger.logout")}}</p>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import tc from 'tinycolor2'
import {logout, config} from '@/api'

export default defineComponent({
  name: "HambugerMenu",
  data() {
    return {
      show: false,
      darkTheme: false,
      idProvider: "",
    }
  },
  async mounted() {
    setTimeout(() => {
      const bgColor = getComputedStyle(document.documentElement).getPropertyValue('--bg-color');

      const color = tc(bgColor);

      this.darkTheme = color.isDark();
    }, 100)
    this.idProvider = (await config()).idProvider
  },
  methods: {
    async signOut() {
      await logout()
      this.$router.push('/')
    }
  },
  watch:{
    $route (){
      this.show = false
    }
  } 
})
</script>


<style scoped>
.icon {
  height: 22px;
  margin-top: 14px;
  margin-bottom: 14px;
}
.background {
  position: absolute;
  top: 50px;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--dialog-bg-color);
  z-index: 99;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.15s ease-out;
}

.fade-enter, .fade-leave-to {
  opacity: 0;
}

.dropdown {
  background: var(--bg-color);
  position: absolute;
  top: 50px;
  left: 0;
  right: 0;
  transition-property: transform, opacity;
  transition-duration: 0.15s;
  transition-timing-function: ease-in-out;
  transform-origin: top;
  overflow: hidden;
  z-index: 100;
  opacity: 0;
}
.dropdown-shown {
  transform: scaleY(1);
  opacity: 1;
}
.dropdown-hidden {
  transform: scaleY(0);
  opacity: 0;
}
.separator {
  margin: 0px 7px;
}
.dropdown-item {
  font-size: 22px;
  line-height: 30px;
  margin: 0;
  padding-left: 7px;
  padding-top: 10px;
  color: var(--fg-color);
  text-decoration: none;
  display: block;
}

#logout-btn, #settings-btn {
  padding-bottom: 10px;
  padding-top: 10px;
}

@media screen and (min-width: 700px) {
  .dropdown {
    transition-property: opacity;
    left: auto;
    right: 15px;
    border: 1px solid var(--separator-color);
    border-radius: 8px;
    padding: 5px 10px;
    min-width: 15vh;
  }
  .dropdown-item {
    padding-left: 20px;
    padding-right: 20px;
    text-align: center;
    padding-top: 15px;
  }
  #logout-btn, #settings-btn {
    padding-bottom: 15px;
    padding-top: 15px;
  }
  #top-separator {
    display: none;
  }
  .background {
    opacity: 0.5;
  }
  .fade-enter-active, .fade-leave-active {
    transition: none;
  }
}
</style>

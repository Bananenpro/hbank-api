<template>
  <div class="page">
    <div class="right">
      <img class="logo" src="@/assets/icon.svg">
      <div class="screenshot-container">
        <img class="screenshot" src="@/assets/screenshot.png">
      </div>
    </div>
    <div class="left">
      <h1 class="title">H-Bank</h1>
      <p class="description">{{$t("home.description")}}</p>
      <p class="description-long">{{$t("home.description-long")}}</p>
      <a class="more-details" href="https://github.com/juho05/hbank-web/blob/main/README.md" target="_blank">{{$t("home.more-details")}}</a>
      <div class="btns">
        <a :href="loginURL" class="btn">{{$t("home.lets-go")}}</a>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import api from '@/api'

export default defineComponent({
  name: 'Home',
  data() {
    return {
      loginURL: api.defaults.baseURL+"auth/login?redirect="+encodeURIComponent(window.location.pathname),
    }
  },
});
</script>

<style scoped>
.page {
  text-align: center;
  margin: auto;
  max-width: 700px;
}

.title {
  font-size: 40px;
}

.description {
  font-size: 15px;
  width: 90%;
  margin: auto;
}

.description-long {
  display: none;
}

.more-details {
  display: inline-block;
  margin-top: 3vh;
  font-size: 15px;
}

.logo {
  margin: auto;
  margin-top: 3vh;
  width: 70%;
  max-height: 320px;
  max-width: 320px;
}

.screenshot {
  display: none;
}

.btns {
  position: absolute;
  margin: auto;
  bottom: 5vh;
  left: 15%;
  right: 15%;

  display: flex;
  flex-direction: column;
}

.or-separator {
  margin: 2vh 0;
  font-weight: 300;
}

.btn {
  display: block;
  width: 100%;
  padding: 10px 0;
}

.login-btn {
  color: var(--fg-color-secondary);
}

@media screen and (min-width: 1000px) {
  .page {
    display: flex;
    flex-direction: row-reverse;
    max-width: 87%;
    gap: 3%;
    align-items: center;
    margin-top: -10vh;
    margin-bottom: 10vh;
    justify-content: space-between;
  }
  .left {
    text-align: left;
    max-width: 580px;
  }
  .left h1 {
    text-align: left;
    padding-top: 0;
  }
  .logo {
    display: none;
  }
  .screenshot {
    display: inline-block;
    max-height: 55vh;
    max-width: 35vw;
    width: auto;
    height: auto;
  }
  .description {
    display: none;
  }
  .description-long {
    display: block;
  }
  .btns {
    position: initial;
    width: 150px;
    margin: 0;
  }
  .more-details {
    margin-top: 0;
    margin-bottom: 8vh;
  }
}
</style>

<template>
  <div class="page form-page">
    <h2 class="title">{{ $t("settings.title") }}</h2>

    <div id="container"> 
      <div>
        <input id="sendInvitationEmail" type="checkbox" v-model="sendInvitationEmail" @change="changed = true">
        <label for="sendInvitationEmail">{{$t("settings.sendInvitationEmail")}}</label>
      </div>

      <div>
        <input id="publiclyVisible" type="checkbox" v-model="publiclyVisible" @change="changed = true">
        <label for="publiclyVisible">{{$t("settings.publiclyVisible")}}</label>
      </div>

      <div class="select-div">
        <label for="theme">{{$t("settings.theme.label")}}</label>
        <select id="theme" @change="changed = true" v-model="theme">
          <option value="system" :selected="theme == 'system'">{{$t("settings.theme.system")}}</option>
          <option value="light" :selected="theme == 'light'">{{$t("settings.theme.light")}}</option>
          <option value="dark" :selected="theme == 'dark'">{{$t("settings.theme.dark")}}</option>
        </select>
      </div>

      <div class="select-div">
        <label for="lang">{{$t("settings.lang.label")}}</label>
        <select id="lang" @change="changed = true" v-model="lang">
          <option value="system" :selected="lang == 'system'">{{$t("settings.lang.system")}}</option>
          <option value="en" :selected="lang == 'en'">{{$t("settings.lang.en")}}</option>
          <option value="de" :selected="lang == 'de'">{{$t("settings.lang.de")}}</option>
        </select>
      </div>
    </div>

    <button class="btn bottom-btn" :disabled="!changed || loading" @click="update">
      {{ loading ? $t("loading") : $t("update") }}
    </button>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";

export default defineComponent({
  name: "Account",
  data() {
    return {
      sendInvitationEmail: false,
      publiclyVisible: false,
      theme: "",
      lang: "",
      changed: false,
      loading: true
    };
  },
  methods: {
    async loadData() {
      this.loading = true

      this.theme = localStorage.getItem("theme") || "system"
      this.lang = localStorage.getItem("lang") || "system"

      const userId = await auth();
      if (userId) {
        try {
          const res = await api.get("/user/" + userId);
          if (!res.data.success) {
            console.error(res.data.message);
            return;
          }

          this.sendInvitationEmail = !res.data.dontSendInvitationEmail
          this.publiclyVisible = res.data.publiclyVisible
          this.changed = false
        } catch (e: any) {
          if (e.response) {
            this.$router.push({
              name: "error",
              query: {
                code: e.response.status,
                message: e.response.data.message,
              },
            });
          } else {
            this.$router.push({ name: "error", query: { code: "offline" } });
          }
        }
      }
      this.loading = false
    },
    async update() {
      this.loading = true

      if (this.changed && await auth()) {
        try {
          const res = await api.put("/user", {
            dontSendInvitationEmail: !this.sendInvitationEmail,
            publiclyVisible: this.publiclyVisible,
          })
          if (!res.data.success) {
            console.error(res.data.message);
            return;
          }

          this.sendInvitationEmail = !res.data.dontSendInvitationEmail
          this.publiclyVisible = res.data.publiclyVisible
          this.changed = false

          this.updateLangAndTheme()
        } catch (e: any) {
          if (e.response) {
            this.$router.push({
              name: "error",
              query: {
                code: e.response.status,
                message: e.response.data.message,
              },
            });
          } else {
            this.$router.push({ name: "error", query: { code: "offline" } });
          }
        }
      }
      this.loading = false
    },
    updateLangAndTheme() {
      if (localStorage.getItem("lang") === this.lang && localStorage.getItem("theme") === this.theme) {
        return
      }
      localStorage.setItem("lang", this.lang)
      localStorage.setItem("theme", this.theme)
      window.location.reload()
    }
  },
  mounted() {
    this.loadData()
  }
});
</script>


<style scoped>
.title {
  margin-bottom: 20vh;
  text-align: center;
  font-size: 28px;
}
#container {
  margin-left: 5%;
  margin-right: 5%;
}
input[type=checkbox] {
  margin-right: 2%;
  margin-bottom: 2vh;
}
.select-div {
  margin-top: 1.5vh;
}
@media screen and (min-width: 410px) {
  input[type=checkbox] {
    margin-right: 7px;
  }
}
</style>

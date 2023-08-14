<template>
  <PageHeader/>
  <div id="content">
    <router-view />
    <PageFooter/>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import PageHeader from '@/components/PageHeader.vue'
import PageFooter from '@/components/PageFooter.vue'

export default defineComponent({
  name: 'App',
  components: {
    PageHeader,
    PageFooter
  },
  data() {
    return {}
  },
  beforeMount() {
    let lang = localStorage.getItem("lang")
    if (lang === null) {
      lang = "system"
      localStorage.setItem("lang", lang)
    }

    if (lang === "system") {
      this.$i18n.locale = navigator.language
    } else {
      this.$i18n.locale = lang
    }
  },
  mounted() {
    let theme = localStorage.getItem("theme")
    if (theme === null) {
      theme = "system"
      localStorage.setItem("theme", theme)
    }

    if (theme === "system") {
      document.documentElement.className = this.getMediaPreference()
    } else {
      document.documentElement.className = theme
    }
  },
  methods: {
    setTheme(theme: string) {
      localStorage.setItem("theme", theme)

    },
    getMediaPreference() : string {
      const prefersDarkTheme = window.matchMedia("(prefers-color-scheme: dark)").matches

      return prefersDarkTheme ? "dark" : "light"
    }
  }
});
</script>

<style>
:root {
  --bg-color: #F5F5F5;
  --fg-color: #000000;
  --fg-color-secondary: #212121;

  --header-bg-color: #F5F5F5;
  --header-fg-color: #000000;

  --border-color: #8A8A8A;

  --button-bg-color: #0E1EAE;
  --button-fg-color: #FFFFFF;
  --button-bg-color-disabled: #8A8A8A;
  --button-fg-color-disabled: #FDFDFD;

  --input-bg-color: #F5F5F5;
  --input-fg-color: #000000;
  --input-border-color: #909090;
  --input-border-color-selected: #0E1EAE;

  --link-color: #1949F1;
  --link-color-disabled: #737373;

  --copy-box-bg-color: #E1E1E1;
  --copy-box-fg-color: #000000;

  --card-bg-color: #F5F5F5;
  --card-bg-color-transparent: #F5F5F500;
  --card-fg-color: #000000;

  --dialog-bg-color: #222222c4;

  --separator-color: #B0B0B0;

  --color-red: #C91A1A;
  --color-green: #169F2D;

  --date-in-card-color: #041185;
}

:root.dark {
  --bg-color: #101010;
  --fg-color: #FFFFFF;
  --fg-color-secondary: #DCDCDC;

  --header-bg-color: #101010;
  --header-fg-color: #FDFDFD;

  --border-color: #E4E4E4;

  --button-bg-color: #0E1EAE;
  --button-fg-color: #FFFFFF;
  --button-bg-color-disabled: #1A1A1A;
  --button-fg-color-disabled: #AEAEAE;

  --input-bg-color: #151515;
  --input-fg-color: #F2F2F2;
  --input-border-color: #434343;
  --input-border-color-selected: #0E1EAE;

  --link-color: #2DA8ED;
  --link-color-disabled: #707070;

  --copy-box-bg-color: #000000;
  --copy-box-fg-color: #FFFFFF;

  --card-bg-color: #1C1C1C;
  --card-bg-color-transparent: #1C1C1C00;
  --card-fg-color: #FFFFFF;

  --dialog-bg-color: #0f0f0fDE;

  --separator-color: #2C2C2C;

  --color-red: #C91A1A;
  --color-green: #1CC838;

  --date-in-card-color: #757ED4;
}

html,
body,
#app {
  margin: 0;
  padding: 0;
  height: 100%;
  background: var(--bg-color);
  color: var(--fg-color);
  font-family: 'Roboto', sans-serif;
}
#content {
  overflow-y: auto;
  padding: 0 2vw;
  height: calc(100% - 50px);
}

.page {
  min-height: 100%;
  position: relative;
}

.separator {
  height: 1px;
  background-color: var(--separator-color);
  margin: 1.5vh 0px;
  z-index: 3;
}

.card {
  background-color: var(--card-bg-color);
  color: var(--card-fg-color);
  border: 1px solid var(--separator-color);
  border-radius: 10px;
  padding: 3%;
  box-shadow: 0px 3px 4px 1px #00000044;
}

.dialog {
  position: absolute;
  background: var(--card-bg-color);
  color: var(--card-fg-color);
  border: 1px solid var(--separator-color);
  border-radius: 10px;
  left: 5%;
  right: 5%;
  top: 15vh;
  bottom: 25vh;
  padding: 5vh 3%;
  z-index: 100;
  overflow-y: auto;
}

.dialog-bg {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--dialog-bg-color);
  z-index: 99;
}

.dialog-title {
  font-size: 24px;
  text-align: center;
  margin-top: 0;
}

.dialog-close-btn {
  position: absolute;
  top: 15px;
  right: 20px;
}

.btn {
  display: inline-block;
  outline: none;
  background-color: var(--button-bg-color);
  color: var(--button-fg-color);
  border: none;
  border-radius: 10px;
  text-decoration: none;
  font-family: inherit;
  font-size: 16px;
  text-align: center;
  line-height: 19px;
  padding: 8px 15px;
  cursor: pointer;
}

.btn-danger {
  background-color: var(--color-red);
}

.btn:disabled {
  background-color: var(--button-bg-color-disabled);
  color: var(--button-fg-color-disabled);
  cursor: default;
}

.btn-sm {
  font-size: 14px;
  padding: 6px 12px;
}

.btn,
.clickable {
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  -webkit-touch-callout: none;
  user-select: none;
  outline: none !important;
}

.clickable:disabled {
  cursor: default;
}

.btn:hover,
.clickable:hover {
  filter: brightness(0.93);
}

.btn:hover:disabled,
.clickable:hover:disabled {
  filter: brightness(1);
}

.btn:active,
.clickable:active {
  filter: brightness(0.85);
}

.btn:active:disabled,
.clickable:active:disabled {
  filter: brightness(1);
}

.floating-action-btn {
  background-color: var(--button-bg-color);
  color: var(--button-bg-color);
  position: absolute;
  right: 30px;
  bottom: 30px;
  padding: 12px;
  border-radius: 100%;
  width: 26px;
  height: 26px;
  z-index: 98;
  box-shadow: 0px 2px 4px 1px #00000044;
}

.floating-action-btn > img {
  width: 100%;
  height: 100%;
}

a {
  color: var(--link-color);
  text-decoration: underline;
  cursor: pointer;
}

.a-disabled {
  color: var(--link-color-disabled);
  cursor: default;
}

.a-small {
  font-size: 12px;
}

form {
  width: 84%;
  margin: 0 8%;
}

label {
  display: inline-block;
  font-size: 16px;
  margin-bottom: 0.3vh;
}

.label-next-to-indicator {
  margin-left: -5px;
}

.label-extra-info {
  display: inline-block;
  font-weight: 300;
  font-size: 13px;
  margin-left: 7px;
}

.invalid-form-field-indicator {
  position: relative;
  color: red;
  font-weight: bolder;
  display: inline-block;
  width: 5px;
  left: -10px;
}

.form-error {
  display: inline-block;
  font-size: 16px;
  margin-left: -5px;
  color: red;
  margin-top: 2vh;
}

.form-error-container {
  text-align: center;
}

textarea,
select,
input[type=text],
input[type=number],
input[type=password],
input[type=date],
input[type=email] {
  display: block;
  width: calc(100% - 16px);
  padding: 8px;
  margin-left: 0;
  margin-right: 0;
  margin-bottom: 3vh;
  background-color: var(--input-bg-color);
  color: var(--input-fg-color);
  border: 1px solid var(--input-border-color);
  border-radius: 5px;
  font-size: 16px;
  line-height: 18px;
  font-family: 'Roboto', sans-serif;
}

textarea {
  resize: none;
}

textarea:focus,
select:focus,
input[type=text]:focus,
input[type=number]:focus,
input[type=password]:focus,
input[type=date]:focus,
input[type=email]:focus,
select:focus {
  outline-style: none;
  border: 1px solid var(--input-border-color-selected);
}

form > button,
.bottom-btn {
  left: 25%;
  right: 25%;
  bottom: 5vh;
  position: absolute;
  padding: 13px !important;
  display: block !important;
}

form > button,
button.bottom-btn {
  width: 50%;
}

.hcaptcha-div {
  display: flex;
  justify-content: center;
}

h1 {
  text-align: center;
  padding-top: 5vh;
  margin-top: 0;
  font-size: 36px;
}

h2 {
  font-size: 24px;
}

.box {
  background-color: var(--copy-box-bg-color);
  color: var(--copy-box-fg-color);
  padding: 10px 8px;
  font-size: 14px;
  line-height: 18px;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  border-radius: 3px;
}

.multiline-box-text {
  font-size: 14px;
  line-height: 18px;
  max-height: 91px;
  margin: 0;
  overflow-y: auto;
  overflow-wrap: anywhere;
}

.multiline-box-container {
  background-color: var(--copy-box-bg-color);
  color: var(--copy-box-fg-color);
  padding: 10px 8px;
  border-radius: 3px;
}

.positive {
  color: var(--color-green)
}
.negative {
  color: var(--color-red)
}

.danger-text {
  color: var(--color-red)
}

.edit-lbl-container {
  display: flex;
  gap: 5px;
}
.edit-btn {
  height: 18px;
}

.form-page {
  max-width: 700px;
  margin-left: auto;
  margin-right: auto;
}

@media screen and (max-height: 760px) {
  form > button,
  .bottom-btn {
    bottom: 5vh
  }
  textarea,
  select,
  input[type=text],
  input[type=number],
  input[type=password],
  input[type=date],
  input[type=email] {
    margin-bottom: 2vh;
  }
  h1 {
    padding-top: 3vh;
  }
}

@media screen and (max-height: 660px) {
  form > button,
  .bottom-btn {
    bottom: 3vh
  }
  textarea,
  select,
  input[type=text],
  input[type=number],
  input[type=password],
  input[type=date],
  input[type=email] {
    margin-bottom: 1vh;
  }
  h1 {
    font-size: 32px;
    padding-top: 2vh;
  }
}

@media screen and (max-height: 650px) {
  form > button,
  .bottom-btn {
    position: static;
    width: 50%;
    margin-left: 25%;
    margin-right: 25%;
    margin-top: 4vh;
  }

  .dialog {
    top: 5vh;
    bottom: 5vh;
    left: 2%;
    right: 2%;
  }
}

@media screen and (min-width: 600px) {
  .dialog {
    padding-left: 20px;
    padding-right: 20px;
    left: calc(50% - 260px);
    right: calc(50% - 260px);
  }
}

@media screen and (min-width: 1000px) {
  #content {
    padding: 0 4vw;
  }
  .page {
    margin-top: 2vh;
  }
}

</style>

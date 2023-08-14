<template>
  <p class="text-box">{{text}}<img @click="copy" v-show="text" class="copy-btn clickable" :src="require('@/assets/' + iconName)" alt="Copy"/></p>
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import tc from 'tinycolor2'
export default defineComponent({
  name: "CopyBtn",
  props: {
    text: String
  },
  data() {
    return {
      theme: "light",
      copied: false
    }
  },
  computed: {
    darkTheme() : boolean {
      const bgColor = getComputedStyle(document.documentElement).getPropertyValue('--copy-box-bg-color');

      const color = tc(bgColor);

      return color.isDark()
    },
    iconName() : string {
      let url = "copy-btn"

      if (this.copied) {
        url += "-check"
      }

      if (this.darkTheme) {
        url += "-dark"
      }

      url += ".svg"

      return url
    }
  },
  methods: {
    copy() {
      if (this.text !== undefined) {
        navigator.clipboard.writeText(this.text)
        this.copied = true
        setTimeout(() => this.copied = false, 20000)
      }
    }
  }
})
</script>

<style scoped>
.text-box {
  background-color: var(--copy-box-bg-color);
  color: var(--copy-box-fg-color);
  padding: 10px 8px;
  font-size: 14px;
  height: 18px;
  line-height: 18px;
  text-align: left;
}
.copy-btn {
  float: right;
  height: 24px;
  margin-top: -3px;
}

@media screen and (max-width: 365px) {
  .text-box {
    font-size: 13px;
    height: 14px;
    line-height: 14px;
  }
  .copy-btn {
    margin-top: -5px;
  }
}

@media screen and (max-width: 345px) {
  .text-box {
    font-size: 12px;
    height: 14px;
    line-height: 14px;
  }
  .copy-btn {
    height: 22px;
    margin-top: -4px;
  }
}
</style>

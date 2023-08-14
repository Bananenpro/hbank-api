<template>
<div class="page form-page">
    <h1 class="face">(._.)</h1>
    <h1 class="title">Error: {{code}}</h1>
    <p class="message">{{message ? message : ($t("errors." + code) != "errors." + code ? $t("errors."+code) : $t("errors.unknown"))}}</p>
    <p @click="back" class="btn bottom-btn">{{$t("back")}}</p>
</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
    name: "Error",
    data() {
        return {
            code: "404",
            message: ""
        }
    },
    beforeMount() {
        if (this.$route.query.code) {
            this.code = this.$route.query.code as string
        }

        if (this.$route.query.message) {
            this.message = this.$route.query.message as string
            if (!this.message.endsWith(".")) {
                this.message += "."
            }
        }
    },
    methods: {
        back() {
            this.$router.back()
        }
    }
})
</script>


<style scoped>
.face {
    padding-top: 8vh;
    font-size: 100px;
    margin-bottom: 0;
}
.title {
    padding-top: 4vh;
    margin: 0;
    font-weight: normal;
    font-size: 28px;
}
.message {
    text-align: center;
    margin: 8vh 5%;
}
.bottom-btn {
  left: 25%;
  right: 25%;
  bottom: 5vh;
  position: absolute;
  padding: 13px !important;
  margin: 0px;
  width: auto;
}
</style>

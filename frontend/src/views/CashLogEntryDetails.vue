<template>
  <div class="page form-page">
    <h1>{{ $t("cash-details.title") }}</h1>

    <label>{{ $t("time") }}</label>
    <p class="box">{{ time }}</p>

    <label>{{ $t("title") }}</label>
    <p class="box">{{ title }}</p>

    <label v-if="description">{{ $t("description") }}</label>
    <div v-if="description" class="multiline-box-container">
      <p class="multiline-box-text">{{ description }}</p>
    </div>

    <label>{{ $t("amount") }}</label>
    <p class="box">{{ amount }}{{ $t("currency") }}</p>

    <label>{{ $t("balance-difference") }}</label>
    <p
      class="box"
      :class="difference.startsWith('-') ? 'negative' : 'positive'"
    >
      {{ difference }}{{ $t("currency") }}
    </p>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";

export default defineComponent({
  name: "TransactionDetails",
  data() {
    return {
      title: "",
      description: "",
      amount: "",
      difference: "",
      time: "",
    };
  },
  async mounted() {
    try {
      const userId = await auth();
      if (userId) {
        const res = await api.get(`/user/cash/${this.$route.params.entryId}`);
        if (!res.data.success) {
          console.error(res.data.message);
          return;
        }
        this.title = res.data.title;
        this.description = res.data.description;
        this.amount = (res.data.amount / 100.0)
          .toFixed(2)
          .replace(".", this.$t("decimal"));

        this.time = new Date(res.data.time * 1000).toLocaleString([], {
          day: "2-digit",
          month: "2-digit",
          year: "2-digit",
          hour: "2-digit",
          minute: "2-digit",
        });

        this.difference = (res.data.difference / 100.0)
          .toFixed(2)
          .replace(".", this.$t("decimal"));
        if (!this.difference.startsWith("-")) {
          this.difference = "+" + this.difference;
        }
      }
    } catch (e: any) {
      if (e.response) {
        this.$router.push({
          name: "error",
          query: { code: e.response.status, message: e.response.data.message },
        });
      } else {
        this.$router.push({ name: "error", query: { code: "offline" } });
      }
    }
  },
});
</script>

<style scoped>
h1 {
  margin-bottom: 7vh;
}

.box,
.multiline-box-container {
  margin-bottom: 2vh;
  min-height: 18px;
}

p {
  margin-bottom: 0;
}
</style>

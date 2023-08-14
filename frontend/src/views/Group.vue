<template>
  <div class="page">
    <div class="top-container">
      <h2 v-if="member || admin" class="balance-lbl">
        {{ !member && admin ? $t('group.total') : $t("group.balance") }}: {{ balance }}{{ $t("currency") }}
      </h2>
      <router-link :to="'/group/' + $route.params.id + '/settings'"
        ><img
          class="settings-btn clickable"
          :src="
            darkTheme
              ? require('@/assets/settings-light.svg')
              : require('@/assets/settings-dark.svg')
          "
      /></router-link>
    </div>
    <div class="cards">
      <TransactionsCard class="card"
        v-if="member || admin"
        :groupId="$route.params.id.toString()"
        :onlyAdmin="!member && admin"
      />
      <PaymentPlansCard class="card"
        v-if="member || admin"
        :groupId="$route.params.id.toString()"
        :onlyAdmin="!member && admin"
      />
      <MembersCard class="card" :groupId="$route.params.id.toString()" :showAddBtn="admin" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
import { pageTitle } from "@/router";
import tc from "tinycolor2";
import TransactionsCard from "@/components/TransactionsCard.vue";
import PaymentPlansCard from "@/components/PaymentPlansCard.vue";
import MembersCard from "@/components/MembersCard.vue";

export default defineComponent({
  name: "Group",
  components: {
    TransactionsCard,
    PaymentPlansCard,
    MembersCard,
  },
  data() {
    return {
      id: "",
      name: "",
      member: false,
      admin: false,
      balance: "",
    };
  },
  computed: {
    darkTheme(): boolean {
      const bgColor = getComputedStyle(
        document.documentElement
      ).getPropertyValue("--bg-color");

      const color = tc(bgColor);

      return color.isDark();
    },
  },
  mounted() {
    this.loadData();
  },
  methods: {
    async loadData() {
      const userId = await auth();
      if (userId) {
        if (this.$route.params.id) {
          try {
            const res = await api.get("/group/" + this.$route.params.id);
            if (!res.data.success) {
              console.error(res.data.message);
              return;
            }

            this.id = res.data.id;
            this.name = res.data.name;
            this.member = res.data.member;
            this.admin = res.data.admin;

            pageTitle.value = this.name;

            if (this.member) {
              const balanceRes = await api.get(
                "/group/" + this.id + "/transaction/balance"
              );
              if (!balanceRes.data.success) {
                console.error(balanceRes.data.message);
                return;
              }
              this.balance = (balanceRes.data.balance / 100.0)
                .toFixed(2)
                .replace(".", this.$t("decimal"));
            } else if (this.admin) {
              const totalRes = await api.get(
                "/group/" + this.id + "/total"
              );
              if (!totalRes.data.success) {
                console.error(totalRes.data.message);
                return;
              }
              this.balance = (totalRes.data.total / 100.0)
                .toFixed(2)
                .replace(".", this.$t("decimal"));
            }
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
      }
    },
  },
});
</script>


<style scoped>
.top-container {
  display: flex;
  margin-top: 1vh;
  margin-bottom: 3vh;
  justify-content: space-between;
}
.balance-lbl {
  flex-grow: 1;
  flex-shrink: 1;
  min-width: 0;
  line-height: 30px;
  margin: 0;
}
.settings-btn {
  height: 30px;
  margin-right: 1%;
}
.card {
  margin-bottom: 3vh;
}
@media screen and (min-width: 1150px) {
  .card {
    padding: 27px;
    flex-grow: 1;
    flex-basis: 100%;
  }
  .cards {
    display: flex;
    justify-content: space-around;
    flex-wrap: nowrap;
    gap: 3vw;
  }
}
</style>

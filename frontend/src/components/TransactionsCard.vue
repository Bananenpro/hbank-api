<template>
  <div class="transactions-card card">
    <div class="card-header">
      <h3 class="title">{{ $t("group.transactions") }}</h3>
      <img
        @click="$router.push('/group/' + groupId + '/transfer')"
        class="clickable"
        :src="
          darkTheme
            ? require('@/assets/add-in-card-light.svg')
            : require('@/assets/add-in-card-dark.svg')
        "
        alt="+"
      />
    </div>
    <div class="separator"></div>
    <div
      class="list"
      @click="$router.push('/group/' + groupId + '/transaction')"
    >
      <div
        class="transaction"
        v-for="transaction in transactions"
        :key="transaction.id"
      >
        <p class="transaction-title">{{ transaction.title }}</p>
        <p
          class="transaction-amount"
          :class="transaction.senderId === userId ? 'negative' : 'positive'"
        >
          {{
            (transaction.senderId === userId ? "-" : "+") +
            (transaction.amount / 100.0)
              .toFixed(2)
              .replace(".", $t("decimal")) +
            $t("currency")
          }}
        </p>
      </div>
      <div class="gradient"></div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
import tc from "tinycolor2";

interface Transaction {
  id: string;
  title: string;
  amount: number;
  senderId: string;
}

export default defineComponent({
  name: "TransactionsCard",
  props: {
    groupId: {
      type: String,
      required: true,
    },
    onlyAdmin: Boolean,
  },
  data() {
    return {
      transactions: [] as Transaction[],
      transactionCount: 0,
      userId: "",
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
  methods: {
    async loadTransactions() {
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.get(
            "/group/" +
              this.groupId +
              "/transaction?pageSize=" +
              this.transactionCount +
              "&bank=" +
              this.onlyAdmin
          );
          if (!res.data.success) {
            console.error(res.data.message);
            return;
          }

          this.transactions = []
          for (let i = 0; i < res.data.transactions.length; i++) {
            this.transactions.push(res.data.transactions[i]);
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
    },
    onResize() {
      if (Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0) >= 1150) {
        this.transactionCount = 20;
      } else {
        this.transactionCount = 5;
      }
    }
  },
  async mounted() {
    window.addEventListener("resize", this.onResize);

    const userId = await auth();
    if (this.onlyAdmin) {
      this.userId = "bank";
    } else {
      this.userId = userId;
    }

    this.onResize()
  },
  unmounted() {
    window.removeEventListener("resize", this.onResize);
  },
  watch: {
    async transactionCount(newVal: number, oldVal: number) {
      if (newVal > oldVal)
        await this.loadTransactions()
    }
  }
});
</script>


<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
}
.title {
  margin: 0;
}
.separator {
  margin-top: 5px;
  margin-bottom: 10px;
}
.list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 52px;
  max-height: 142px;
  overflow: hidden;
  position: relative;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  -webkit-touch-callout: none;
  user-select: none;
  outline: none !important;
}
.gradient {
  position: absolute;
  top: 15%;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(
    0deg,
    var(--card-bg-color) 0%,
    var(--card-bg-color-transparent) 100%
  );
}
.transaction {
  display: flex;
  gap: 7px;
  justify-content: space-between;
}
.transaction-title {
  margin: 0;
  font-size: 18px;
  line-height: 22px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex-shrink: 1;
}
.transaction-amount {
  margin: 0;
  font-size: 18px;
  line-height: 22px;
  text-align: right;
}

@media screen and (min-width: 1150px) {
  .list {
    min-height: 90%;
    max-height: 50vh;
  }
  .transactions-card {
    min-height: 25vh;
  }
}
</style>

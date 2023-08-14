<template>
  <div class="page form-page">
    <input
      class="search"
      :class="isMember && isAdmin ? 'no-search-padding' : ''"
      v-model="searchInput"
      type="text"
      :placeholder="$t('placeholders.search')"
    />
    <div v-if="isMember && isAdmin" class="bank">
      <input type="checkbox" name="bank" v-model="bank" id="bank" />
      <label for="bank">{{ $t("bank") }}</label>
    </div>
    <router-link :to="'/group/' + groupId + '/transfer'" id="create-transaction-btn-desktop" class="btn clickable">+ {{ $t("transaction-log.create") }}</router-link>
    <div ref="list">
      <div
        class="entry card clickable"
        @click="$router.push('/group/' + groupId + '/transaction/' + entry.id)"
        v-for="entry in log"
        :key="entry.id"
      >
        <p class="time">{{ entry.time }}</p>
        <p class="title">{{ entry.title }}</p>
        <p
          class="amount"
          :class="entry.amount.startsWith('-') ? 'negative' : 'positive'"
        >
          {{ entry.amount + $t("currency") }}
        </p>
      </div>
    </div>
    <teleport to="#app">
      <router-link
        :to="'/group/' + groupId + '/transfer'"
        id="create-transaction-btn-mobile"
        class="floating-action-btn clickable"
        ><img src="@/assets/add.svg" alt="+"
      /></router-link>
    </teleport>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";

interface Entry {
  id: string;
  time: string;
  title: string;
  amount: string;
}

export default defineComponent({
  name: "TransactionLog",
  data() {
    return {
      log: [] as Entry[],
      searchInput: "",
      searchTimeout: 0,
      page: 0,
      pageSize: 20,
      groupId: this.$route.params.id,
      onScrollInterval: 0,
      loading: false,
      isMember: false,
      isAdmin: false,
      bank: false,
    };
  },
  methods: {
    async load() {
      if (!this.loading && this.log.length >= this.page * this.pageSize) {
        this.loading = true;
        const userId = await auth();
        if (userId) {
          const res = await api.get(
            `/group/${this.groupId}/transaction?page=${this.page}&pageSize=${this.pageSize}&search=${this.searchInput}&bank=${this.bank}`
          );
          if (!res.data.success) {
            console.error(res.data.message);
            this.loading = false;
            return;
          }

          for (let i = 0; i < res.data.transactions.length; i++) {
            const time = new Date(
              res.data.transactions[i].time * 1000
            ).toLocaleDateString([], {
              day: "2-digit",
              month: "2-digit",
              year: "2-digit",
            });

            let amount = (res.data.transactions[i].amount / 100)
              .toFixed(2)
              .replace(".", this.$t("decimal"));
            if (
              res.data.transactions[i].receiverId ==
              (this.bank ? "bank" : userId)
            ) {
              amount = "+" + amount;
            } else {
              amount = "-" + amount;
            }

            this.log.push({
              id: res.data.transactions[i].id,
              time: time,
              title: res.data.transactions[i].title,
              amount: amount,
            });
          }
          this.page++;
        }
        this.loading = false;
      }
    },
    async onScroll(): Promise<boolean> {
      const contentElement = document.getElementById("content");
      const list = this.$refs.list as HTMLElement;

      if (contentElement) {
        const nearBottom =
          contentElement.scrollTop + window.innerHeight >=
          list.offsetHeight * 0.8;
        if (nearBottom) {
          await this.load();
        }
        return nearBottom;
      }

      return false;
    },
  },
  watch: {
    searchInput: function () {
      clearTimeout(this.searchTimeout);
      this.searchTimeout = setTimeout(() => {
        this.log = [];
        this.page = 0;
        this.load();
      }, 500);
    },
    bank() {
      this.log = [];
      this.page = 0;
      this.load();
    },
  },
  async mounted() {
    const userId = await auth();
    if (userId) {
      try {
        const groupRes = await api.get(`/group/${this.$route.params.id}`);
        if (!groupRes.data.success) {
          console.error(groupRes.data.message);
          return;
        }
        this.isAdmin = groupRes.data.admin;

        const res = await api.get("/group/" + this.$route.params.id);
        if (!res.data.success) {
          console.error(res.data.message);
        }
        this.isMember = res.data.member;
        this.isAdmin = res.data.admin;
        this.bank = this.isAdmin && !this.isMember;
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

    this.onScrollInterval = setInterval(this.onScroll, 200);

    await this.load();

    const contentElement = document.getElementById("content");
    if (contentElement) {
      contentElement.addEventListener("scroll", this.onScroll);
    }
  },
  unmounted() {
    clearInterval(this.onScrollInterval);
    const contentElement = document.getElementById("content");
    if (contentElement) {
      contentElement.removeEventListener("scroll", this.onScroll);
    }
  },
});
</script>


<style scoped>
.search {
  margin-top: 1vh;
}
.no-search-padding {
  margin-bottom: 0;
}
.bank {
  margin-top: 1vh;
  margin-bottom: 1.5vh;
}
.entry {
  display: flex;
  padding: 2.5% 2%;
  gap: 7px;
  margin-bottom: 1.5vh;
}
.title {
  margin: 0;
  flex-grow: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.time {
  margin: 0;
  color: var(--date-in-card-color);
}
.amount {
  margin: 0;
  text-align: right;
}

#create-transaction-btn-desktop {
  margin-bottom: 1.5vh;
  display: none;
}

@media screen and (min-width: 470px) {
  .entry {
    padding: 12px 2%;
  }
}

@media screen and (min-width: 700px){
  #create-transaction-btn-desktop {
    display: inline-block;
  }
  #create-transaction-btn-mobile {
    display: none;
  }
}
</style>

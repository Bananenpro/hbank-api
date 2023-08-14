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
    <router-link :to="'/group/' + groupId + '/payment-plan/create'" id="create-payment-plan-btn-desktop" class="btn clickable">+ {{ $t("payment-plans.create") }}</router-link>
    <div ref="list">
      <div
        v-for="plan in paymentPlans"
        :key="plan.id"
        class="payment-plan card clickable"
        @click="$router.push('/group/' + groupId + '/payment-plan/' + plan.id)"
      >
        <p class="next-payment">{{ plan.nextPayment }}</p>
        <p class="name">{{ plan.name }}</p>
        <p
          class="amount"
          :class="
            plan.senderId === (bank ? 'bank' : userId) ? 'negative' : 'positive'
          "
        >
          {{
            (plan.senderId === (bank ? "bank" : userId) ? "-" : "+") +
            (plan.amount / 100.0).toFixed(2).replace(".", $t("decimal")) +
            $t("currency")
          }}
        </p>
      </div>
    </div>
    <teleport to="#app">
      <router-link
        :to="'/group/' + groupId + '/payment-plan/create'"
        id="create-payment-plan-btn-mobile"
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
import { DateTime } from "luxon";

interface PaymentPlan {
  id: string;
  name: string;
  amount: number;
  senderId: string;
  nextPayment: string;
}

export default defineComponent({
  name: "TransactionLog",
  data() {
    return {
      paymentPlans: [] as PaymentPlan[],
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
      userId: "",
    };
  },
  methods: {
    async load() {
      if (
        !this.loading &&
        this.paymentPlans.length >= this.page * this.pageSize
      ) {
        this.loading = true;
        const userId = await auth();
        if (userId) {
          const res = await api.get(
            "/group/" +
              this.groupId +
              "/paymentPlan?page=" +
              this.page +
              "&pageSize=" +
              this.paymentPlans +
              "&search=" +
              this.searchInput +
              "&bank=" +
              this.bank
          );
          if (!res.data.success) {
            console.error(res.data.message);
            return;
          }

          for (let i = 0; i < res.data.paymentPlans.length; i++) {
            const nextPaymentDate = DateTime.fromSeconds(
              res.data.paymentPlans[i].nextExecute
            );

            let nextPayment = "";

            const diffInMonths = nextPaymentDate.diffNow("months");
            if (Math.ceil(diffInMonths.months) >= 12) {
              const diffInYears = nextPaymentDate.diffNow("years");
              nextPayment = Math.ceil(diffInYears.years) + "a";
            } else {
              const diffInWeeks = nextPaymentDate.diffNow("weeks");
              if (Math.ceil(diffInWeeks.weeks) >= 5) {
                nextPayment = Math.ceil(diffInMonths.months) + "m";
              } else {
                if (diffInWeeks.weeks >= 1) {
                  nextPayment = Math.ceil(diffInWeeks.weeks) + "w";
                } else {
                  const diffInDays = nextPaymentDate.diffNow("days");
                  nextPayment = Math.ceil(diffInDays.days) + "d";
                }
              }
            }

            this.paymentPlans.push({
              id: res.data.paymentPlans[i].id,
              name: res.data.paymentPlans[i].name,
              amount: res.data.paymentPlans[i].amount,
              senderId: res.data.paymentPlans[i].senderId,
              nextPayment: nextPayment,
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
        this.paymentPlans = [];
        this.page = 0;
        this.load();
      }, 500);
    },
    bank() {
      this.paymentPlans = [];
      this.page = 0;
      this.load();
    },
  },
  async mounted() {
    const userId = await auth();
    if (userId) {
      this.userId = userId;
      try {
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
.payment-plan {
  display: flex;
  padding: 2.5% 2%;
  gap: 7px;
  margin-bottom: 1.5vh;
}
.name {
  margin: 0;
  flex-grow: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.next-payment {
  margin: 0;
  color: var(--date-in-card-color);
  width: 36px;
}
.amount {
  margin: 0;
  text-align: right;
}

#create-payment-plan-btn-desktop {
  margin-bottom: 1.5vh;
  display: none;
}

@media screen and (min-width: 470px) {
  .payment-plan {
    padding: 12px 2%;
  }
}

@media screen and (min-width: 700px){
  #create-payment-plan-btn-desktop {
    display: inline-block;
  }
  #create-payment-plan-btn-mobile {
    display: none;
  }
}
</style>

<template>
  <div class="payment-plans-card card">
    <div class="card-header">
      <h3 class="title">{{ $t("group.payment-plans") }}</h3>
      <img
        @click="$router.push('/group/' + groupId + '/payment-plan/create')"
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
      @click="$router.push('/group/' + groupId + '/payment-plan')"
    >
      <div class="payment-plan" v-for="plan in paymentPlans" :key="plan.id">
        <p class="payment-plan-next">{{ plan.nextPayment }}</p>
        <p class="payment-plan-name">{{ plan.name }}</p>
        <p
          class="payment-plan-amount"
          :class="plan.senderId === userId ? 'negative' : 'positive'"
        >
          {{
            (plan.senderId === userId ? "-" : "+") +
            (plan.amount / 100.0).toFixed(2).replace(".", $t("decimal")) +
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
import { DateTime } from "luxon";
import tc from "tinycolor2";

interface PaymentPlan {
  id: string;
  name: string;
  amount: number;
  senderId: string;
  nextPayment: string;
}

export default defineComponent({
  name: "PaymentPlans",
  props: {
    groupId: {
      type: String,
      required: true,
    },
    onlyAdmin: Boolean,
  },
  data() {
    return {
      paymentPlans: [] as PaymentPlan[],
      paymentPlanCount: 0,
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
    async load() {
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.get(
            "/group/" +
              this.groupId +
              "/paymentPlan?pageSize=" +
              this.paymentPlanCount +
              "&bank=" +
              this.onlyAdmin
          );
          if (!res.data.success) {
            console.error(res.data.message);
            return;
          }

          this.paymentPlans = []
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
        this.paymentPlanCount = 20;
      } else {
        this.paymentPlanCount = 5;
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
    async paymentPlanCount(newVal: number, oldVal: number) {
      if (newVal > oldVal)
        await this.load()
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
.payment-plan {
  display: flex;
  gap: 7px;
  justify-content: space-between;
}
.payment-plan-next {
  margin: 0;
  font-size: 18px;
  line-height: 22px;
  color: var(--date-in-card-color);
  text-align: right;
  width: 36px;
}
.payment-plan-name {
  margin: 0;
  font-size: 18px;
  line-height: 22px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex-grow: 1;
}
.payment-plan-amount {
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
  .payment-plans-card {
    min-height: 25vh;
  }
}
</style>

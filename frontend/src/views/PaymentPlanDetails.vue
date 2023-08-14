<template>
  <div class="page form-page">
    <ConfirmDialog
      :show="showDeleteDialog"
      @yes="deletePaymentPlan"
      @close="showDeleteDialog = false"
      :description="$t('payment-plan-details.delete-dialog-description')"
    />

    <div class="title-container">
      <h1 class="title">{{ $t("payment-plan-details.title") }}</h1>
      <img
        v-if="senderId == userId || (senderId == 'bank' && isAdmin)"
        @click="edit"
        class="edit-btn clickable"
        :src="
        darkTheme
        ? require('@/assets/edit-light.svg')
        : require('@/assets/edit-dark.svg')
        "
        />
    </div>

    <label>{{ $t("name") }}</label>
    <p class="box">{{ name }}</p>

    <label v-if="description">{{ $t("description") }}</label>
    <div v-if="description" class="multiline-box-container">
      <p class="multiline-box-text">{{ description }}</p>
    </div>

    <label>{{ $t("amount") }}</label>
    <p
      class="box"
      :class="
        amount.startsWith('+')
          ? 'positive'
          : amount.startsWith('-')
          ? 'negative'
          : ''
      "
    >
      {{ amount }}{{ $t("currency") }}
    </p>

    <div class="users">
      <div
        v-if="senderId && !amount.startsWith('-')"
        :class="
          !amount.startsWith('+') && !amount.startsWith('-') ? 'two-users' : ''
        "
      >
        <label>{{ $t("sender") }}</label>
        <div class="box profile-picture-box">
          <img
            v-if="senderId == 'bank'"
            class="profile-picture"
            :src="
              darkTheme
                ? require('@/assets/bank-light.svg')
                : require('@/assets/bank-dark.svg')
            "
          />
          <ProfilePicture
            v-if="senderId != 'bank'"
            class="profile-picture"
            :user-id="senderId"
          />
          <p class="name">{{ senderName }}</p>
        </div>
      </div>

      <div
        v-if="receiverId && !amount.startsWith('+')"
        :class="
          !amount.startsWith('+') && !amount.startsWith('-') ? 'two-users' : ''
        "
      >
        <label>{{ $t("receiver") }}</label>
        <div class="box profile-picture-box">
          <img
            v-if="receiverId == 'bank'"
            class="profile-picture"
            :src="
              darkTheme
                ? require('@/assets/bank-light.svg')
                : require('@/assets/bank-dark.svg')
            "
          />
          <ProfilePicture
            v-if="receiverId != 'bank'"
            class="profile-picture"
            :user-id="receiverId"
          />
          <p class="name">{{ receiverName }}</p>
        </div>
      </div>
    </div>

    <label>{{ $t("payment-plan-details.recurrences") }}</label>
    <p class="box">{{ recurrences }}</p>

    <label>{{ $t("payment-plan-details.next-payment") }}</label>
    <p class="box">{{ nextPayment }} ({{ nextPaymentDistance }})</p>

    <button
      v-if="senderId == userId || (senderId == 'bank' && isAdmin)"
      class="btn bottom-btn btn-danger"
      @click="showDeleteDialog = true"
    >
      {{ $t("delete") }}
    </button>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
import tc from "tinycolor2";
import { DateTime } from "luxon";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import ProfilePicture from "@/components/ProfilePicture.vue";

export default defineComponent({
  name: "PaymentPlanDetails",
  components: {
    ConfirmDialog,
    ProfilePicture,
  },
  data() {
    return {
      name: "",
      description: "",
      amount: "",
      nextPayment: "",
      nextPaymentDistance: "",
      recurrences: "",
      senderId: "",
      senderName: "",
      receiverId: "",
      receiverName: "",
      baseUrl: api.defaults.baseURL,
      showDeleteDialog: false,
      isAdmin: false,
      userId: "",
    };
  },
  computed: {
    darkTheme(): boolean {
      const bgColor = getComputedStyle(
        document.documentElement
      ).getPropertyValue("--copy-box-bg-color");

      const color = tc(bgColor);

      return color.isDark();
    },
  },
  methods: {
    async deletePaymentPlan() {
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.delete(
            `/group/${this.$route.params.id}/paymentPlan/${this.$route.params.paymentPlanId}`
          );
          if (res.data.success) {
            this.$router.back();
          } else {
            console.error(res.data.message);
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
    edit() {
      this.$router.push(`/group/${this.$route.params.id}/payment-plan/${this.$route.params.paymentPlanId}/update`)
    }
  },
  async mounted() {
    try {
      this.userId = await auth();
      if (this.userId) {
        const groupRes = await api.get(`/group/${this.$route.params.id}`);
        if (!groupRes.data.success) {
          console.error(groupRes.data.message);
          return;
        }
        this.isAdmin = groupRes.data.admin;

        const res = await api.get(
          `/group/${this.$route.params.id}/paymentPlan/${this.$route.params.paymentPlanId}`
        );
        if (!res.data.success) {
          console.error(res.data.message);
          return;
        }
        this.name = res.data.name;
        this.description = res.data.description;

        this.recurrences = this.$t(
          "payment-plan-details.next-payment-distance-" + res.data.scheduleUnit,
          res.data.schedule
        );

        this.nextPayment = new Date(res.data.nextExecute * 1000).toLocaleString(
          [],
          {
            day: "2-digit",
            month: "2-digit",
            year: "2-digit",
          }
        );

        const nextPaymentDate = DateTime.fromSeconds(res.data.nextExecute);

        const diffInMonths = nextPaymentDate.diffNow("months");
        if (Math.ceil(diffInMonths.months) >= 12) {
          const diffInYears = nextPaymentDate.diffNow("years");
          this.nextPaymentDistance = Math.ceil(diffInYears.years) + "a";
        } else {
          const diffInWeeks = nextPaymentDate.diffNow("weeks");
          if (Math.ceil(diffInWeeks.weeks) >= 5) {
            this.nextPaymentDistance = Math.ceil(diffInMonths.months) + "m";
          } else {
            if (diffInWeeks.weeks >= 1) {
              this.nextPaymentDistance = Math.ceil(diffInWeeks.weeks) + "w";
            } else {
              const diffInDays = nextPaymentDate.diffNow("days");
              this.nextPaymentDistance = Math.ceil(diffInDays.days) + "d";
            }
          }
        }

        if (
          this.isAdmin &&
          (res.data.senderId == "bank" || res.data.receiverId == "bank")
        ) {
          this.amount = (res.data.amount / 100.0)
            .toFixed(2)
            .replace(".", this.$t("decimal"));
        } else {
          if (res.data.receiverId == this.userId) {
            this.amount =
              "+" +
              (res.data.amount / 100.0)
                .toFixed(2)
                .replace(".", this.$t("decimal"));
          } else {
            this.amount = (-res.data.amount / 100.0)
              .toFixed(2)
              .replace(".", this.$t("decimal"));
          }
        }

        if (res.data.senderId == "bank") {
          this.senderId = "bank";
          this.senderName = this.$t("bank");
        } else {
          const userRes = await api.get(`/user/${res.data.senderId}`);
          if (!userRes.data.success) {
            console.error(userRes.data.message);
            return;
          }
          this.senderId = userRes.data.id;
          this.senderName = userRes.data.name;
        }

        if (res.data.receiverId == "bank") {
          this.receiverId = "bank";
          this.receiverName = this.$t("bank");
        } else {
          const userRes = await api.get(`/user/${res.data.receiverId}`);
          if (!userRes.data.success) {
            console.error(userRes.data.message);
            return;
          }
          this.receiverId = userRes.data.id;
          this.receiverName = userRes.data.name;
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

.title-container {
  text-align: center;
  padding-top: 5vh;
  margin-bottom: 2vh;
}

.title {
  display: inline;
}

.edit-btn {
  height: 100%;
  margin-left: 1.5%;
  margin-bottom: -2px;
}

.users {
  display: flex;
  justify-content: space-evenly;
  gap: 4%;
}

.users > div {
  flex-grow: 1;
}

.two-users {
  width: 48%;
}

.name {
  line-height: 32px;
  margin: 0;
  flex-grow: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.box,
.multiline-box-container {
  margin-bottom: 2vh;
  min-height: 18px;
}

p {
  margin-bottom: 0;
}

.profile-picture-box {
  padding-top: 5px;
  padding-bottom: 5px;
  display: flex;
  gap: 7px;
}

.profile-picture {
  border-radius: 100%;
  width: 32px;
  height: 32px;
}

.payment-plan {
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bottom-btn {
  bottom: 3vh;
}

@media screen and (max-height: 740px) {
  .bottom-btn {
    bottom: 2vh;
  }
}

@media screen and (max-height: 720px) {
  .bottom-btn {
    bottom: 1vh;
  }
}

@media screen and (max-height: 700px) {
  .bottom-btn {
    position: static;
    margin-left: 25%;
    margin-right: 25%;
    margin-top: 6vh;
  }
}
</style>

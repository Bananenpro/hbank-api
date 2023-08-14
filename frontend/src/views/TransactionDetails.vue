<template>
  <div class="page form-page">
    <h1>{{ $t("transaction.details") }}</h1>

    <label>{{ $t("title") }}</label>
    <p class="box">{{ title }}</p>

    <label v-if="description">{{ $t("description") }}</label>
    <div v-if="description" class="multiline-box-container">
      <p class="multiline-box-text">{{ description }}</p>
    </div>

    <label
      v-if="
        balanceDifference.startsWith('+') || balanceDifference.startsWith('-')
      "
      >{{ $t("transaction.new-balance") }}</label
    >
    <p
      v-if="
        balanceDifference.startsWith('+') || balanceDifference.startsWith('-')
      "
      class="box"
    >
      {{ balance }}{{ $t("currency") }}
    </p>

    <label>{{
      !balanceDifference.startsWith("+") && !balanceDifference.startsWith("-")
        ? $t("amount")
        : $t("balance-difference")
    }}</label>
    <p
      class="box"
      :class="
        balanceDifference.startsWith('+')
          ? 'positive'
          : balanceDifference.startsWith('-')
          ? 'negative'
          : ''
      "
    >
      {{ balanceDifference }}{{ $t("currency") }}
    </p>

    <div class="users">
      <div
        v-if="senderId && !balanceDifference.startsWith('-')"
        :class="
          !balanceDifference.startsWith('+') &&
          !balanceDifference.startsWith('-')
            ? 'two-users'
            : ''
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
        v-if="receiverId && !balanceDifference.startsWith('+')"
        :class="
          !balanceDifference.startsWith('+') &&
          !balanceDifference.startsWith('-')
            ? 'two-users'
            : ''
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

    <label>{{ $t("time") }}</label>
    <p class="box">{{ time }}</p>

    <p v-if="paymentPlanId" class="payment-plan">
      {{ $t("payment-plan") }}:
      <router-link
        :to="'/group/' + $route.params.id + '/payment-plan/' + paymentPlanId"
        >{{ paymentPlanName }}</router-link
      >
    </p>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
import tc from "tinycolor2";
import ProfilePicture from "@/components/ProfilePicture.vue";

export default defineComponent({
  name: "TransactionDetails",
  components: {
    ProfilePicture,
  },
  data() {
    return {
      title: "",
      description: "",
      balance: "",
      balanceDifference: "",
      time: "",
      senderId: "",
      senderName: "",
      receiverId: "",
      receiverName: "",
      paymentPlanId: "",
      paymentPlanName: "",
      baseUrl: api.defaults.baseURL,
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
          `/group/${this.$route.params.id}/transaction/${this.$route.params.transactionId}`
        );
        if (!res.data.success) {
          console.error(res.data.message);
          return;
        }
        this.title = res.data.title;
        this.description = res.data.description;
        this.balance = (res.data.newBalance / 100.0)
          .toFixed(2)
          .replace(".", this.$t("decimal"));

        this.time = new Date(res.data.time * 1000).toLocaleString([], {
          day: "2-digit",
          month: "2-digit",
          year: "2-digit",
          hour: "2-digit",
          minute: "2-digit",
        });

        if (
          this.isAdmin &&
          (res.data.senderId == "bank" || res.data.receiverId == "bank")
        ) {
          this.balanceDifference = (res.data.amount / 100.0)
            .toFixed(2)
            .replace(".", this.$t("decimal"));
        } else {
          if (res.data.receiverId == this.userId) {
            this.balanceDifference =
              "+" +
              (res.data.amount / 100.0)
                .toFixed(2)
                .replace(".", this.$t("decimal"));
          } else {
            this.balanceDifference = (-res.data.amount / 100.0)
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
          } else {
            this.receiverId = userRes.data.id;
            this.receiverName = userRes.data.name;
          }
        }

        this.paymentPlanId = res.data.paymentPlanId;
        if (this.paymentPlanId) {
          const planRes = await api.get(
            `/group/${this.$route.params.id}/paymentPlan/${this.paymentPlanId}`
          );
          if (!planRes.data.success) {
            console.error(planRes.data.message);
          } else {
            this.paymentPlanName = planRes.data.name;
          }
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
</style>

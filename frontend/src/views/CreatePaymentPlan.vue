<template>
  <div class="page form-page">
    <h1>{{ $t("payment-plan") }}</h1>
    <form @submit.prevent="create">
      <UserChooser
        :includeSelf="fromBank"
        :showBank="!fromBank"
        :url="membersUrl"
        :label="$t('to-lbl')"
        @selected="userSelected"
      />
      <span class="invalid-form-field-indicator">{{
        validAmount ? "" : "!"
      }}</span
      ><label class="label-next-to-indicator" for="amount">{{
        $t("amount")
      }}</label>
      <MoneyInput @changed="amountChanged" name="amount" />

      <div v-if="isMember && isAdmin" class="from-bank">
        <input
          type="checkbox"
          name="from-bank"
          v-model="fromBank"
          id="from-bank"
        />
        <label for="from-bank">{{
          $t("transaction.transfer-from-bank")
        }}</label>
      </div>

      <span class="invalid-form-field-indicator">{{
        validFirstExecute ? "" : "!"
      }}</span
      ><label class="label-next-to-indicator" for="first-execute">{{
        $t("payment-plan-create.first-execute")
      }}</label>
      <input
        type="date"
        :min="tomorrow.toISOString().split('T')[0]"
        placeholder="yyyy-mm-dd"
        name="firstExecute"
        v-model="firstExecute"
        id="first-execute"
      />

      <div class="schedule-container">
        <span class="invalid-form-field-indicator schedule-lbl-error">{{
          validSchedule ? "" : "!"
        }}</span>
        <div class="schedule-input-container">
          <label class="label-next-to-indicator schedule-lbl" for="schedule">{{
            $t("payment-plan-create.every")
          }}</label>
          <input
            class="schedule"
            type="number"
            min="1"
            name="schedule"
            v-model="schedule"
            id="schedule"
          />
          <select
            class="schedule-unit"
            name="schedule-unit"
            v-model="scheduleUnit"
          >
            <option value="day">{{ $t("days") }}</option>
            <option value="week">{{ $t("weeks") }}</option>
            <option value="month">{{ $t("months") }}</option>
            <option value="year">{{ $t("years") }}</option>
          </select>
        </div>
      </div>

      <span class="invalid-form-field-indicator">{{
        validName ? "" : "!"
      }}</span
      ><label class="label-next-to-indicator" for="name">{{
        $t("name")
      }}</label>
      <input type="text" name="name" v-model="name" id="name" />

      <span class="invalid-form-field-indicator">{{
        validDescription ? "" : "!"
      }}</span
      ><label class="label-next-to-indicator" for="description">{{
        $t("description")
      }}</label>
      <textarea
        type="text"
        name="description"
        v-model="description"
        id="description"
        rows="4"
      ></textarea>

      <button
        class="btn"
        :disabled="
          !receiverId ||
          !validAmount ||
          !validName ||
          !validFirstExecute ||
          !validSchedule ||
          !validDescription ||
          loading
        "
        type="submit"
      >
        {{ loading ? $t("loading") : $t("create") }}
      </button>
    </form>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import MoneyInput from "@/components/MoneyInput.vue";
import UserChooser from "@/components/UserChooser.vue";
import api, { auth } from "@/api";
import { config } from "@/api";

export default defineComponent({
  name: "CreatePaymentPlan",
  components: {
    MoneyInput,
    UserChooser,
  },
  data() {
    return {
      name: "",
      description: "",
      amount: 0,
      validAmount: false,
      loading: false,
      receiverId: "",
      fromBank: false,
      isAdmin: false,
      isMember: false,
      firstExecute: "",
      schedule: 1,
      scheduleUnit: "day",
      minNameLength: 0,
      maxNameLength: 0,
      minDescriptionLength: 0,
      maxDescriptionLength: 0
    };
  },
  async beforeCreate() {
    this.minNameLength = (await config()).minNameLength
    this.maxNameLength = (await config()).maxNameLength
    this.minDescriptionLength = (await config()).minDescriptionLength
    this.maxDescriptionLength = (await config()).maxDescriptionLength
  },
  computed: {
    membersUrl(): string {
      return (
        "/group/" + this.$route.params.id + "/member"
      );
    },
    validName(): boolean {
      return (
        this.name.length >= this.minNameLength &&
        this.name.length <= this.maxNameLength
      );
    },
    validDescription(): boolean {
      return (
        this.description.length <= this.maxDescriptionLength &&
        this.description.length >= this.minDescriptionLength
      );
    },
    validFirstExecute(): boolean {
      return new Date(this.firstExecute) > new Date();
    },
    validSchedule(): boolean {
      return this.schedule > 0;
    },
    tomorrow(): Date {
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      return tomorrow;
    },
  },
  methods: {
    async create() {
      if (
        this.receiverId &&
        this.validAmount &&
        this.validName &&
        this.validFirstExecute &&
        this.validSchedule &&
        this.validDescription &&
        !this.loading
      ) {
        this.loading = true;
        const userId = await auth();
        if (userId) {
          try {
            const res = await api.post(
              "/group/" + this.$route.params.id + "/paymentPlan",
              {
                name: this.name,
                description: this.description,
                amount: this.amount,
                receiverId: this.receiverId,
                fromBank: this.fromBank,
                firstPayment: this.firstExecute,
                schedule: this.schedule,
                scheduleUnit: this.scheduleUnit,
              }
            );

            if (!res.data.success) {
              console.error(res.data.message);
            } else {
              this.$router.back();
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
        this.loading = false;
      }
    },
    userSelected(id: string) {
      this.receiverId = id;
    },
    amountChanged(valid: boolean, amount: number) {
      this.validAmount = valid && amount > 0;
      this.amount = amount;
    },
  },
  async mounted() {
    const userId = await auth();
    if (userId) {
      try {
        const res = await api.get("/group/" + this.$route.params.id);
        if (!res.data.success) {
          console.error(res.data.message);
        }
        this.isMember = res.data.member;
        this.isAdmin = res.data.admin;
        this.fromBank = res.data.admin && !res.data.member;
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
});
</script>


<style scoped>
form {
  margin-top: 4vh;
}
.from-bank {
  margin-top: -1vh;
  margin-bottom: 2vh;
}
.schedule-container {
  display: flex;
}
.schedule-input-container {
  display: flex;
  gap: 7px;
}
.schedule-lbl {
  line-height: 35px;
  height: 35px;
  margin-bottom: 3vh;
}
.schedule-lbl-error {
  line-height: 35px;
}
@media screen and (max-height: 640px) {
  .from-bank {
    margin-top: -0.5vh;
  }
}
</style>

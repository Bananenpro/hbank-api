<template>
  <div class="page form-page">
    <h1>{{ $t("payment-plan") }}</h1>
    <form @submit.prevent="update">
      <span class="invalid-form-field-indicator">{{
        validAmount ? "" : "!"
      }}</span
      ><label class="label-next-to-indicator" for="amount">{{
        $t("amount")
      }}</label>
      <MoneyInput @changed="amountChanged" :initialAmount="initialAmount" name="amount" />

      <span class="invalid-form-field-indicator">{{
        validNextExecute ? "" : "!"
      }}</span
      ><label class="label-next-to-indicator" for="next-execute">{{
        $t("payment-plan-update.next-execute")
      }}</label>
      <input
        type="date"
        :min="tomorrow.toISOString().split('T')[0]"
        placeholder="yyyy-mm-dd"
        name="nextExecute"
        v-model="nextExecute"
        id="next-execute"
        @change="changed = true"
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
            @change="changed = true"
          />
          <select
            class="schedule-unit"
            name="schedule-unit"
            v-model="scheduleUnit"
            @change="changed = true"
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
      <input type="text" name="name" v-model="name" id="name" @input="changed = true"/>

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
        @input="changed = true"
      ></textarea>

      <button
        class="btn"
        :disabled="
          !changed ||
          !validAmount ||
          !validName ||
          !validNextExecute ||
          !validSchedule ||
          !validDescription ||
          loading
        "
        type="submit"
      >
        {{ loading ? $t("loading") : $t("update") }}
      </button>
    </form>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import MoneyInput from "@/components/MoneyInput.vue";
import api, { auth } from "@/api";
import { config } from "@/api";

export default defineComponent({
  name: "CreatePaymentPlan",
  components: {
    MoneyInput,
  },
  data() {
    return {
      name: "",
      description: "",
      amount: 0,
      initialAmount: 0,
      validAmount: false,
      loading: false,
      isAdmin: false,
      nextExecute: "",
      schedule: 0,
      scheduleUnit: "",
      changed: false,
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
    validNextExecute(): boolean {
      return new Date(this.nextExecute) > new Date();
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
    async update() {
      if (
        this.validAmount &&
        this.validName &&
        this.validNextExecute &&
        this.validSchedule &&
        this.validDescription &&
        !this.loading
      ) {
        this.loading = true;
        const userId = await auth();
        if (userId) {
          try {
            const res = await api.put(
              "/group/" + this.$route.params.id + "/paymentPlan/" + this.$route.params.paymentPlanId,
              {
                name: this.name,
                description: this.description,
                amount: this.amount,
                nextPayment: this.nextExecute,
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
    amountChanged(valid: boolean, amount: number) {
      this.validAmount = valid && amount > 0;
      this.amount = amount;
      if (this.validAmount && this.amount !== this.initialAmount) {
        this.changed = true
      }
    },
  },
  async mounted() {
    const userId = await auth();
    if (userId) {
      try {
        const res = await api.get("/group/" + this.$route.params.id);
        if (!res.data.success) {
          console.error(res.data.message);
          return
        }
        this.isAdmin = res.data.admin;

        const planRes = await api.get("/group/" + this.$route.params.id + "/paymentPlan/" + this.$route.params.paymentPlanId)
        if (!planRes.data.success) {
          console.error(res.data.message);
          return
        }

        this.nextExecute = new Date(planRes.data.nextExecute*1000).toISOString().split('T')[0]
        this.name = planRes.data.name
        this.description = planRes.data.description ? planRes.data.description : ""
        this.schedule = planRes.data.schedule
        this.scheduleUnit = planRes.data.scheduleUnit
        this.initialAmount = planRes.data.amount
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

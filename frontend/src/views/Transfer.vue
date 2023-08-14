<template>
  <div class="page form-page">
    <h1>{{ $t("transfer") }}</h1>
    <form @submit.prevent="transfer">
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
        validTitle ? "" : "!"
      }}</span
      ><label class="label-next-to-indicator" for="title">{{
        $t("title")
      }}</label>
      <input type="text" name="title" v-model="title" id="title" />

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
        rows="5"
      ></textarea>

      <span v-if="errorText" class="form-error">! {{ errorText }}</span>

      <button
        class="btn"
        :disabled="
          !receiverId ||
          !validAmount ||
          !validTitle ||
          !validDescription ||
          loading
        "
        type="submit"
      >
        {{ loading ? $t("loading") : $t("transfer") }}
      </button>
    </form>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import MoneyInput from "@/components/MoneyInput.vue";
import UserChooser from "@/components/UserChooser.vue";
import api, { auth, config } from "@/api";

export default defineComponent({
  name: "Transfer",
  components: {
    MoneyInput,
    UserChooser,
  },
  data() {
    return {
      title: "",
      description: "",
      amount: 0,
      validAmount: false,
      loading: false,
      receiverId: "",
      fromBank: false,
      isAdmin: false,
      isMember: false,
      errorText: "",
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
        api.defaults.baseURL + "group/" + this.$route.params.id + "/member"
      );
    },
    validTitle(): boolean {
      return (
        this.title.length >= this.minNameLength &&
        this.title.length <= this.maxNameLength
      );
    },
    validDescription(): boolean {
      return (
        this.description.length <= this.maxDescriptionLength &&
        this.description.length >= this.minDescriptionLength
      );
    },
  },
  methods: {
    async transfer() {
      if (
        this.receiverId &&
        this.validAmount &&
        this.validTitle &&
        this.validDescription &&
        !this.loading
      ) {
        this.loading = true;
        const userId = await auth();
        if (userId) {
          try {
            const res = await api.post(
              "/group/" + this.$route.params.id + "/transaction",
              {
                title: this.title,
                description: this.description,
                amount: this.amount,
                receiverId: this.receiverId,
                fromBank: this.fromBank,
              }
            );

            if (!res.data.success) {
              this.errorText = res.data.message;
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
  margin-top: 6vh;
}
.from-bank {
  margin-top: -1vh;
  margin-bottom: 2vh;
}
@media screen and (max-height: 640px) {
  .from-bank {
    margin-top: -0.5vh;
  }
}
</style>

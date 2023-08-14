<template>
  <div class="page form-page">
    <h1>{{ $t("group.create") }}</h1>
    <form @submit.prevent="create">
      <span class="invalid-form-field-indicator">{{
        validName ? "" : "!"
      }}</span
      ><label for="name" class="label-next-to-indicator">{{
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
        rows="7"
      ></textarea>

      <div>
        <input
          type="checkbox"
          name="only-admin"
          v-model="onlyAdmin"
          id="only-admin"
        />
        <label for="only-admin">{{ $t("group.only-admin") }}</label>
      </div>

      <div v-if="errorText" class="form-error-container">
        <span class="form-error">! {{ errorText }}</span>
      </div>

      <button
        type="submit"
        class="btn"
        :disabled="!validName || !validDescription || loading"
      >
        {{ loading ? $t("loading") : $t("create") }}
      </button>
    </form>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
import { config } from "@/api";

export default defineComponent({
  name: "CreateGroup",
  data() {
    return {
      name: "",
      description: "",
      loading: false,
      onlyAdmin: false,
      serverError: "",
      minNameLength: 0,
      maxNameLength: 0,
      minDescLength: 0,
      maxDescLength: 0
    };
  },
  async beforeCreate() {
    this.minNameLength = (await config()).minNameLength
    this.maxNameLength = (await config()).maxNameLength
    this.minDescLength = (await config()).minDescriptionLength
    this.maxDescLength = (await config()).maxDescriptionLength
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
        this.description.length <= this.maxDescLength &&
        this.description.length >= this.minDescLength
      );
    },
    errorText(): string {
      if (this.serverError) {
        return this.serverError;
      }

      if (!this.validName || !this.validDescription) {
        return this.$t("invalid-fields");
      }

      return "";
    },
  },
  methods: {
    async create() {
      this.loading = true;
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.post("/group", {
            name: this.name,
            description: this.description,
            onlyAdmin: this.onlyAdmin,
          });

          if (!res.data.success) {
            this.serverError = res.data.message;
          } else {
            this.$router.push("/group/" + res.data.id);
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
    },
  },
});
</script>

<style scoped>
form {
  margin-top: 10vh;
}
.form-error-container {
  margin-top: 2vh;
}
</style>


<template>
  <div class="page form-page">
    <teleport to="#app">
      <div
        v-if="showSaveDialog"
        class="dialog-bg"
        @click="showSaveDialog = false"
      ></div>
      <div v-if="showSaveDialog" class="dialog">
        <img
          @click="showSaveDialog = false"
          class="dialog-close-btn clickable"
          src="@/assets/close.svg"
          alt="X"
        />
        <h3 class="dialog-title">{{ $t("save") }}</h3>
        <form @submit.prevent="save">
          <span class="invalid-form-field-indicator">{{
            validTitle ? "" : "!"
          }}</span
          ><label for="title" class="label-next-to-indicator">{{
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
            rows="7"
          ></textarea>

          <button
            type="submit"
            class="btn"
            :disabled="!validTitle || !validDescription || loading"
          >
            {{ loading ? $t("loading") : $t("save") }}
          </button>
        </form>
      </div>
    </teleport>

    <div class="total-cash">
      <h2 class="total-cash-lbl">
        {{ $t("dashboard.cash-lbl") }}:
        {{ totalAmount.toFixed(2).replace(".", $t("decimal"))
        }}{{ $t("currency") }}
      </h2>
      <router-link to="/cash/log" class="btn btn-sm view-log-btn">{{
        $t("log")
      }}</router-link>
    </div>
    <div class="cash-units-container">
      <div class="coin-container">
        <div class="cash-item">
          <input
            ref="ct1"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="ct1"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">1ct</span>
        </div>
        <div class="cash-item">
          <input
            ref="ct2"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="ct2"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">2ct</span>
        </div>
        <div class="cash-item">
          <input
            ref="ct5"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="ct5"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">5ct</span>
        </div>
        <div class="cash-item">
          <input
            ref="ct10"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="ct10"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">10ct</span>
        </div>
        <div class="cash-item">
          <input
            ref="ct20"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="ct20"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">20ct</span>
        </div>
        <div class="cash-item">
          <input
            ref="ct50"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="ct50"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">50ct</span>
        </div>
        <div class="cash-item">
          <input
            ref="eur1"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="eur1"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">1{{ $t("currency") }}</span>
        </div>
        <div class="cash-item">
          <input
            ref="eur2"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="eur2"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">2{{ $t("currency") }}</span>
        </div>
      </div>
      <div class="note-container">
        <div class="cash-item">
          <input
            ref="eur5"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="eur5"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">5{{ $t("currency") }}</span>
        </div>
        <div class="cash-item">
          <input
            ref="eur10"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="eur10"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">10{{ $t("currency") }}</span>
        </div>
        <div class="cash-item">
          <input
            ref="eur20"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="eur20"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">20{{ $t("currency") }}</span>
        </div>
        <div class="cash-item">
          <input
            ref="eur50"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="eur50"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">50{{ $t("currency") }}</span>
        </div>
        <div class="cash-item">
          <input
            ref="eur100"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="eur100"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">100{{ $t("currency") }}</span>
        </div>
        <div class="cash-item">
          <input
            ref="eur200"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="eur200"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">200{{ $t("currency") }}</span>
        </div>
        <div class="cash-item">
          <input
            ref="eur500"
            class="cash-amount"
            type="number"
            min="0"
            max="99999"
            v-model="eur500"
          />
          <span class="cash-x-sign">x</span>
          <span class="cash-unit">500{{ $t("currency") }}</span>
        </div>
      </div>
    </div>

    <button
      @click="showSaveDialog = true"
      :disabled="!changed"
      class="btn bottom-btn"
    >
      {{ $t("save") }}
    </button>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
export default defineComponent({
  name: "Cash",
  data() {
    return {
      showSaveDialog: false,
      title: "",
      description: "",
      loading: false,

      totalAmount: 0,
      initCt1: 0,
      initCt2: 0,
      initCt5: 0,
      initCt10: 0,
      initCt20: 0,
      initCt50: 0,
      initEur1: 0,
      initEur2: 0,
      initEur5: 0,
      initEur10: 0,
      initEur20: 0,
      initEur50: 0,
      initEur100: 0,
      initEur200: 0,
      initEur500: 0,

      ct1: 0,
      ct2: 0,
      ct5: 0,
      ct10: 0,
      ct20: 0,
      ct50: 0,
      eur1: 0,
      eur2: 0,
      eur5: 0,
      eur10: 0,
      eur20: 0,
      eur50: 0,
      eur100: 0,
      eur200: 0,
      eur500: 0,
    };
  },
  computed: {
    validTitle(): boolean {
      return this.title.length >= 3 && this.title.length <= 20;
    },
    validDescription(): boolean {
      return this.description.length <= 256;
    },
    changed(): boolean {
      return (
        (this.ct1 !== this.initCt1 ||
          this.ct2 !== this.initCt2 ||
          this.ct5 !== this.initCt5 ||
          this.ct10 !== this.initCt10 ||
          this.ct20 !== this.initCt20 ||
          this.ct50 !== this.initCt50 ||
          this.eur1 !== this.initEur1 ||
          this.eur2 !== this.initEur2 ||
          this.eur5 !== this.initEur5 ||
          this.eur10 !== this.initEur10 ||
          this.eur20 !== this.initEur20 ||
          this.eur50 !== this.initEur50 ||
          this.eur100 !== this.initEur100 ||
          this.eur200 !== this.initEur200 ||
          this.eur500 !== this.initEur500) &&
        (this.$refs.ct1 as HTMLFormElement).checkValidity() &&
        (this.$refs.ct2 as HTMLFormElement).checkValidity() &&
        (this.$refs.ct5 as HTMLFormElement).checkValidity() &&
        (this.$refs.ct10 as HTMLFormElement).checkValidity() &&
        (this.$refs.ct20 as HTMLFormElement).checkValidity() &&
        (this.$refs.ct50 as HTMLFormElement).checkValidity() &&
        (this.$refs.eur1 as HTMLFormElement).checkValidity() &&
        (this.$refs.eur2 as HTMLFormElement).checkValidity() &&
        (this.$refs.eur5 as HTMLFormElement).checkValidity() &&
        (this.$refs.eur10 as HTMLFormElement).checkValidity() &&
        (this.$refs.eur20 as HTMLFormElement).checkValidity() &&
        (this.$refs.eur50 as HTMLFormElement).checkValidity() &&
        (this.$refs.eur100 as HTMLFormElement).checkValidity() &&
        (this.$refs.eur200 as HTMLFormElement).checkValidity() &&
        (this.$refs.eur500 as HTMLFormElement).checkValidity() &&
        this.ct1.toString().length > 0 &&
        this.ct2.toString().length > 0 &&
        this.ct5.toString().length > 0 &&
        this.ct10.toString().length > 0 &&
        this.ct20.toString().length > 0 &&
        this.ct50.toString().length > 0 &&
        this.eur1.toString().length > 0 &&
        this.eur2.toString().length > 0 &&
        this.eur5.toString().length > 0 &&
        this.eur10.toString().length > 0 &&
        this.eur20.toString().length > 0 &&
        this.eur50.toString().length > 0 &&
        this.eur100.toString().length > 0 &&
        this.eur200.toString().length > 0 &&
        this.eur500.toString().length > 0
      );
    },
  },
  methods: {
    async load() {
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.get("/user/cash/current");
          if (!res.data.success) {
            console.error(res.data.message);
            return;
          }

          this.totalAmount = res.data.amount / 100.0;
          this.ct1 = res.data.ct1;
          this.ct2 = res.data.ct2;
          this.ct5 = res.data.ct5;
          this.ct10 = res.data.ct10;
          this.ct20 = res.data.ct20;
          this.ct50 = res.data.ct50;
          this.eur1 = res.data.eur1;
          this.eur2 = res.data.eur2;
          this.eur5 = res.data.eur5;
          this.eur10 = res.data.eur10;
          this.eur20 = res.data.eur20;
          this.eur50 = res.data.eur50;
          this.eur100 = res.data.eur100;
          this.eur200 = res.data.eur200;
          this.eur500 = res.data.eur500;

          this.initCt1 = res.data.ct1;
          this.initCt2 = res.data.ct2;
          this.initCt5 = res.data.ct5;
          this.initCt10 = res.data.ct10;
          this.initCt20 = res.data.ct20;
          this.initCt50 = res.data.ct50;
          this.initEur1 = res.data.eur1;
          this.initEur2 = res.data.eur2;
          this.initEur5 = res.data.eur5;
          this.initEur10 = res.data.eur10;
          this.initEur20 = res.data.eur20;
          this.initEur50 = res.data.eur50;
          this.initEur100 = res.data.eur100;
          this.initEur200 = res.data.eur200;
          this.initEur500 = res.data.eur500;
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
    async save() {
      await auth();

      this.loading = true;

      try {
        const res = await api.post("/user/cash", {
          title: this.title,
          description: this.description,

          ct1: this.ct1,
          ct2: this.ct2,
          ct5: this.ct5,
          ct10: this.ct10,
          ct20: this.ct20,
          ct50: this.ct50,
          eur1: this.eur1,
          eur2: this.eur2,
          eur5: this.eur5,
          eur10: this.eur10,
          eur20: this.eur20,
          eur50: this.eur50,
          eur100: this.eur100,
          eur200: this.eur200,
          eur500: this.eur500,
        });

        if (res.data.success) {
          this.showSaveDialog = false;
          this.load();
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

      this.loading = false;
    },
  },
  async mounted() {
    await this.load();
  },
});
</script>

<style scoped>
.total-cash {
  display: flex;
  justify-content: space-between;
  margin: 2.5vh 0px;
}
.total-cash-lbl {
  margin: 0;
}
.view-log-btn {
  float: right;
}
.cash-units-container {
  display: flex;
  flex-direction: column;
  gap: 1vh;
}

.coin-container,
.note-container {
  display: flex;
  flex-direction: column;
  gap: 1vh;
}
.cash-item {
  display: flex;
}
.cash-amount {
  width: 40%;
  padding: 0px 5px;
  margin-bottom: 0;
}
.cash-x-sign {
  line-height: 27px;
  font-size: 22px;
  width: 10%;
  text-align: center;
}
.cash-unit {
  line-height: 27px;
  font-size: 22px;
}

@media screen and (max-height: 710px) {
  .cash-units-container {
    margin-top: 5vh;
    flex-direction: row;
  }
  .bottom-btn {
    position: absolute !important;
    margin: 0 !important;
  }
}
</style>

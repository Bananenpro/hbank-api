<template>
  <div class="money-input">
    <input @change="done" @blur="done" @focus="focus" type="text" :id="name" :name="name" v-model="amount" autocomplete="off">
    <span class="currency">{{$t("currency")}}</span>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: "MoneyInput",
  data() {
    return {
      amount: "0.00",
    }
  },
  props: {
    name: String,
    max: Number,
    initialAmount: Number
  },
  computed: {
    valid() : boolean {
      return /\d+(?:,\d{1,2})?/.test(this.amount)
    }
  },
  methods: {
    change() {
      this.$emit("changed", this.valid, this.valid ? Math.round(parseFloat(this.amount.replace(this.$t("decimal"), ".")) * 100) : 0)
    },
    done() {
      if (!this.valid) {
        this.amount = "0.00"
      }

      if (this.valid) {
        this.amount = parseFloat(this.amount.replace(this.$t("decimal"), ".")).toFixed(2).replace(".", this.$t("decimal"))
      }
    },
    focus() {
      if (parseFloat(this.amount.replace(this.$t("decimal"), ".")) == 0) {
        this.amount = ""
      }
    }
  },
  watch: {
    amount() {
      this.change()
    },
    initialAmount() {
      if (this.initialAmount) {
        this.amount = (this.initialAmount/100.0).toFixed(2).replace(".", this.$t("decimal"))
        this.done()
      }
    }
  },
  mounted() {
    this.done()
    this.change()
  },
  emits: ["changed"]
})
</script>


<style scoped>
.money-input {
  position: relative;
}
.currency {
  position: absolute;
  top: 9px;
  line-height: 17px;
  right: 8px;
}
</style>

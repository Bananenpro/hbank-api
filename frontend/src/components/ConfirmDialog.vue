<template>
  <teleport to="#app">
  <div v-if="show" class="dialog-bg" @click="$emit('close')"></div>
  <div v-if="show" class="dialog">
    <img @click="$emit('close')" class="dialog-close-btn clickable" src="@/assets/close.svg" alt="X">
    <h3 class="dialog-title">{{$t('are-you-sure')}}</h3>
    <p class="description">{{description}}</p>
    <div class="btns">
      <button class="btn" @click="$emit('yes'); $emit('close')">{{$t('yes')}}</button>
      <button class="btn" @click="$emit('no'); $emit('close')">{{$t('no')}}</button>
    </div>
  </div>
  </teleport>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: "ConfirmDialog",
  props: {
    show: Boolean,
    description: String
  },
  emits: ["yes", "no", "close"]
})
</script>


<style scoped>
.description {
  text-align: center;
  color: var(--color-red);
}
.btns {
  margin-top: 3vh;
  display: flex;
  justify-content: space-evenly;
}
.btn {
  width: 70px;
}
.dialog {
  bottom: auto;
  right: auto;
  width: 80%;
  max-width: 400px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -80%);
  padding-bottom: 3vh;
}
</style>

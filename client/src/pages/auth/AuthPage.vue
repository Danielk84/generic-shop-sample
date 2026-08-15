<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'

const BackBtn = defineAsyncComponent(() => import('@/components/ui/button/BackBtn.vue'))
const identifier = ref<string>('') 
</script>

<template>
  <div class="auth-page c-flex-all-center">
    <div class="auth-box">
      <div class="info">
        <BackBtn
          to="/"
          :icon="{
            strokeColor: '--color-auth-p-icon',
            fillColor: '--color-auth-p-icon',
          }"
        />
        <h1>
          <span>Login or Register</span>
        </h1>
      </div>
      <form class="auth-form">
        <label for="identifier">Please Enter your email address</label>
        <input
          v-model="identifier"
          class="item"
          placeholder="person@email.com"
        />
        <button class="item">Login</button>
      </form>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.auth-page {
  @apply w-screen h-screen;
}

.auth-page > .auth-box {
  @apply w-100 h-120 border-2 p-6 rounded-2xl
    flex flex-col justify-between items-center
    bg-auth-p-bg
    text-auth-p-text
    border-auth-p-border;
}

.auth-box .info {
  @apply w-full h-fit
    flex flex-row justify-between items-center;
}

.auth-box h1 {
  @apply text-2xl font-bold;
}

.auth-form label {
  @apply text-sm font-light p-4;
}

.auth-form .item {
  @apply rounded-xl h-15 w-full text-xl;
}

.auth-form input {
  @apply outline-2 hover:brightness-110
    focus:brightness-110
    p-4 outline-auth-p-outline;
}

.auth-form button {
  @apply my-10 hover:brightness-90 cursor-pointer
    font-bold
    bg-auth-p-btn
    text-auth-p-btn-text;
}
</style>
<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

const BaseBtn = defineAsyncComponent(() => import("@/components/common/button/BaseBtn.vue"))
</script>

<template>
  <form class="login c-flex-all-center">
    <div class="item">
        <label for="username">Username: </label>
        <input type="text" />
    </div>
    <div class="item">
      <label for="password">Password: </label>
      <input type="password" />
    </div>
    <div class="login-btn">
      <BaseBtn :style="{
        backgroundColor: '--color-tab-1',
        outlineWidth: `0px`,
        color: `--color-primary-text`
      }">
        Login
      </BaseBtn>
    </div>
  </form>
</template>

<style scoped>
@reference "@/styles/index.css";

.login {
  @apply flex-col text-xl gap-10 p-50 rounded-4xl
    outline-4 outline-tab-1
    bg-primary-text text-secondary-text;
}

.login .item {
  @apply flex flex-col gap-3;
}

.login .login-btn {
  @apply w-55 h-15;
}

.login .item input {
  @apply rounded-lg outline-4 text-xl py-2 px-4
    outline-tab-1 bg-default-icon focus:brightness-125;
}
</style>

<script setup lang="ts">
import { computed, defineAsyncComponent, ref } from 'vue'

const LoginTab = defineAsyncComponent(() => import('@/components/common/profile/LoginTab.vue'))
const SignUpTab = defineAsyncComponent(() => import('@/components/common/profile/SignUpTab.vue'))

const tabs = [
  { index: 0, label: 'Login', },
  { index: 1, label: 'Sign-Up' },
]

const activeTab = ref(0)

const tabMap = [LoginTab, SignUpTab]

const currentComponent = computed(() => tabMap[activeTab.value])
</script>

<template>
  <div class="login-sign-up c-flex-all-center">
    <div class="tabs c-flex-all-center">
      <button
        v-for="item of tabs"
        :key="item.index"
        @click="activeTab = item.index"
        class="btn c-flex-all-center"
        :class="[
          activeTab === item.index ? `active tab-${item.index+1}` : '',
        ]">
        {{ item.label }}
      </button>
    </div>
    <div class="box">
      <component :is="currentComponent" />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.login-sign-up {
  @apply flex-col gap-10 w-full h-full;
}

.login-sign-up .tabs {
  @apply flex-row
    w-fit h-fit border border-baseline rounded-4xl
    font-bold text-lg text-nowrap;
}

.login-sign-up .btn {
  @apply h-13 w-70;
}

.login-sign-up .active {
  @apply transition-colors delay-250 rounded-4xl;
}

.login-sign-up .tab-1 {
  @apply bg-tab-1 text-primary-text;
}

.login-sign-up .tab-2 {
  @apply bg-tab-2 text-primary-text;
}
</style>

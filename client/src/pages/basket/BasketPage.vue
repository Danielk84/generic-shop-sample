<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

const OrderItems = defineAsyncComponent(() => import('@/pages/basket/OrderItems.vue'))
// const CardTotal = defineAsyncComponent(() => import('@/pages/basket/CartTotal.vue'))
</script>

<template>
  <div class="size-300">
    <OrderItems :orders="[]" />

    <!-- <CardTotal :data="{
      totalItem: 10,
      total: 100,
    }" /> -->
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";


</style>
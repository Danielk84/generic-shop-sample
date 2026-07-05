<script setup lang="ts">
const props = defineProps<{
  to: string,
  name: string,
  price: string,
}>()
</script>

<template>
  <RouterLink :to="props.to">
    <div class="product-v-card">
      <div class="banner"></div>
      <div class="content">
        <h3>{{ props.name }}</h3>
        <span>{{ props.price }}</span>
      </div>
    </div>
  </RouterLink>
</template>

<style scoped>
@reference "@/styles/index.css";

.product-v-card {
  @apply flex flex-col items-center gap-6 p-3 rounded-2xl
    border border-navigation-bar
    hover:shadow-[0px_0px_8px] shadow-navigation-bar
    w-95 h-145;
}

.product-v-card .banner {
  @apply w-98/100 h-9/12 bg-landing-banner-bg rounded-lg;
}

.product-v-card .content {
  @apply flex flex-col items-start justify-between;
}

.product-v-card .content h3 {
  @apply text-2xl font-bold;
}
</style>
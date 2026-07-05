<script setup lang="ts">

</script>

<template>
  <div class="category-card">
    <div class="banner">
    </div>
    <div class="tag">
      <RouterLink to="/">
        <span>NO Content</span>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.category-card {
  @apply flex flex-col items-center gap-4 p-2 rounded-2xl
    border border-navigation-bar
    hover:shadow-[0px_0px_8px] shadow-navigation-bar
    w-60 h-42;
}

.category-card .banner {
  @apply rounded-lg bg-landing-banner-bg
  w-98/100 h-10/12;
}

.category-card .tag {
  @apply font-bold;
}
</style>
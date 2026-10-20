<script setup lang="ts">
const links = [
  { name: 'admin-users-list', label: 'Users' },
  { name: 'admin-categories', label: 'Categories' },
  { name: 'admin-products-list', label: 'Products' },
  { name: 'admin-comments-list', label: 'Comments' },
  { name: 'admin-issues', label: 'Issues' },
  { name: 'admin-orders', label: 'Order Administration' },
  { name: 'admin-shops-list', label: 'Shops' },
  { name: 'orders-list', label: 'My Orders' },
  { name: 'vendor-orders', label: 'Vendor Orders' },
]
</script>

<template>
  <div class="admin-page">
    <h1 class="page-title">Admin Panel</h1>
    <div class="grid">
      <RouterLink
        v-for="link in links"
        :key="link.name"
        class="tile"
        :to="{ name: link.name }"
      >
        <span>{{ link.label }}</span>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.admin-page {
  @apply w-full min-h-screen p-4 md:p-10
    flex flex-col items-center gap-6 md:gap-8;
}

.admin-page .page-title {
  @apply text-2xl md:text-3xl font-bold;
}

.admin-page .grid {
  @apply w-full max-w-200 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 md:gap-6;
}

.admin-page .tile {
  @apply h-28 md:h-32 rounded-2xl
    border border-(--c-v-5)
    flex items-center justify-center text-center
    text-lg md:text-2xl font-bold
    bg-(--c-v-6) dark:bg-(--c-v-4)
    hover:brightness-110
    hover:shadow-[0_0_15px_4px_var(--c-v-9)]
    transition duration-150;
}
</style>

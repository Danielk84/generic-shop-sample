<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'

import type { Category } from '@/contracts/categories/response.interface'
import { useRouter } from 'vue-router'

const CreateTag = defineAsyncComponent(
  () => import('@/pages/panel/admin/categories/CreateTag.vue'),
)
const DeleteTag = defineAsyncComponent(
  () => import('@/pages/panel/admin/categories/DeleteTag.vue'),
)
const CategoriesView = defineAsyncComponent(
  () => import('@/components/common/categories/CategoriesView.vue'),
)

const router = useRouter()

const showFloatingBox = ref<boolean>(false)
const tag = ref<Category>({ tag: '', id: 0 })
const categoryRef = ref<InstanceType<typeof CategoriesView> | null>(null)

function tagHandler(category: Category) {
  tag.value = category
  showFloatingBox.value = true
}
</script>

<template>
  <div class="categories-page">
    <div v-if="showFloatingBox" class="c-floating-window c-flex-all-center">
      <div class="delete-box c-floating-box c-flex-all-center">
        <h2>
          Are you sure for deleting tag
          <span class="text-wrap truncate wrap-break-word max-w-90/100">
            ( {{ tag.tag }} )
          </span>
          <span>?</span>
        </h2>
        <div class="btn-box c-flex-all-center">
          <div class="btn">
            <DeleteTag
              :id="tag.id"
              @after-delete="
                () => {
                  showFloatingBox = false
                  categoryRef?.refetch()
                }
              "
            />
          </div>
          <div class="btn">
            <button
              class="cancel-delete c-flex-all-center"
              @click="showFloatingBox = false"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
    <div class="add c-flex-all-center">
      <CreateTag
        @refetch="
          () => {
            router.go(0)
          }
        "
      />
    </div>
    <div class="list">
      <CategoriesView @tag-handler="tagHandler" ref="categoryRef" />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.categories-page {
  @apply w-screen min-h-screen p-4;
}

.categories-page .delete-box {
  @apply w-120 h-80 flex-col;
}

.floating-box h2 {
  @apply text-center text-wrap text-xl p-4
    text-(--c-v-1);
}

.c-floating-window .btn-box {
  @apply flex-row gap-10;
}

.btn-box .btn {
  @apply w-40 h-15 flex items-center justify-center;
}

.btn-box .cancel-delete {
  @apply size-full
    rounded-2xl hover:brightness-110
    border-4 border-(--c-v-10)
    text-2xl font-bold cursor-pointer;
}

.categories-page .add {
  @apply w-full h-fit p-4;
}

.categories-page .list {
  @apply size-full border-t border-(--c-v-1) my-10;
}
</style>

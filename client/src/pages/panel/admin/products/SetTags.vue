<script setup lang="ts">
import { ref, watch, defineAsyncComponent } from 'vue'
import { useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import icons from '@/utils/icons'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { errorStatusHandler } from '@/utils/helper'
import type { Category } from '@/contracts/categories/response.interface'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)
const TagsListView = defineAsyncComponent(
  () => import('@/components/common/categories/TagsListView.vue'),
)

const props = defineProps<{
  id: string
}>()

const store = useStore()
const router = useRouter()
const notification = useNotificationStore()

const isBoxOpen = ref<boolean>(false)
const tagsList = ref<Category[]>([])
const choosedTags = ref<string[]>([])

watch(
  isBoxOpen,
  (value) => {
    if (value) {
      let tags: string[] = []
      tagsList.value.forEach((t) => {
        tags.push(t.tag)
      })
      choosedTags.value = tags
    }
  },
  { immediate: true },
)

const pcQuery = useQuery({
  queryKey: ['admin-product-tags', props.id],
  queryFn: async () => api.get<string[]>(`categories/pc/${props.id}`),
  select: (res) => res.data,
})

watch(
  pcQuery.error,
  (err) => {
    if (err != null) {
      errorStatusHandler(err, router, {
        notFound() {
          // empty block
          return
        },
      })
    }
  },
  { immediate: true },
)

watch(
  pcQuery.data,
  (data) => {
    if (data === undefined) {
      tagsList.value = []
    } else {
      let tags: Category[] = []
      data.forEach((value, index) => {
        tags.push({ id: index, tag: value })
      })
      tagsList.value = tags
    }
  },
  { immediate: true },
)

const categoriesQuery = useQuery({
  queryKey: ['admin-pc-categories-list'],
  queryFn: async () => api.get<Category[]>('categories/'),
  select: (res) => res.data,
})

watch(
  categoriesQuery.error,
  (err) => {
    if (err != null) {
      errorStatusHandler(err, router, {
        notFound() {
          // empty block
          return
        },
      })
    }
  },
  { immediate: true },
)

const pcMutation = useMutation({
  mutationKey: ['admin-products-set-tags', props.id],
  mutationFn: async (input: string[]) => {
    return api.post(
      `categories/pc/set-tags/${props.id}`,
      {
        tags: input,
      },
      {
        headers: {
          Authorization: store.getAccessToken,
        },
      },
    )
  },
  async onSuccess() {
    notification.success('Product tags updated.')
    await pcQuery.refetch()
  },
  onError(error) {
    errorStatusHandler(error, router, {
      badRequest() {
        notification.error('Invalid tags.')
        return
      },
    })
  },
})
</script>

<template>
  <section class="set-tags c-flex-all-center">
    <button class="edit-btn" @click="isBoxOpen = true">Edit Tags</button>
    <section class="border-t border-(--c-v-15) mt-10 w-full">
      <TagsListView :data="tagsList" @tag-handler="() => {}" />
    </section>
    <section v-if="isBoxOpen" class="c-floating-window c-flex-all-center p-30">
      <div class="c-floating-box max-w-80/100">
        <div class="w-full flex justify-between p-5">
          <button
            class="close-btn"
            @click="
              () => {
                isBoxOpen = false
              }
            "
          >
            <span>X</span>
          </button>
          <button
            class="save-btn c-flex-all-center"
            @click="
              async () => {
                await pcMutation.mutateAsync(choosedTags)
                await pcQuery.refetch()
                isBoxOpen = false
              }
            "
            :disabled="pcMutation.isPending.value"
          >
            <BaseIcon
              :icon="icons.pages.panel.products.save"
              size="32px"
              stroke-color="--c-v-7"
              stroke-dark-color="--c-v-7"
              fill-color="none"
              fill-dark-color="none"
            />
            <span>Save</span>
          </button>
        </div>
        <div class="choosed-box c-flex-all-center">
          <div class="choosed-tags">
            <div v-if="choosedTags.length === 0">
              <h2 class="text-2xl font-bold">There are not any tags.</h2>
            </div>
            <div v-for="item of choosedTags" :key="item" class="choosed-item">
              <span>{{ item }}</span>
            </div>
          </div>
        </div>
        <div class="all-tags">
          <TagsListView
            :data="categoriesQuery.data.value"
            @tag-handler="
              (category: Category) => {
                if (choosedTags.includes(category.tag)) {
                  const index = choosedTags.indexOf(category.tag)
                  if (index > -1) {
                    choosedTags.splice(index, 1)
                  }
                } else {
                  choosedTags.push(category.tag)
                }
              }
            "
          />
        </div>
      </div>
    </section>
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";

.set-tags {
  @apply w-80/100 flex-col
    border-4 border-(--c-v-15)
    p-4 rounded-2xl;
}

.set-tags .edit-btn {
  @apply rounded-2xl cursor-pointer px-10 py-2.5
    bg-(--c-v-15) text-(--c-v-1) text-2xl font-bold
    mt-6 hover:brightness-120;
}

.set-tags .close-btn {
  @apply text-2xl cursor-pointer font-bold;
}

.set-tags .save-btn {
  @apply py-2 px-5 rounded-2xl
    bg-(--c-v-10) text-(--c-v-7) text-2xl cursor-pointer
    gap-4 hover:brightness-110;
}

.set-tags .all-tags {
  @apply max-h-110 overflow-y-scroll overflow-x-hidden;
}

.set-tags .choosed-box {
  @apply h-20 p-10 w-full
    overflow-x-scroll overflow-y-hidden
    border-t-2 border-b-2 border-(--c-v-8);
}

.set-tags .choosed-tags {
  @apply flex items-center justify-start flex-nowrap gap-5 shrink-0;
}

.set-tags .choosed-item {
  @apply rounded-lg p-2 text-lg font-bold
    bg-(--c-v-8) text-(--c-v-1) text-nowrap;
}
</style>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery, useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler, formatDate } from '@/utils/helper'
import type { RelatedCommentResponse } from '@/contracts/comments/response.interface'

const store = useStore()
const route = useRoute()
const router = useRouter()

const page = ref<number | null>(null)
const maxPage = ref<number>(1)

const deleteID = ref<string | null>(null)
const overviewItem = ref<RelatedCommentResponse | null>(null)

const listQuery = useQuery({
  queryKey: ['admin-commnets-full-list', page.value],
  enabled: computed(() => page.value !== null),
  queryFn: async () =>
    api.get<RelatedCommentResponse[]>('comments/full', {
      params: {
        page: page.value,
      },
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  select: (res) => {
    const mp = Number(res.headers['x-max-page'])
    if (mp !== undefined) {
      maxPage.value = mp
    }
    return res.data
  },
})

watch(listQuery.error, (err) => {
  errorStatusHandler(err, router, {
    notFound() {
      const p = route.query?.page
      if (p !== null && Number(p) > 1) {
        router.back()
      }
    },
  })
})

const setActiveMutate = useMutation({
  mutationKey: ['admin-comments-set-active'],
  mutationFn: async (input: { id: string; status: boolean }) => {
    return api.put(
      `comments/set-active/${input.id}`,
      { accepted: input.status },
      {
        headers: {
          Authorization: store.getAccessToken,
        },
      },
    )
  },
  async onSuccess() {
    await listQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router)
  },
})

const deleteMutate = useMutation({
  mutationKey: ['admin-comments-delete'],
  mutationFn: async (input: string) => {
    return api.delete(`/comments/${input}`, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  async onSuccess() {
    await listQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router)
  },
})
</script>

<template>
  <div class="comments-list-page">
    <section v-if="overviewItem != null" class="c-floating-box">
      <div class="overview-box c-floating-window c-flex-all-center">
        <div class="w-full">
          <button
            @click="
              () => {
                overviewItem = null
              }
            "
            class="cursor-pointer text-2xl"
          >
            X
          </button>
        </div>
        <div class="id-list">
          <h3>ID: {{ overviewItem.id }}</h3>
          <p>Parent: {{ overviewItem.parent }}</p>
          <p>Referrer: {{ overviewItem.referrer }}</p>
        </div>
        <div class="intro">
          <p>{{ overviewItem.name }}</p>
          <p>{{ formatDate(overviewItem.pub_date) }}</p>
          <p>Children amount: {{ overviewItem.children_amount }}</p>
          <p>is_active: {{ overviewItem.is_active }}</p>
        </div>
        <div class="content-body">
          {{ overviewItem.body }}
        </div>
      </div>
    </section>
    <section v-if="deleteID != null" class="c-floating-box">
      <div class="delete-box c-floating-window c-flex-all-center">
        <h2>Are you sure to delete?</h2>
        <p>ID: ( {{ deleteID }} )</p>
        <div>
          <button
            @click="
              async () => {
                await deleteMutate.mutateAsync(deleteID as string)
                deleteID = null
              }
            "
            class="base-btn delete-btn"
            :class="{ 'c-is-pending': !deleteMutate.isPending }"
          >
            Delete
          </button>
          <button
            @click="
              () => {
                deleteID = null
              }
            "
            class="base-btn cancel-btn"
          >
            Cancel
          </button>
        </div>
      </div>
    </section>
    <section class="list c-flex-all-center">
      <div
        v-for="item of listQuery.data.value"
        :key="item.id"
        class="item c-flex-all-center"
      >
        <p>{{ item.id }}</p>
        <p>{{ item.name }}</p>
        <p>{{ item.children_amount }}</p>
        <p>{{ item.pub_date }}</p>
        <button
          @click="
            () => {
              setActiveMutate.mutate({
                id: item.id,
                status: !item.is_active,
              })
            }
          "
          class="base-btn active-btn"
          :class="{
            'off-btn': !item.is_active,
            'c-is-pending': !setActiveMutate.isPending,
          }"
        >
          Active
        </button>
        <button
          @click="
            () => {
              deleteID = item.id
            }
          "
          class="base-btn delete-btn"
        >
          Delete
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.comments-list-page {
  @apply w-screen min-h-screen;
}

.comments-list-page .overview {
  @apply w-60/100 min-h-20 max-h-80/100 overflow-y-auto;
}

.comments-list-page .id-list {
  @apply text-lg border-b border-admin-comments-list-border;
}

.comments-list-page .intro {
  @apply border-b border-admin-comments-list-border;
}

.comments-list-page .content-body {
  @apply text-sm font-bold text-center w-full;
}

.comments-list-page .delete-box {
  @apply w-140 h-90 flex-col;
}

.comments-list-page .list {
  @apply w-full h-fit flex-col gap-10;
}

.comments-list-page .item {
  @apply w-80/100 h-50
    shadow hover:shadow-[0px_0px_20px_5px]
    shadow-admin-comments-list-shadow
    rounded-2xl p-4;
}

.comments-list-page .base-btn {
  @apply rounded-2xl w-100 h-18
    hover:brightness-95
    flex justify-center items-center;
}

.comments-list-page .active-btn {
  @apply bg-admin-comments-list-active-bt;
}

.comments-list-page .off-btn {
  @apply bg-admin-comments-list-off-btn;
}

.comments-list-page .delete-btn {
  @apply bg-admin-comments-list-delete-btn;
}

.comments-list-page .cancel-btn {
  @apply border-2 border-admin-comments-list-cancel-border;
}
</style>

<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery, useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { errorStatusHandler, formatDate } from '@/utils/helper'
import type { RelatedCommentResponse } from '@/contracts/comments/response.interface'

const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const store = useStore()
const notification = useNotificationStore()
const route = useRoute()
const router = useRouter()

const page = ref<number | null>(null)
const maxPage = ref<number>(1)
const finderID = ref<string>('')
const foundComments = ref<RelatedCommentResponse[] | null>(null)

const deleteID = ref<string | null>(null)
const overviewItem = ref<RelatedCommentResponse | null>(null)

const listQuery = useQuery({
  queryKey: computed(() => ['admin-comments-full-list', page.value]),
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
    if (!Number.isNaN(mp) && mp > 0) {
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

const findMutation = useMutation({
  mutationKey: ['admin-comments-find'],
  mutationFn: async (id: string) => {
    return api.get<RelatedCommentResponse[]>(`comments/find/${id}`, {
      params: { page: page.value },
      headers: { Authorization: store.getAccessToken },
    })
  },
  onSuccess: (res) => {
    foundComments.value = res.data
    const mp = Number(res.headers['x-max-page'])
    if (!Number.isNaN(mp) && mp > 0) {
      maxPage.value = mp
    }
  },
  onError: (err) => {
    foundComments.value = null
    errorStatusHandler(err, router)
  },
})

function findComments() {
  const id = finderID.value.trim()
  if (id !== '') {
    findMutation.mutate(id)
  }
}

const overviewMutation = useMutation({
  mutationKey: ['admin-comment-overview'],
  mutationFn: async (id: string) => {
    return api.get<RelatedCommentResponse>(`comments/overview/${id}`, {
      headers: { Authorization: store.getAccessToken },
    })
  },
  onSuccess: (res) => {
    overviewItem.value = res.data
  },
  onError: (err) => {
    errorStatusHandler(err, router)
  },
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
    notification.success('Saved comment activation status.')
    await listQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router)
  },
})

const cleanCacheMutation = useMutation({
  mutationKey: ['clean-comments-cache'],
  mutationFn: async () =>
    api.delete('comments/clean-cache', {
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  async onSuccess() {
    notification.success('Comments cache cleand.')
  },
  onError(error) {
    errorStatusHandler(error, router, {
      notAcceptable() {
        notification.error('Failed to clean cache.')
        return
      },
    })
  },
})

const deleteMutate = useMutation({
  mutationKey: ['admin-comments-delete'],
  mutationFn: async (input: string) => {
    return api.delete(`comments/${input}`, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  async onSuccess() {
    notification.success('Comments deleted.')
    await listQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router)
  },
})
</script>

<template>
  <div class="comments-list-page">
    <section
      v-if="overviewItem != null"
      class="c-floating-window c-flex-all-center"
    >
      <div class="overview-box c-floating-box c-flex-all-center flex-col">
        <div class="w-full">
          <button
            @click="
              () => {
                overviewItem = null
              }
            "
            class="cursor-pointer text-2xl font-bold"
          >
            <span>X</span>
          </button>
        </div>
        <div class="id-list">
          <h3>ID: {{ overviewItem.id }}</h3>
          <p>Parent: {{ overviewItem.parent }}</p>
          <p>Referrer: {{ overviewItem.referrer }}</p>
        </div>
        <div class="intro">
          <p>Name: {{ overviewItem.name }}</p>
          <p>Pub_Date: {{ formatDate(overviewItem.pub_date) }}</p>
          <p>Children amount: {{ overviewItem.children_amount }}</p>
          <p>is_active: {{ overviewItem.is_active }}</p>
        </div>
        <div class="content-body">
          {{ overviewItem.body }}
        </div>
      </div>
    </section>
    <section
      v-if="deleteID != null"
      class="c-floating-window c-flex-all-center"
    >
      <div class="delete-box c-floating-box c-flex-all-center">
        <div>
          <h2>Are you sure to delete?</h2>
          <p>ID: ( {{ deleteID }} )</p>
        </div>
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
    <section class="w-full c-flex-all-center mt-2">
      <button
        class="clean-cache"
        @click="
          () => {
            cleanCacheMutation.mutate()
          }
        "
        :disabled="cleanCacheMutation.isPending.value"
      >
        Clean Cache
      </button>
    </section>

    <form class="finder c-flex-all-center" @submit.prevent="findComments">
      <label for="comment-id">Find comment</label>
      <input
        class="c-clean-input"
        id="comment-id"
        v-model="finderID"
        type="text"
      />
      <button type="submit" :disabled="findMutation.isPending.value">
        Search
      </button>
      <button
        v-if="foundComments !== null"
        type="button"
        @click="foundComments = null"
      >
        Clear
      </button>
    </form>
    <section class="list c-flex-all-center">
      <div
        v-for="item of foundComments ?? listQuery.data.value ?? []"
        :key="item.id"
        class="item c-flex-all-center"
      >
        <div
          class="border-b border-(--c-v-9) p-4 flex flex-col gap-2 w-full justify-center text-center"
        >
          <p>id: {{ item.id }}</p>
          <p>name: {{ item.name }}</p>
          <p>children_amount: {{ item.children_amount }}</p>
          <p>pub_date: {{ item.pub_date }}</p>
        </div>
        <button
          class="base-btn border-4 border-(--c-v-10)"
          @click="overviewMutation.mutate(item.id)"
        >
          Overview
        </button>
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
    <section class="pagination c-flex-all-center">
      <ListPagination
        :last="maxPage"
        page-name="admin-comments-list"
        @change-page="
          (newValue: number) => {
            page = newValue
          }
        "
      />
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
  @apply w-full text-lg border-b border-(--c-v-15);
}

.comments-list-page .intro {
  @apply w-full border-b border-(--c-v-15);
}

.comments-list-page .content-body {
  @apply text-sm font-bold text-center w-150 max-h-50
    overflow-y-scroll overflow-x-scroll;
}

.comments-list-page .clean-cache {
  @apply rounded-2xl p-4 text-2xl font-bold
    hover:brightness-110 text-(--c-v-1) bg-(--c-v-13);
}

.comments-list-page .delete-box {
  @apply w-140 h-90 flex-col p-8;
}

.comments-list-page .finder {
  @apply w-full gap-3 p-6 border-b border-(--c-v-15) mb-10;
}

.comments-list-page .finder label {
  @apply font-bold;
}

.comments-list-page .finder input {
  @apply min-w-0 flex-1 p-2 rounded-lg border-2
    border-(--c-v-1) dark:border-(--c-v-7)
    bg-transparent;
}

.comments-list-page .finder button {
  @apply p-2 rounded-lg font-bold cursor-pointer
    bg-(--c-v-8) text-(--c-v-1) disabled:opacity-50;
}

.comments-list-page .list {
  @apply w-full h-fit flex-col gap-10;
}

.comments-list-page .item {
  @apply w-60/100 h-fit flex-col
    shadow-[0_1px_3px_0_var(--c-v-10),0_1px_2px_-1px_var(--c-v-10)]
    hover:shadow-[0px_0px_20px_5px_var(--c-v-10)]
    rounded-2xl p-4;
}

.comments-list-page .base-btn {
  @apply rounded-2xl w-90 h-13 mt-5
    hover:brightness-110 text-2xl font-bold
    flex justify-center items-center;
}

.comments-list-page .active-btn {
  @apply bg-(--c-v-14) text-(--c-v-1);
}

.comments-list-page .off-btn {
  @apply bg-(--c-v-11) text-(--c-v-7);
}

.comments-list-page .delete-btn {
  @apply bg-(--c-v-12) text-(--c-v-7);
}

.comments-list-page .cancel-btn {
  @apply border-2 border-(--c-v-9);
}

.comments-list-page .pagination {
  @apply w-full py-6;
}
</style>

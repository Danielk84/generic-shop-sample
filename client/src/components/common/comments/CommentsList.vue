<script setup lang="ts">
import { computed, ref, defineAsyncComponent, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { errorStatusHandler } from '@/utils/helper'
import { formatDate } from '@/utils/helper'
import { useValidator } from '@/utils/validator'
import {
  CommentRequest,
  type CommentInput,
  type RelatedCommentsRequest,
} from '@/contracts/comments/request.schema'
import type { CommentResponse } from '@/contracts/comments/response.interface'

const IndentedComments = defineAsyncComponent(
  () => import('@/components/common/comments/IndentedComments.vue'),
)
const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const props = withDefaults(
  defineProps<{
    isParent?: boolean
    pageName?: string
    relation: RelatedCommentsRequest
  }>(),
  {
    isParent: true,
    pageName: 'product',
  },
)

const store = useStore()
const router = useRouter()
const notification = useNotificationStore()

const { formData, errors, validate } = useValidator(CommentRequest)

const createTarget = ref<RelatedCommentsRequest | null>(null)
const isError = ref<boolean>(false)
const page = ref(1)
const maxPage = ref(1)

const { data, error, refetch } = useQuery({
  queryKey: computed(() => [
    'comments-list',
    props.relation.parent,
    props.relation.referrer,
    page.value,
  ]),
  queryFn: async () => {
    const params: { parent?: string; referrer: string; page: number } = {
      referrer: props.relation.referrer,
      page: page.value,
    }
    if (props.relation.parent) {
      params.parent = props.relation.parent
    }
    return api.get<CommentResponse[]>('comments/', { params })
  },
  select: (res) => {
    const mp = Number(res.headers['x-max-page'])
    if (!Number.isNaN(mp) && mp > 0) {
      maxPage.value = mp
    }
    return res.data
  },
})

watch(error, (err) => {
  errorStatusHandler(err, router, {
    notFound() {
      // empty block
      return
    },
  })
})

const createMutate = useMutation({
  mutationFn: async (input: CommentInput) =>
    api.post('comments/', input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  async onSuccess() {
    formData.value.body = ''
    notification.success('Comment added.')
    createTarget.value = null
    isError.value = false
    await refetch()
  },
  onError(err) {
    errorStatusHandler(err, router, {
      notFound() {
        notification.error('Unable to add comment.')
        return
      },
      badRequest() {
        notification.error('Invalid comment body')
      },
    })
    isError.value = true
  },
})

async function createComments(event: MouseEvent) {
  event.preventDefault()

  formData.value.referrer = props.relation.referrer
  formData.value.parent = createTarget.value?.parent || undefined
  const { input, isValid } = await validate()
  if (!isValid) {
    notification.error('Invalid comment body.')
    isError.value = true
  }
  if (input.data !== undefined) {
    createMutate.mutate(input.data)
  }
}
</script>

<template>
  <section class="comments-list">
    <div class="w-full flex justify-center">
      <button
        v-if="props.isParent"
        class="create-btn"
        @click="
          () => {
            createTarget = {
              parent: '',
              referrer: props.relation.referrer,
            }
          }
        "
      >
        Send Your experience
      </button>
    </div>
    <div v-if="data === undefined" class="empty-comments">
      <span>There are not any comments.</span>
    </div>
    <div v-else class="mt-8">
      <div class="list">
        <div v-for="item of data" :key="item.id" class="item">
          <div class="content">
            <p class="name">{{ item.name }}</p>
            <p class="date">{{ formatDate(item.pub_date) }}</p>
            <p class="body">{{ item.body }}</p>
          </div>
          <button
            @click="
              () => {
                createTarget = {
                  parent: item.id,
                  referrer: props.relation.referrer,
                }
              }
            "
            class="answer-btn"
          >
            Answer
          </button>
          <div>
            <IndentedComments
              :relation="{
                parent: item.id,
                referrer: props.relation.referrer,
              }"
              :children-amount="item.children_amount"
            />
          </div>
        </div>
      </div>
      <ListPagination
        :last="maxPage"
        :page-name="props.pageName"
        @change-page="(value: number) => (page = value)"
      />
    </div>
    <section
      v-if="createTarget != null"
      class="c-floating-window c-flex-all-center top-0 left-0"
    >
      <form
        class="comment-box c-form c-form-bg justify-between"
        :class="{ 'c-form-error-shadow': isError }"
      >
        <div>
          <button
            class="text-2xl font-bold cursor-pointer"
            type="button"
            @click="createTarget = null"
          >
            <span>X</span>
          </button>
        </div>
        <label class="c-form-label" for="body">Your experience:</label>
        <textarea
          class="c-form-input"
          v-model="(formData as CommentInput).body"
          rows="4"
          cols="5"
        ></textarea>
        <p class="c-form-error" v-if="errors['body'] != undefined">
          {{ errors['body'] }}
        </p>
        <button @click="createComments($event)" class="c-form-btn mb-0">
          Send
        </button>
      </form>
    </section>
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";

.comments-list {
  @apply w-full h-fit;
}

.comments-list .comment-box {
  @apply rounded-2xl p-4 size-120;
}

.comments-list textarea {
  @apply resize-none h-60;
}

.comments-list .create-btn {
  @apply rounded-2xl p-4 bg-(--c-v-13) text-(--c-v-1) text-2xl font-bold size-fit;
}

.comments-list .empty-comments {
  @apply w-full p-5 text-xl text-center;
}

.comments-list .list {
  @apply border-l border-t p-4;
}

.comments-list .item {
  @apply border-b;
}

.comments-list .name {
  @apply text-xl w-full flex items-center justify-start;
}

.comments-list .date {
  @apply text-sm w-full flex items-center justify-start;
}

.comments-list .body {
  @apply text-wrap text-start w-full h-fit p-5;
}
</style>

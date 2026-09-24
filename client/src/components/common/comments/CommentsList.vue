<script setup lang="ts">
import { ref, defineAsyncComponent, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
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

const props = withDefaults(
  defineProps<{
    isParent?: boolean
    relation: RelatedCommentsRequest
  }>(),
  {
    isParent: true,
  },
)

const store = useStore()
const router = useRouter()

const { formData, errors, validate } = useValidator(CommentRequest)

const createTarget = ref<RelatedCommentsRequest | null>(null)
const errorMsg = ref<string | null>(null)

const { data, error } = useQuery({
  queryKey: ['comments-list', props.relation.parent, props.relation.referrer],
  queryFn: async () =>
    api.get<CommentResponse[]>(
      `commnets/?parent=${props.relation.parent}&referrer=${props.relation.referrer}`,
    ),
  select: (res) => res.data,
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
  onSuccess() {
    errorMsg.value = null
  },
  onError() {
    errorMsg.value = 'invalid comments'
  },
})

async function createComments(event: MouseEvent) {
  event.preventDefault()

  const { input, isValid } = await validate()
  if (!isValid) {
    errorMsg.value = 'invalid comments'
  }
  if (input.data !== undefined) {
    createMutate.mutate(input.data)
  }
}
</script>

<template>
  <section class="comments-list">
    <section v-if="createTarget != null" class="c-floating-window">
      <form
        class="c-form c-form-bg"
        :class="{ 'c-form-error-shadow': errorMsg != null }"
      >
        <div class="c-form-item">
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
        </div>
        <button @click="createComments($event)" class="c-form-btn">Send</button>
        <p class="c-form-error" v-if="errorMsg !== null">
          {{ errorMsg }}
        </p>
      </form>
    </section>
    <div class="create-btn">
      <button
        v-if="props.isParent"
        @click="
          () => {
            createTarget = {
              parent: '',
              referrer: props.relation.referrer,
            }
          }
        "
        class="c-form-btn"
      >
        Send Your experience
      </button>
    </div>
    <div v-if="data === undefined" class="empty-comments">
      <span>There are not any comments.</span>
    </div>
    <div v-else class="c-flex-all-center flex-col">
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
    </div>
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";

.comments-list {
  @apply w-full h-fit;
}

.comments-list textarea {
  @apply resize-y;
}

.comments-list .empty-comments {
  @apply w-full p-5 text-xl text-center;
}

.comments-list .list {
  @apply border-r border-b p-4;
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
  @apply text-wrap text-center w-full h-fit;
}
</style>

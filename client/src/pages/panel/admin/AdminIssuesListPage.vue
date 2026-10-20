<script setup lang="ts">
import { ref, computed, watch, defineAsyncComponent } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useQuery, useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { useValidator } from '@/utils/validator'
import { errorStatusHandler } from '@/utils/helper'
import {
  SetIssuesResRequest,
  type SetIssuesResInput,
} from '@/contracts/issues/request.schema'
import type {
  IssuesResponse,
  IssuesSummaryResponse,
} from '@/contracts/issues/response.interface'

const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const store = useStore()
const notification = useNotificationStore()
const router = useRouter()
const route = useRoute()

const page = ref<number>(1)
const maxPage = ref<number>(1)
const getIssueDetail = ref<{
  id: string
  user_id: string
} | null>(null)

const {
  formData: SetResForm,
  errors: SetResErrors,
  validate: SetResValidator,
} = useValidator(SetIssuesResRequest)

const issuesQuery = useQuery({
  queryKey: computed(() => ['issues-list', page.value]),
  queryFn: async () =>
    api.get<IssuesSummaryResponse[]>('issues/admin', {
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

watch(issuesQuery.error, (err) => {
  if (err !== undefined) {
    errorStatusHandler(err, router, {
      notFound() {
        return
      },
    })
  }
})

const getIssueQuery = useQuery({
  queryKey: computed(() => ['get-issue-query', getIssueDetail]),
  enabled: computed(() => getIssueDetail.value !== null),
  queryFn: async () =>
    api.get<IssuesResponse>(
      `issues/${getIssueDetail.value?.id}/${getIssueDetail.value?.user_id}`,
      {
        headers: {
          Authorization: store.getAccessToken,
        },
      },
    ),
  select(res) {
    return res?.data
  },
})

watch(getIssueQuery.error, (err) => {
  if (err !== undefined) {
    errorStatusHandler(err, router, {
      notFound() {
        return
      },
    })
  }
})

watch(getIssueQuery.data, (value) => {
  SetResForm.value.id = value?.id
  SetResForm.value.req = value?.req
  SetResForm.value.is_done = value?.is_done
})

const SetResMutation = useMutation({
  mutationKey: ['set-issues-res'],
  mutationFn: async (input: SetIssuesResInput) =>
    api.put('/issues/set-response', input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  async onSuccess() {
    notification.success('changes saved.')
    await issuesQuery.refetch()
  },
  onError(error) {
    errorStatusHandler(error, router, {
      badRequest() {
        notification.error('invalid response.')
        return
      },
    })
  },
})

async function saveResponse() {
  const { input, isValid } = await SetResValidator()
  if (!isValid) {
    notification.error('Invalid response.')
    return
  }
  if (input.data === undefined) {
    notification.error('Unexpected error')
    return
  }
  await SetResMutation.mutateAsync(input.data)
}
</script>

<template>
  <section class="issues-list-page">
    <section
      v-if="
        issuesQuery.data.value === undefined ||
        issuesQuery.data.value?.length === 0
      "
      class="issues-empty-list"
    >
      <h1>Empty List</h1>
    </section>
    <section class="issues-box">
      <div class="list c-flex-all-center">
        <button
          type="button"
          v-for="item of issuesQuery.data.value"
          :key="item.id"
          class="item-btn"
          @click="
            () => {
              getIssueDetail = {
                id: item.id,
                user_id: item.user_id,
              }
            }
          "
        >
          <p class="content">ID: {{ item.id }}</p>
          <p class="is-done" :class="[item.is_done ? 'done' : 'not-done']">
            Is Done
          </p>
        </button>
      </div>
      <div>
        <ListPagination
          :last="maxPage"
          :page-name="String(route.name)"
          @change-page="
            (newValue: number) => {
              page = newValue
            }
          "
        />
      </div>
    </section>
    <section
      v-if="getIssueDetail !== null"
      class="c-floating-window c-flex-all-center"
    >
      <div
        class="c-floating-box c-form-bg max-h-90/100 overflow-x-hidden overflow-y-scroll"
      >
        <div class="w-full p-2">
          <button
            type="button"
            class="text-2xl font-bold cursor-pointer"
            @click="
              () => {
                getIssueDetail = null
              }
            "
          >
            <span>X</span>
          </button>
        </div>
        <div
          v-if="getIssueQuery.data.value === undefined"
          class="issues-empty-list"
        >
          <h2>Not Founded</h2>
        </div>
        <form v-else class="c-form gap-5">
          <p class="content">ID: {{ getIssueQuery.data.value?.id }}</p>
          <p class="content">
            USER_ID: {{ getIssueQuery.data.value?.user_id }}
          </p>
          <label class="c-form-label">Req</label>
          <p class="body-text">
            {{ getIssueQuery.data.value?.req }}
          </p>
          <div>
            <label class="c-form-label" for="res">Res</label>
            <textarea
              v-model="(SetResForm as SetIssuesResInput).res"
            ></textarea>
            <p class="c-form-error" v-if="SetResErrors['res']">
              {{ SetResErrors['res'] }}
            </p>
          </div>
          <div class="flex gap-4">
            <label class="c-form-label" for="is_done">Is Done</label>
            <input
              v-model="(SetResForm as SetIssuesResInput).is_done"
              type="checkbox"
            />
          </div>
          <button
            class="c-form-btn"
            :disabled="SetResMutation.isPending.value"
            type="button"
            @click="saveResponse"
          >
            Save
          </button>
        </form>
      </div>
    </section>
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";

.issues-list-page {
  @apply w-full min-h-screen flex items-center flex-col;
}

.issues-empty-list {
  @apply w-full text-2xl font-bold text-(--c-v-11)
    flex items-center justify-center p-4;
}

.issues-list-page .create-btn {
  @apply rounded-2xl text-(--c-v-7) bg-(--c-v-15) hover:brightness-120
    text-lg md:text-2xl font-bold p-4 md:px-8
    cursor-pointer;
}

.issues-list-page .issues-box {
  @apply w-full mt-5;
}

.issues-list-page .list {
  @apply flex-col gap-5 w-full h-fit;
}

.issues-list-page textarea {
  @apply outline-4
    focus:brightness-120
    p-4 outline-(--c-v-7)
    rounded-xl h-60 w-full text-xl
    resize-none;
}

.issues-list-page .item-btn {
  @apply border-4 border-(--c-v-15) rounded-2xl
    p-2 max-[454px]:max-w-95/100 h-fit
    flex justify-center items-center flex-col gap-4
    cursor-pointer hover:border-(--c-v-10);
}

.issues-list-page .content {
  @apply bg-(--c-v-0) text-(--c-v-7) w-full
    p-2 text-start rounded-2xl text-wrap wrap-break-word;
}

.issues-list-page .is-done {
  @apply text-sm rounded-xl py-2 px-4;
}

.issues-list-page .done {
  @apply bg-(--c-v-13) text-(--c-v-1);
}

.issues-list-page .is-done {
  @apply bg-(--c-v-11) text-(--c-v-7);
}

.issues-list-page .body-text {
  @apply border-4 p-2 rounded-2xl text-xl font-bold text-wrap wrap-break-word;
}
</style>

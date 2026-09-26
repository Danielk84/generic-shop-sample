import * as z from 'zod'

export const CommentRequest = z.object({
  parent: z.uuid().optional(),
  referrer: z.uuid(),
  body: z.string().min(1).max(5000),
})
export type CommentInput = z.infer<typeof CommentRequest>

export interface RelatedCommentsRequest {
  parent?: string
  referrer: string
}

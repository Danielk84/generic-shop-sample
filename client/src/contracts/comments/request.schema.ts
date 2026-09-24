import * as z from 'zod'

export const CommentRequest = z.object({
  parent: z.uuid().optional(),
  referrer: z.uuid(),
  body: z.string(),
})
export type CommentInput = z.infer<typeof CommentRequest>

export interface RelatedCommentsRequest {
  parent?: string
  referrer: string
}

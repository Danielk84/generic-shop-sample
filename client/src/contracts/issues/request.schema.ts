import * as z from 'zod'

export const CreateIssuesRequest = z.object({
  user_id: z.uuid(),
  req: z.string().min(4).max(5000),
})
export type CreateIssuesInput = z.infer<typeof CreateIssuesRequest>

export const UpdateIssuesRequest = z
  .object({
    id: z.uuid(),
  })
  .extend(CreateIssuesRequest.shape)
export type UpdateIssuesInput = z.infer<typeof UpdateIssuesRequest>

export const SetIssuesResRequest = z.object({
  id: z.uuid(),
  res: z.string().min(1).max(5000),
  is_done: z.boolean(),
})
export type SetIssuesResInput = z.infer<typeof SetIssuesResRequest>

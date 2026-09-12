import * as z from 'zod'

export const CategoryTag = z.object({
  tag: z.string().min(1),
})
export type CategoryTagInput = z.infer<typeof CategoryTag>

export const PC = z.object({
  tags: z.array(z.string().min(1)),
})
export type PCInput = z.infer<typeof PC>

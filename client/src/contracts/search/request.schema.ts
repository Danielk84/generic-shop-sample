import * as z from 'zod'

export const SearchRequest = z.object({
  query_str: z.string().min(1).max(500),
})
export type SearchInput = z.infer<typeof SearchRequest>

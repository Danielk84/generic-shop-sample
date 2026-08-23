import * as z from 'zod'

import { EmailAddrRequest } from '@/contracts/users/request.schema'

const PassKeyRequest = z.object({
  pass_key: z.string().max(8)
})

export const LoginRequest = z
  .object({})
  .extend(EmailAddrRequest)
  .extend(PassKeyRequest)
export type LoginInput = z.infer<typeof LoginRequest>

  
export const RegisterRequest = z
  .object({})
  .extend(EmailAddrRequest)
  .extend(PassKeyRequest)
export type RegisterInput = z.infer<typeof RegisterRequest>
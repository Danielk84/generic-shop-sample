import * as z from 'zod'

import {
  EmailAddrRequest,
  RegisterUserRequest,
} from '@/contracts/users/request.schema'

const PassKeyRequest = z.object({
  pass_key: z.string().max(8),
})

export const LoginRequest = z
  .object({})
  .extend(EmailAddrRequest.shape)
  .extend(PassKeyRequest.shape)
export type LoginInput = z.infer<typeof LoginRequest>

export const RegisterRequest = z
  .object({})
  .extend(PassKeyRequest.shape)
  .extend(RegisterUserRequest.shape)
export type RegisterInput = z.infer<typeof RegisterRequest>

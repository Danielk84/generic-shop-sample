import * as z from 'zod'

// user model
export const PermissionType = {
  Admin: 0,
  Vendor: 1,
  Customer: 3,
  BlockUser: 4,
} as const

export const EmailAddrRequest = z.object({
  email: z.email().min(10).max(254)
})
export type EmailAddrInput = z.infer<typeof EmailAddrRequest>

export const PhoneNumberRequest = z.object({
  phone_number: z.string(),
})
export type PhoneNumberInput = z.infer<typeof PhoneNumberRequest>

export const PermissionRequest = z.object({
  permission_type: z.number().gte(0).lt(4),
})

export const UserPermissionRequest = z
  .object({
    is_active: z.boolean(),
  })
  .extend(PermissionRequest)
export type UserPermissionInput = z.infer<typeof UserPermissionRequest>

export const CreateUserRequest = z
  .object()
  .extend(EmailAddrRequest)
  .extend(UserPermissionRequest)
export type CreateUserInput = z.infer<typeof CreateUserRequest>

export const UserInfoRequest = z.object({
  first_name: z.string().min(1).max(50),
  last_name: z.string().min(1).max(60),
  national_code: z.string().max(10),
})
export type UserInfoInput = z.infer<typeof UserInfoRequest>

// shop model
export const ShopPhoneNumberRequest = z.object({
  phone_number: z.string().max(15)
})
export type ShopPhoneNumberInput = z.infer<typeof ShopPhoneNumberRequest>

export const UpsertShopRequest = z.object({
  brand: z.string().min(2).max(100),
  shop_addr: z.string().max(1000),
  zip_code: z.string().max(10),
  business_code: z.string().max(100),
  bio: z.string().max(650),
})
export type UpsertShopInput = z.infer<typeof UpsertShopRequest>

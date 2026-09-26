import * as z from 'zod'

import { ProductProperty } from '../products/request.schema'

export const OrderItem = z.object({
  order_id: z.uuid(),
  product_id: z.uuid(),
})

export const OrderItemIDRequest = z
  .object({
    user_id: z.uuid(),
  })
  .extend(OrderItem.shape)

export const OrderItemRequest = z
  .object({
    property: ProductProperty,
    price: z.number().min(0),
  })
  .extend(OrderItem.shape)

export const OrderUserInfoRequest = z.object({
  address: z.string().min(4).max(1000),
  zip_code: z.string().min(3).max(10),
})
export type OrderUserInfoInput = z.infer<typeof OrderUserInfoRequest>

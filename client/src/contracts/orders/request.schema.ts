import * as z from "zod"

import { ProductProperty } from "../products/request.schema"

export const OrderItem = z.object({
  order_id: z.uuid(),
  product_id: z.uuid(),
})

export const OrderItemIDRequest = z
  .object({
    user_id: z.uuid(),
  })
  .extend(OrderItem)

export const OrderItemRequest = z
  .object({
    property: ProductProperty,
    Price: z.number().min(0),
  })
  .extend(OrderItem)

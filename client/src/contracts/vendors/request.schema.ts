import * as z from 'zod'

import { ProductProperty } from '../products/request.schema'

export const VendorOrderDelivere = z.object({
  user_id: z.uuid(),
  order_id: z.uuid(),
  product_id: z.uuid(),
  property: ProductProperty,
  is_delivered: z.boolean(),
})
export type VendorOrderDelivereInput = z.infer<typeof VendorOrderDelivere>
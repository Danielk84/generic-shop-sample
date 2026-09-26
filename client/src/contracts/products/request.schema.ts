import * as z from 'zod'

export const ProductProperty = z.record(z.string(), z.string())

export const CreateProductRequest = z.object({
  name: z.string().min(4).max(256),
  description: z.string(),
  common_detail: ProductProperty,
})
export type CreateProductInput = z.infer<typeof CreateProductRequest>

const ProductIDRequest = z.object({
  id: z.uuid(),
})

export const UpdateProductRequest = z
  .object({})
  .extend(ProductIDRequest.shape)
  .extend(CreateProductRequest.shape)
export type UpdateProductInput = z.infer<typeof UpdateProductRequest>

export const ProductVendorRequest = z.object({
  user_id: z.uuid(),
  quantity: z.number().min(0),
})
export type ProductVendorInput = z.infer<typeof ProductVendorRequest>

export const ProductPropertyRequest = z.object({
  property: ProductProperty,
})

export const ProductVariantDetailRequest = z.object({
  items: z.array(
    z.object({
      property: ProductProperty,
      price: z.number().min(0),
      vendors: z.array(ProductVendorRequest),
    }),
  ),
})
export type ProductVariantDetailInput = z.infer<
  typeof ProductVariantDetailRequest
>

export const UpdateProductVendor = z
  .object({})
  .extend(ProductIDRequest.shape)
  .extend(ProductVendorRequest.shape)
  .extend(ProductPropertyRequest.shape)
export type UpdateProductVendorInput = z.infer<typeof UpdateProductVendor>

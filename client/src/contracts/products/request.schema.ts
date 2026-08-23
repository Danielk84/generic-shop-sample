import * as z from 'zod'

export const ProductProperty = z.record(z.string(), z.string())

export const CreateProductRequest = z.object({
  name: z.string().min(4).max(256),
  describtion: z.string(),
  common_detail: ProductProperty,
})
export type CreateProductInput = z.infer<typeof CreateProductRequest>

const ProductIDRequest = z.object({
  id: z.uuid(),
})

export const UpdateProductRequest = z
  .object({})
  .extend(ProductIDRequest)
  .extend(CreateProductRequest)
export type UpdateProductInput = z.infer<typeof UpdateProductRequest>

export const ProductVendorRequest = z.object({
  user_id: z.uuid(),
  quantity: z.number().min(0),
})
export type ProductVendorInput = z.infer<typeof ProductVendorRequest>

export const ProductPropertyRequest = z.object({
  property: ProductProperty
})

export const ProductVariantDetailRequet = z
  .object({
    price: z.number().min(0),
    variant_detail: z.array(ProductPropertyRequest),
  })
  .extend(ProductPropertyRequest)
export type ProductVariantDetailInput = z.infer<typeof ProductVariantDetailRequet>

export const UpdateProductVendor = z
  .object({})
  .extend(ProductIDRequest)
  .extend(ProductVendorRequest)
  .extend(ProductPropertyRequest)
export type UpdateProductVendorInput = z.infer<typeof UpdateProductVendor>
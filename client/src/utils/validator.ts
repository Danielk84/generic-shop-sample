import { ref } from 'vue'
import * as z from 'zod'

export function useValidator<T extends z.ZodType>(schema: T) {
  const formData = ref<Record<string, unknown>>({})
  const errors = ref<Record<string, string>>({})

  const validate = async () => {
    errors.value = {}

    let isValid = false
    const input = await schema.safeParseAsync(formData.value)
    if (input.success) {
      isValid = true
    }
    if (input.error !== undefined) {
      input.error.issues.forEach((err) => {
        errors.value[err.path.join('')] = err.message
      })
    }
    return { input, isValid }
  }

  return { formData, errors, validate }
}

export function isValidID(id: string | string[]): boolean {
  if (Array.isArray(id)) {
    id = id.join('')
  }
  return typeof id === 'string' && id !== ''
}

import type { ImageFrameCardProps } from '@/components/card/types';

export interface CartToralProps {
  totalItem?: number
  discount?: number
  total?: number
}

export interface OrderItemsCardProps extends ImageFrameCardProps {
  name: string
  price: number
  count: number
  total: number
}
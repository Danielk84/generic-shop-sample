export interface ImageFrameCardProps {
  img?: string
  alt?: string
  tag?: string
  loading?: "lazy" | "eager" | undefined
}

export interface ProductCardProps extends ImageFrameCardProps {
  to: string
  title: string
  price: number
}

export interface BlogCardProps {
  to: string
  banner?: string
  profileImage?: string
  username: string
  pudDate: string
  title: string
}

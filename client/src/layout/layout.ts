import { h, defineComponent, type Component, defineAsyncComponent } from 'vue'

const layouts = {
  FooterBlockLayout: defineAsyncComponent(
    () => import('@/layout/FooterBlockLayout.vue'),
  ),
  HasAccessTokenLayout: defineAsyncComponent(
    () => import('@/layout/HasAccessTokenLayout.vue'),
  ),
  BaseFooterLayout: defineAsyncComponent(
    () => import('@/layout/BaseFooterLayout.vue'),
  ),
} as const

export type LayoutName = keyof typeof layouts

export const setLayout = (layer: LayoutName[], base: Component): Component => {
  return defineComponent({
    name: 'DynamicLayout',

    setup(_, { slots }) {
      return () => {
        return h(base, null, {
          default: () =>
            layer.reduce((child: Component, layoutName: LayoutName) => {
              const Layout = layouts[layoutName]
              if (Layout === undefined) {
                throw new Error(`failed to wrap layout "${layoutName}"`)
              }
              return h(Layout, null, {
                default: () => child,
              })
            }, slots.default?.() ?? []),
        })
      }
    },
  })
}

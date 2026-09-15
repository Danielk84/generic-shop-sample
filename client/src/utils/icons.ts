const icon = {
  card: {
    plusCircle: 'card/PlusCircleIcon.vue',
  },
  common: {
    navBar: {
      infinity: 'common/nav-bar/InfinityIcon.vue',
      profile: 'common/nav-bar/ProfileIcon.vue',
      shoppingBag: 'common/nav-bar/ShoppingBagIcon.vue',
      dropdown: 'common/nav-bar/DropdownIcon.vue',
    },
    footer: {
      send: 'common/footer/SendIcon.vue',
    },
    loading: {
      loading: 'common/loading/LoadingIcon.vue',
    },
  },
  ui: {
    button: {
      back: 'ui/BackArrowIcon.vue',
    },
    search: {
      btn: 'ui/search/SearchIcon.vue',
    },
    pagination: {
      next: 'ui/pagination/NextIcon.vue',
      previous: 'ui/pagination/PreviousIcon.vue',
    },
  },
  pages: {
    landing: {
      rightArrow: 'pages/landing/RightArrowIcon.vue',
    },
    auth: {
      eyeOn: 'pages/auth/EyeOnIcon.vue',
      eyeOff: 'pages/auth/EyeOffIcon.vue',
    },
    basket: {
      close: 'pages/basket/CloseIcon.vue',
      plus: 'pages/basket/PlusIcon.vue',
      minus: 'pages/basket/MinusIcon.vue',
    },
    products: {
      close: 'pages/products/CloseIcon.vue',
      fullScreen: 'pages/products/FullScreenIcon.vue',
    },
    panel: {
      products: {
        upload: 'pages/panel/products/UploadIcon.vue',
      },
    },
  },
} as const

export default icon

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
    products: {
      close: 'common/products/CloseIcon.vue',
      fullScreen: 'common/products/FullScreenIcon.vue',
      basket: 'common/products/BasketIcon.vue',
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
    theme: {
      sun: 'ui/SunIcon.vue',
      moon: 'ui/MoonIcon.vue',
    },
  },
  pages: {
    account: {
      verified: 'pages/account/VerifiedIcon.vue',
      notVerified: 'pages/account/NotVerifiedIcon.vue',
    },
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
    panel: {
      products: {
        upload: 'pages/panel/products/UploadIcon.vue',
        save: 'pages/panel/products/SaveIcon.vue',
      },
    },
  },
} as const

export default icon

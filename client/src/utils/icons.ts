const icon = {
  card: {
    plusCircle: 'card/plus-circle-icon.svg',
  },
  common: {
    navBar: {
      infinity: 'ui/nav-bar/infinity-icon.svg',
      profile: 'ui/nav-bar/profile-icon.svg',
      shoppingBag: 'ui/nav-bar/shopping-bag-icon.svg',
      dropdown: 'ui/nav-bar/dropdown-icon.svg',
    },
    footer: {
      send: 'common/footer/send-icon.svg',
    }
  },
  ui: {
    search: {
      btn: 'ui/search/search-icon.svg',
    },
    pagination: {
      next: 'ui/pagination/next-icon.svg',
      previous: 'ui/pagination/previous-icon.svg',
    },
  },
  pages: {
    basket: {
      close: 'pages/basket/close-icon.svg',
      plus: 'pages/basket/plus-icon.svg',
      minus: 'pages/basket/minus-icon.svg',
    },
    products: {
      close: 'pages/products/close-icon.svg',
      fullScreen: 'pages/products/full-screen-icon.svg',
    }
  }
} as const

export default icon
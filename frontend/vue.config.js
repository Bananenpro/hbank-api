module.exports = {
  pluginOptions: {
    i18n: {
      locale: 'en',
      fallbackLocale: 'en',
      localeDir: 'locales',
      enableLegacy: false,
      runtimeOnly: false,
      compositionOnly: false,
      fullInstall: true
    }
  },
  pwa: {
    name: 'H-Bank',
    themeColor: '#101010',
    msTileColor: '#101010'
  },
  configureWebpack: {
    performance: {
        maxEntrypointSize: 512000,
        maxAssetSize: 512000
    },
  }
}

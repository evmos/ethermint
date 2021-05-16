const path = require('path')

module.exports = {
  devServer: {
    disableHostCheck: true,
  },
	configureWebpack: {
		resolve: {
			symlinks: false,
			alias: {
				vue$: path.resolve('./node_modules/vue/dist/vue.esm-bundler.js')
			}
		}
	}
}
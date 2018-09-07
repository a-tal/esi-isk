const path = require('path');
const webpack = require('webpack');

module.exports = {
  entry: {
    index: './src/index.js'
  },
  module: {
    rules: [
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader']
      }
    ]
  },
  plugins: [
    // expose $ and jQuery to global scope.
    new webpack.ProvidePlugin({
      $: 'jquery',
      jQuery: 'jquery',
      'window.jQuery': 'jquery',
      'window.$': 'jquery',
    })
  ],
  output: {
    filename: 'main.js',
    chunkFilename: '[name].bundle.js',
    path: path.resolve(__dirname, 'dist')
  },
  devServer: {
   port: 3200,
   publicPath: "/",
   noInfo: true,
   hot: true,
   disableHostCheck: true,
   stats: {
     colors: true
   }
  }
};

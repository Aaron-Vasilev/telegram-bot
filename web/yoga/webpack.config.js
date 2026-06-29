const path = require('path')
const HtmlWebpackPlugin = require('html-webpack-plugin')
const webpack = require('webpack')

module.exports = {
  entry: './src/entry.js',
  output: {
    filename: 'bundle.js',
    path: path.resolve(__dirname, 'dist'),
    publicPath: process.env.PUBLIC_PATH || '/',
    clean: true,
  },
  mode: 'development',
  devServer: {
    static: {
      directory: path.join(__dirname, 'dist'),
    },
    hot: true,
    port: 3030,
    watchFiles: ['src/**/*'],
    open: true,
    allowedHosts: 'all',
    proxy: [
      {
        context: ['/create-payment', '/payment-success'],
        target: `http://localhost:${process.env.HTTP_PORT || 8080}`,
        changeOrigin: true,
      },
    ],
  },

  plugins: [
    require('postcss-nested'),
    new HtmlWebpackPlugin({
      template: './src/pages/index.html',
      filename: 'index.html',
    }),
    new webpack.DefinePlugin({
      'process.env.PAYMENT_API_URL': JSON.stringify(process.env.PAYMENT_API_URL || ''),
    }),
  ],

  module: {
    rules: [
      {
        test: /\.css$/i,
        use: [
          'style-loader',
          'css-loader',
          {
            loader: 'postcss-loader',
            options: {
              postcssOptions: {
                plugins: [
                  [
                    'postcss-preset-env',
                  ],
                ],
              },
            },
          }
        ],
      },
      {
        test: /\.(png|jpe?g|gif|svg)$/i,
        type: 'asset/resource',
      },
      {
        test: /\.html$/,
        exclude: /src\/pages\//, // Exclude pages that HtmlWebpackPlugin processes                                                                                                                                                            
        use: 'html-loader',
      },
    ],
  },
}

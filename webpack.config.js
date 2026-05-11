const path = require('path');
// Import HtmlWebpackPlugin to generate the HTML file that loads the bundle
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
    mode: 'production',

    entry: './web/static/js/index.js',

    output: {
        filename: 'bundle.js',
        path: path.resolve(__dirname, './web/assets/js/'),
        clean: true, // Cleans the output directory before build (Webpack 5 feature)
    },

    module: {
        rules: [
            {
                test: /\.(js|jsx)$/,
                exclude: /node_modules/,
                use: {
                    loader: 'babel-loader',
                    options: {
                        presets: ['@babel/preset-env', '@babel/preset-react'],
                    },
                },
            },
        ],
    },

    plugins: [
        new HtmlWebpackPlugin({
            template: './web/static/index.html',
            filename: '../../index.html',
        }),
    ],

    resolve: {
        // Add '.jsx' to the list of extensions so you can import 'Component' instead of 'Component.jsx'
        extensions: ['.js', '.jsx']
    },

    performance: {
        maxAssetSize: 1000000,
        maxEntrypointSize: 1000000,
    }
};

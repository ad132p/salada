const path = require('path');
// Import HtmlWebpackPlugin to generate the HTML file that loads the bundle
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
    // Kept user's production mode setting
    mode: 'production',

    // Kept user's specific entry point (assuming this is your main React file, e.g., 'index.js' where you call ReactDOM.createRoot())
    entry: './web/static/js/index.js',

    output: {
        // Kept user's specific output paths
        filename: 'bundle.js',
        path: path.resolve(__dirname, './web/assets/js/'),
        clean: true, // Cleans the output directory before build (Webpack 5 feature)
    },

    module: {
        rules: [
            // Rule to process JavaScript and JSX files with Babel
            {
                // Target both .js and .jsx files
                test: /\.(js|jsx)$/,
                // Exclude the node_modules folder from being processed
                exclude: /node_modules/,
                use: {
                    loader: 'babel-loader',
                    // Note: It's often cleaner to put Babel configuration
                    // in a separate file (e.g., .babelrc or babel.config.js)
                    // but you can configure it here if you prefer.
                    options: {
                        // The @babel/preset-env handles modern JS features.
                        // The @babel/preset-react handles JSX and React-specific features.
                        presets: ['@babel/preset-env', '@babel/preset-react'],
                        // Add plugins here if necessary for features not covered by presets
                        // e.g., Hot Module Replacement (HMR) for development
                    },
                },
            },
        ],
    },

    plugins: [
        // Generates an index.html file that automatically includes your 'bundle.js'.
        new HtmlWebpackPlugin({
            template: './web/static/index.html',
            filename: '../../index.html',
        }),
    ],

    resolve: {
        // Add '.jsx' to the list of extensions so you can import 'Component' instead of 'Component.jsx'
        extensions: ['.js', '.jsx']
    }
};
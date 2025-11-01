const path = require('path');

module.exports = {
    // Set mode to production for optimized, smaller output
    mode: 'production',
    
    // The entry point for our JavaScript application
    entry: './web/static/js/index.js',
    
    // How and where to output the final bundle
    output: {
        filename: 'bundle.js',
        // Output path resolves to the 'dist' directory relative to the project root
        path: path.resolve(__dirname, './web/assets/js/'),
        // Clean the 'assets' folder before each build
        clean: true,
    },
    
    // Define external libraries that should not be bundled but are expected to be available 
    // globally or via CDN (not strictly necessary here, but good practice if you had many dependencies)
    resolve: {
        extensions: ['.js']
    }
};


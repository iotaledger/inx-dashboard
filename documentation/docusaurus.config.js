const path = require('path');

module.exports = {
    plugins: [
        [
            '@docusaurus/plugin-content-docs',
            {
                id: 'inx-dashboard',
                path: path.resolve(__dirname, 'docs'),
                routeBasePath: 'inx-dashboard',
                sidebarPath: path.resolve(__dirname, 'sidebars.js'),
                editUrl: 'https://github.com/iotaledger/inx-dashboard/edit/develop/documentation/docs',
            }
        ],
    ],
    staticDirectories: [path.resolve(__dirname, 'static')],
};

function require(jspath) {
    var scriptName = 'gbi.js';
    var r = new RegExp('(^|(.*?\\/))(' + scriptName + ')(\\?|$)'),
        s = document.getElementsByTagName('script'),
        src, m, l = '';
    for(var i=0, len=s.length; i<len; i++) {
        src = s[i].getAttribute('src');
        if(src) {
            m = src.match(r);
            if(m) {
                l = m[1];
                break;
            }
        }
    }
    document.write('<script type="text/javascript" src="'+ l + jspath+'"><\/script>');
}
require('libs/detect-element-resize.js');
require('libs/ol.js');
require('libs/ace.js');
require('libs/ace-mode-sql.js');
require('libs/angular.js');
require('libs/angular-gridster.js');
require('libs/angular-route.js');
require('libs/ng-websocket.js');
require('libs/ui-bootstrap-tpls-0.13.3.js');
require('libs/ng-sortable.js');
require('libs/ui-ace.js');
require('src/app.js');
require('src/helper/focusme-directive.js');
require('src/helper/stopevent-directive.js');
require('src/helper/float-directive.js');
require('src/helper/resizer.js');
require('src/helper/title-case-filter.js');
require('src/logging/logging-service.js');
require('src/logging/logging-controller.js');
require('src/notification/notification-directive.js');
require('src/notification/notification-controller.js');
require('src/layer/layer-service.js');
require('src/layer/edit-layer-form-status-service.js');
require('src/layer/edit-layer-controller.js');
require('src/layer/remove-layer-controller.js');
require('src/style/style-service.js');
require('src/projects/projects-service.js');
require('src/projects/project-service.js');
require('src/projects/projects-controller.js');
require('src/sidebar/side-nav-service.js');
require('src/sidebar/side-nav-controller.js');
require('src/sidebar/style-list-controller.js');
require('src/sidebar/layer-list-controller.js');
require('src/dashboard/dashboard-service.js');
require('src/dashboard/ol3-directive.js');
require('src/dashboard/dashboard-controller.js');
require('src/dashboard/bookmark-map-controller.js');
require('src/bookmarks/bookmarks-controller.js');
require('src/about/about-controller.js');
require('src/form/form-group-directive.js');
require('src/form/input-dropdown-directive.js');

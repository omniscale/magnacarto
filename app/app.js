angular.module('magna-app', ['ngRoute', 'ngWebsocket', 'gridster', 'ui.bootstrap']);

// TODO get config values from elsewhere?
angular.module('magna-app').constant('magnaConfig', {
    socketUrl: 'ws://localhost:7070/changes?',
    mapnikUrl: 'http://localhost:7070/mapnik?',
    mapnikLayers: 'osm',
    mapnikImageFormat: 'image/png',
    defaultCenter: [8, 53],
    defaultZoom: 12,
    mml: 'omni-live.mml'
})

.config(function($routeProvider){
  $routeProvider
  .when('/dashboard', {
    templateUrl: 'app/dashboard/template.html',
    controller: 'DashboardCtrl'
  })
  .when('/storage', {
    templateUrl: 'app/storage/template.html',
    controller: 'StorageCtrl'
  })
  .otherwise({
    redirectTo: '/dashboard'
  });
})

.config(function(MMLServiceProvider) {
  MMLServiceProvider.setSaveUrl('http://localhost:8000/save');
  MMLServiceProvider.setLoadUrl('http://localhost:8000/');
})

.run(function($websocket, $rootScope, magnaConfig, MMLService, DashboardService, StyleService) {
  // Load project file (mml)
  var promise = MMLService.load(magnaConfig.mml);
  promise.success(function() {
    // add all style files to dashboard object
    StyleService.setStyles(MMLService.styles);
    DashboardService.maps = MMLService.dashboardMaps;
    DashboardService.layers = [{
      styles: StyleService.activeStyles,
      mml: magnaConfig.mml
    }];

    // create websocket
    magnaConfig.socketUrl += 'mml=' + MMLService.mml + '&mss=' + StyleService.styles;
    $websocket.$new({
      url: magnaConfig.socketUrl,
      reconnect: true,
      reconnectInterval: 100
    });
  });
});

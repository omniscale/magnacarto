angular.module('magna-app', ['ngRoute', 'ngCookies', 'ngWebsocket', 'gridster', 'ui.bootstrap']);

// TODO get config values from elsewhere?
angular.module('magna-app').constant('magnaConfig', {
    socketUrl: 'ws://localhost:7070/changes?',
    mapnikUrl: 'http://localhost:7070/mapnik?',
    mapnikLayer: 'osm',
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

.run(function($websocket, $rootScope, magnaConfig, MMLService, DashboardService, StyleService) {
  // Load project file (mml)
  var promise = MMLService.load(magnaConfig.mml);
  promise.success(function() {
    // add all style files to dashboard object
    StyleService.setStyles(MMLService.styles);

    DashboardService.layers = [{
      styles: StyleService.activeStyles,
      mml: magnaConfig.mml
    }];

    // create websocket
    var webSocketURL = magnaConfig.socketUrl + 'mml=' + magnaConfig.mml + '&mss=' + StyleService.activeStyles;
    var ws = $websocket.$new({
      url: webSocketURL,
      reconnect: true,
      reconnectInterval: 100
    });

    // TODO check if we need this realy
    ws.$on('$open', function () {
       $rootScope.$broadcast('socketOpen');
    })

    .$on('$message', function (resp) {
      // TODO get rid of broadcast event
      // show fancy modal with error msg
      if(resp.error !== undefined) {
        $rootScope.$broadcast('socketError', resp);
      } else {
        // reload map
        $rootScope.$broadcast('socketUpdateImage', resp);
      }
    });
  });
});

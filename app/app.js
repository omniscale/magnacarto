angular.module('magna-app', ['ngRoute', 'ngCookies', 'ngWebsocket', 'gridster', 'ui.bootstrap']);

angular.module('magna-app').constant('socketConfig', {
    'url': 'ws://localhost:7070/changes?'
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

.run(function($websocket, $rootScope, socketConfig) {
  $rootScope.dashboard = dashboard;

  var mss;
  var mml;

  // add websocket for each layer
  angular.forEach($rootScope.dashboard.layers, function(layer) {
    mss = layer.mss.join(',');
    mml = layer.mml;

    var webSocketURL =  socketConfig.url + 'mml=' + mml + '&mss=' + mss;
    var ws = $websocket.$new({
      url: webSocketURL,
      reconnect: true,
      reconnectInterval: 100
    });

    ws.$on('$open', function () {
       $rootScope.$broadcast('socketOpen');
    })

    .$on('$message', function (resp) {
      if(resp.error !== undefined) {
        $rootScope.$broadcast('socketError', resp);
      } else {
        $rootScope.$broadcast('socketUpdateImage', resp);
      }
    });

  });
})

.controller('SideNavCtrl',['$scope', function($scope){
  $scope.active = true;
}])

.controller('CollapseDemoCtrl', ['$scope', function ($scope) {
  $scope.isCollapsed = false;
}]);
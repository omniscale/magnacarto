angular.module('magna-app', ['ngRoute', 'ngWebsocket', 'gridster', 'ui.bootstrap']);

// TODO get config values from elsewhere?
angular.module('magna-app').constant('magnaConfig', {
    socketUrl: 'ws://localhost:7070/api/v1/changes?',
    mapnikUrl: 'http://localhost:7070/api/v1/map?',
    mapnikLayers: 'osm',
    mapnikImageFormat: 'image/png',
    defaultCenter: [8, 53],
    defaultZoom: 12,
    mml: '/example/example.mml'
})

.config(function($routeProvider){
  $routeProvider
  .when('/projects', {
    templateUrl: 'src/projects/template.html',
    controller: 'ProjectsCtrl'
  })
  .when('/dashboard', {
    templateUrl: 'src/dashboard/template.html',
    controller: 'DashboardCtrl'
  })
  .when('/storage', {
    templateUrl: 'src/storage/template.html',
    controller: 'StorageCtrl'
  })
  .otherwise({
    redirectTo: '/projects'
  });
})

.config(function(ProjectsServiceProvider, MMLServiceProvider) {
  ProjectsServiceProvider.setProjectsUrl('http://localhost:8888/proxy/http://localhost:7070/api/v1/projects');
  // MMLServiceProvider.setBaseUrl('http://localhost:8888/proxy/')
  MMLServiceProvider.setBaseUrl('http://localhost:8888/proxy/http://localhost:7070/api/v1/projects/');
  MMLServiceProvider.setSaveUrl('http://localhost:8000/save');
  MMLServiceProvider.setLoadUrl('http://localhost:8000/');
})

.run(function($websocket, $rootScope, magnaConfig, ProjectsService, MMLService, DashboardService, StyleService) {
  var projectsPromise = ProjectsService.load();
});

angular.module('magna-app', ['ngRoute', 'ngWebsocket', 'gridster', 'ui.bootstrap']);

// TODO get config values from elsewhere?
angular.module('magna-app').constant('magnaConfig', {
  projectsUrl: 'http://localhost:8888/proxy/http://localhost:7070/api/v1/projects',
  projectBaseUrl: 'http://localhost:8888/proxy/http://localhost:7070/api/v1/projects/',
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

.run(function(ProjectsService) {
  ProjectsService.load();
});

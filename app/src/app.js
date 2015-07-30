angular.module('magna-app', ['ngRoute', 'ngWebsocket', 'gridster', 'ui.bootstrap', 'as.sortable']);

// TODO get config values from elsewhere?
angular.module('magna-app').constant('magnaConfig', {
  projectsUrl: '/api/v1/projects',
  projectBaseUrl: '/api/v1/projects/',
  socketUrl: 'ws://' + window.location.host + '/api/v1/changes?',
  mapnikUrl: '/api/v1/map?',
  mapnikLayers: 'osm',
  mapnikImageFormat: 'image/png',
  defaultCenter: [8, 53],
  defaultZoom: 12
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
  .when('/bookmarks', {
    templateUrl: 'src/bookmarks/template.html',
    controller: 'BookmarksCtrl'
  })
  .otherwise({
    redirectTo: '/projects'
  });
})

.run(function(ProjectsService) {
  ProjectsService.load();
});

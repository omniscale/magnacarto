angular.module('magna-app', ['ngRoute', 'ngWebsocket', 'gridster', 'ui.bootstrap', 'as.sortable', 'ui.ace']);

// TODO get config values from elsewhere?
angular.module('magna-app').constant('magnaConfig', {
  projectsUrl: '/api/v1/projects',
  projectBaseUrl: '/api/v1/projects/',
  socketUrl: 'ws://' + window.location.host + '/api/v1/changes?',
  mapnikUrl: '/api/v1/map?',
  mapnikLayers: 'osm',
  mapnikImageFormat: 'image/png',
  version: '0.1',
  defaultSuggestions: {
    srs: [
      '+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0.0 +k=1.0 +units=m +nadgrids=@null +wktext +no_defs +over',
      '+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs'
    ],
    extent: [
      '-90, -180, 90, 180',
      '-20026376.39, -20048966.10, 20026376.39, 20048966.10'
    ],
    geometry_field: ['geometry', 'geom'],
    srid: [],
    host: ['localhost'],
    port: [5432],
    dbname: ['osm'],
    user: ['osm'],
    password: ['osm']
  },
  selectMultipleBookmarksKey: 'shiftKey'
})

.config(function($routeProvider){
  var loadProjectOrRedirect = function($q, $location, $route, ProjectsService, ProjectService) {
    var deferred = $q.defer();
    ProjectsService.loaded().then(function() {
      var project = ProjectsService.projectByUrl($route.current.params.projectUrl);
      if(project === undefined) {
        $location.path('projects');
        deferred.reject();
        return;
      }

      if(ProjectService.projectLoaded() === undefined || ProjectService.project !== project) {
        var loadedPromise = ProjectService.loadProject(project);
        loadedPromise.then(function() {
          deferred.resolve();
        });
      }
      deferred.resolve();
    });
    return deferred.promise;
  };

  $routeProvider
    .when('/projects', {
      templateUrl: 'src/projects/projects-template.html',
      controller: 'ProjectsCtrl'
    })
    .when('/dashboard/:projectUrl*', {
      templateUrl: 'src/dashboard/dashboard-template.html',
      controller: 'DashboardCtrl',
      resolve: {
        projectPromise: function ($q, $location, $route, ProjectsService, ProjectService) {
          return loadProjectOrRedirect($q, $location, $route, ProjectsService, ProjectService);
        }
      }
    })
    .when('/bookmarks/:projectUrl*', {
      templateUrl: 'src/bookmarks/bookmarks-template.html',
      controller: 'BookmarksCtrl',
      resolve: {
        projectPromise: function ($q, $location, $route, ProjectsService, ProjectService) {
          return loadProjectOrRedirect($q, $location, $route, ProjectsService, ProjectService);
        }
      }
    })
    .otherwise({
      redirectTo: '/projects'
    });
})

.controller('AppCtrl', ['$scope', 'ProjectService', function($scope, ProjectService) {
  $scope.appSettings = ProjectService.appSettings;

  $scope.$watch(function() {
    return ProjectService.appSettings;
  }, function() {
    $scope.appSettings = ProjectService.appSettings;
  });
}])

.run(function($q, $templateRequest, ProjectsService) {
  // fill template cache before load projects
  // needed because resizer otherwise might not find all elements
  $q.all([
    $templateRequest('src/sidebar/layer-list-template.html'),
    $templateRequest('src/sidebar/style-list-template.html'),
    $templateRequest('src/logging/logging-template.html'),
    $templateRequest('src/sidebar/side-nav-template.html')
  ]).then(function() {
    ProjectsService.load();
  });
});

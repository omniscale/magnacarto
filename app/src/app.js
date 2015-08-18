angular.module('magna-app', ['ngRoute', 'ngWebsocket', 'gridster', 'ui.bootstrap', 'as.sortable', 'ui.ace']);

// TODO get config values from elsewhere?
angular.module('magna-app').constant('magnaConfig', {
  projectsUrl: '/api/v1/projects',
  projectBaseUrl: '/api/v1/projects/',
  socketUrl: 'ws://' + window.location.host + '/api/v1/changes?',
  mapnikUrl: '/api/v1/map?',
  mapnikLayers: 'osm',
  mapnikImageFormat: 'image/png',
  defaultCenter: [8, 53],
  defaultZoom: 12,
  version: '0.1'
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

.run(function(ProjectsService) {
  ProjectsService.load();
});

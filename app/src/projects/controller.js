angular.module('magna-app')

.controller('ProjectsCtrl', ['$scope', '$location', 'ProjectsService', 'MMLService',
  function($scope, $location, ProjectsService, MMLService) {
    $scope.projects = [];

    MMLService.unloadProject();

    ProjectsService.loaded().success(function() {
      $scope.projects = ProjectsService.projects;
    });

    $scope.openProject = function(project) {
      var promise = MMLService.loadProject(project);
      promise.then(function() {
        $location.path('dashboard');
      });
    };
  }
]);

angular.module('magna-app')

.controller('ProjectsCtrl', ['$scope', '$location', 'ProjectsService', 'ProjectService',
  function($scope, $location, ProjectsService, ProjectService) {
    $scope.projects = [];
    $scope.navItemName = 'projects';

    ProjectService.unloadProject();

    ProjectsService.loaded().success(function() {
      $scope.projects = ProjectsService.projects;
    });

    $scope.openProject = function(project) {
      var promise = ProjectService.loadProject(project);
      promise.then(function() {
        $location.path('dashboard/' + project.base + '/' + project.mml);
      });
    };
  }
]);

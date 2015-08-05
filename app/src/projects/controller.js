angular.module('magna-app')

.controller('ProjectsCtrl', ['$scope', '$location', 'ProjectsService', 'ProjectService', 'EditLayerFormStatus', 'SideNavStatusService',
  function($scope, $location, ProjectsService, ProjectService, EditLayerFormStatus, SideNavStatusService) {
    $scope.projects = [];
    $scope.navItemName = 'projects';

    ProjectService.unloadProject();
    EditLayerFormStatus.reset();
    SideNavStatusService.reset();

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

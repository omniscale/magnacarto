angular.module('magna-app')

.controller('ProjectsCtrl', ['$scope', '$location', 'ProjectsService', 'ProjectService', 'EditLayerFormStatusService', 'SideNavService',
  function($scope, $location, ProjectsService, ProjectService, EditLayerFormStatusService, SideNavService) {
    $scope.projects = [];

    ProjectService.unloadProject();

    SideNavService.reset();
    SideNavService.currentPage('projects');

    EditLayerFormStatusService.reset();

    ProjectsService.loaded().then(function() {
      $scope.projects = ProjectsService.projectsList;
    });

    $scope.openProject = function(project) {
      var promise = ProjectService.loadProject(project);
      promise.then(function() {
        $location.path('dashboard/' + project.url);
      });
    };
  }
]);

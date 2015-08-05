angular.module('magna-app')

.controller('SideNavCtrl',['$scope', '$route', 'DashboardService', 'LayerService', 'ProjectService', function($scope, $route, DashboardService, LayerService, ProjectService) {
  $scope.active = true;
  $scope.selectedNavItem = undefined;
  $scope.styles = DashboardService.styles;
  $scope.activeStyles = DashboardService.activeStyles;
  $scope.layers = LayerService.layers;
  $scope.dashboardUrl = '#/dashboard/';
  $scope.bookmarksUrl = '#/bookmarks/';

  $scope.addMap = function() {
    DashboardService.addMap();
  };

  $scope.createLayer = function() {
    LayerService.addLayer();
  };

  $scope.$watch(function() {
    return $route.current && $route.current.scope ? $route.current.scope.navItemName : undefined;
  }, function(newNavItem) {
    $scope.selectedNavItem = newNavItem;
  });

  $scope.$watch(function() {
    return LayerService.layers;
  }, function(newLayers) {
    $scope.layers = newLayers;
  }, true);

  $scope.$watch(function() {
    return ProjectService.project;
  }, function(project) {
    if(project !== undefined) {
      $scope.dashboardUrl = '#/dashboard/' + project.base + '/' + project.mml;
      $scope.bookmarksUrl = '#/bookmarks/' + project.base + '/' + project.mml;
    } else {
      $scope.dashboardUrl = '#/dashboard/';
      $scope.bookmarksUrl = '#/bookmarks/';
    }
  });
}]);

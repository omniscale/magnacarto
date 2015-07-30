angular.module('magna-app')

.controller('SideNavCtrl',['$scope', '$route', '$modal', 'DashboardService', 'LayerService', function($scope, $route, $modal, DashboardService, LayerService) {
  $scope.active = true;
  $scope.selectedNavItem = undefined;
  $scope.styles = DashboardService.styles;
  $scope.activeStyles = DashboardService.activeStyles;

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
}]);

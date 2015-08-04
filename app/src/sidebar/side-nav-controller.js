angular.module('magna-app')

.controller('SideNavCtrl',['$scope', '$route', 'DashboardService', 'LayerService', function($scope, $route, DashboardService, LayerService) {
  $scope.active = true;
  $scope.selectedNavItem = undefined;
  $scope.styles = DashboardService.styles;
  $scope.activeStyles = DashboardService.activeStyles;
  $scope.layers = LayerService.layers;

  $scope.addMap = function() {
    DashboardService.addMap();
  };

  $scope.createLayer = function() {
    LayerService.addLayer();
  };

  $scope.copyLayer = function(layer) {
    LayerService.copyLayer(angular.copy(layer));
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
}]);

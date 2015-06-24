angular.module('magna-app')

.controller('SideNavCtrl',['$scope', '$route', 'DashboardService', function($scope, $route, DashboardService) {
  $scope.active = true;
  $scope.selectedNavItem = undefined;
  $scope.styles = DashboardService.styles;
  $scope.activeStyles = DashboardService.activeStyles;

  $scope.addMap = function() {
    DashboardService.addMap();
  };

  $scope.clearMaps = function() {
    DashboardService.clearMaps();
  };

  $scope.$watch(function() {
    return $route.current && $route.current.scope ? $route.current.scope.navItemName : undefined;
  }, function(newNavItem) {
    $scope.selectedNavItem = newNavItem;
  });
}]);

angular.module('magna-app')

.controller('SideNavCtrl',['$scope', '$route', 'DashboardService', function($scope, $route, DashboardService) {
  $scope.active = true;
  $scope.selectedNavItem = undefined;
  $scope.mss = DashboardService.mss;
  $scope.activeMss = DashboardService.activeMss;

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

angular.module('magna-app')

.controller('LayerListCtrl', ['$scope', 'DashboardService',
  function($scope, DashboardService) {
    $scope.mss = DashboardService.mss;
    $scope.activeMss = DashboardService.activeMss;

    $scope.$watch(function() {
      return DashboardService.mss;
    }, function(newMss) {
      $scope.mss = newMss;
    }, true);

    $scope.$watch(function() {
      return DashboardService.activeMss;
    }, function(newMss) {
      $scope.activeMss = newMss;
    }, true);

    $scope.toggleSelection = function(style) {
      DashboardService.toggleMss(style);
    };
}]);

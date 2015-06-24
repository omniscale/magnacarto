angular.module('magna-app')

.controller('LayerListCtrl', ['$scope', 'DashboardService',
  function($scope, DashboardService) {
    $scope.styles = DashboardService.styles;
    $scope.activeStyles = DashboardService.activeStyles;

    $scope.$watch(function() {
      return DashboardService.styles;
    }, function(newStyles) {
      $scope.styles = newStyles;
    }, true);

    $scope.$watch(function() {
      return DashboardService.activeStyles;
    }, function(newStyles) {
      $scope.activeStyles = newStyles;
    }, true);

    $scope.toggleSelection = function(style) {
      DashboardService.toggleStyle(style);
    };
}]);

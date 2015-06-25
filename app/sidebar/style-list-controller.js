angular.module('magna-app')

.controller('StyleListCtrl', ['$scope', 'StyleService',
  function($scope, StyleService) {
    $scope.collapsed = false;
    $scope.styles = StyleService.styles;
    $scope.activeStyles = StyleService.activeStyles;

    $scope.$watch(function() {
      return StyleService.styles;
    }, function(newStyles) {
      $scope.styles = newStyles;
    }, true);

    $scope.$watch(function() {
      return StyleService.activeStyles;
    }, function(newStyles) {
      $scope.activeStyles = newStyles;
    }, true);

    $scope.toggleSelection = function(style) {
      StyleService.toggleStyle(style);
    };
}]);

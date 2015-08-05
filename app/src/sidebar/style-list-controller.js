angular.module('magna-app')

.controller('StyleListCtrl', ['$scope', 'StyleService', 'SideNavStatusService',
  function($scope, StyleService, SideNavStatusService) {
    $scope.collapsed = SideNavStatusService.hideStyles();

    $scope.styles = StyleService.styles;
    $scope.activeStyles = StyleService.activeStyles;

    $scope.toggleCollapsed = function() {
      $scope.collapsed = $scope.selectedNavItem === 'projects' ? true : !$scope.collapsed;
      SideNavStatusService.hideStyles($scope.collapsed);
    };

    $scope.toggleSelection = function(style) {
      StyleService.toggleStyle(style);
    };

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

    $scope.$on('$routeChangeSuccess', function(event, toState) {
      if(toState.controller !== 'ProjectsController') {
        $scope.collapsed = SideNavStatusService.hideStyles();
      }
    });
}]);

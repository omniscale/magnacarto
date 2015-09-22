angular.module('magna-app')

.controller('StyleListCtrl', ['$scope', 'StyleService', 'SideNavService',
  function($scope, StyleService, SideNavService) {
    $scope.collapsed = SideNavService.hideStyles();

    $scope.styles = StyleService.styles;
    $scope.activeStyles = StyleService.activeStyles;

    $scope.toggleCollapsed = function() {
      $scope.collapsed = $scope.selectedNavItem === 'projects' ? true : !$scope.collapsed;
      SideNavService.hideStyles($scope.collapsed);
    };

    $scope.toggleSelection = function(style) {
      StyleService.toggleStyle(style);
    };

    $scope.inActiveStyles = function(style) {
      return StyleService.inActiveStyles(style);
    };

    $scope.$watch(function() {
      return StyleService.styles;
    }, function(newStyles) {
      $scope.styles = newStyles;
    }, true);

    $scope.$watch(function() {
      return StyleService.activeStyles;
    }, function(newStyles) {
      $scope.activeStyles = StyleService.activeStyles;
    }, true);

    $scope.$on('$routeChangeSuccess', function(event, toState) {
      if(toState.controller !== 'ProjectsController') {
        $scope.collapsed = SideNavService.hideStyles();
      }
    });
}]);

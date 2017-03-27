angular.module('magna-app')

.controller('LayerListCtrl', ['$scope', 'LayerService', 'SideNavService',
  function($scope, LayerService, SideNavService) {
    $scope.collapsed = SideNavService.hideLayers();
    $scope.layers = LayerService.layers;

    $scope.toggleCollapsed = function() {
      $scope.collapsed = $scope.selectedNavItem === 'projects' ? true : !$scope.collapsed;
      SideNavService.hideLayers($scope.collapsed);
    };

    $scope.toggle = function(layer) {
      layer.status = layer.status === 'off' ? '' : 'off';
    };

    $scope.showOnly = function(layer) {
      angular.forEach($scope.layers, function(_layer) {
        _layer.status = _layer === layer ? '' : 'off';
      });
    };

    $scope.openEditLayerModal = function(layer) {
      LayerService.editLayer(layer);
    };

    $scope.$watch(function() {
      return LayerService.layers;
    }, function(newLayers) {
      $scope.layers = newLayers;
    }, true);

    $scope.$on('$routeChangeSuccess', function(event, toState) {
      if(toState.controller !== 'ProjectsController') {
        $scope.collapsed = SideNavService.hideLayers();
      }
    });
}]);

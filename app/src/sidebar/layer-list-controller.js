angular.module('magna-app')

.controller('LayerListCtrl', ['$scope', 'LayerService',
  function($scope, LayerService) {
    $scope.collapsed = false;
    $scope.layers = LayerService.layers;

    $scope.$watch(function() {
      return LayerService.layers;
    }, function(newLayers) {
      $scope.layers = newLayers;
    }, true);

    $scope.toggle = function(layer) {
      layer.status = layer.status === 'off' ? '' : 'off';
    };

    $scope.openEditLayerModal = function(layer) {
      LayerService.editLayer(layer);
    };
}]);

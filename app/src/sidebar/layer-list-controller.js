angular.module('magna-app')

.controller('LayerListCtrl', ['$scope', '$modal', 'LayerService',
  function($scope, $modal, LayerService) {
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
      var modalInstance = $modal.open({
        templateUrl: 'src/edit-layer/edit-layer-template.html',
        controller: 'EditLayerCtrl',
        resolve: {
          layer: function () {
            return layer;
          }
        }
      });

      modalInstance.result.then(function (item) {
        var layerIdx = $scope.layers.indexOf(layer);
        $scope.layers[layerIdx] = item;
      });
    };
}]);

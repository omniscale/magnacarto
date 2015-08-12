angular.module('magna-app')

.controller('RemoveLayerCtrl', ['$scope', '$modalInstance', 'layer',
  function($scope, $modalInstance, layer) {
    $scope.layer = layer;

    $scope.ok = function () {
      $modalInstance.close(true);
    };
    $scope.cancel = function () {
      $modalInstance.dismiss('cancel');
    };
}]);
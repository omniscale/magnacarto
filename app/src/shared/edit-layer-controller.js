angular.module('magna-app')

.controller('EditLayerCtrl', ['$scope', '$modalInstance', 'layer',
  function($scope, $modalInstance, layer) {
    $scope.form = {};
    $scope.layer = angular.copy(layer);

    $scope.ok = function () {
      if ($scope.layerForm.$invalid) {
        return false;
      }
      $modalInstance.close($scope.layer);
    };

    $scope.cancel = function () {
      $modalInstance.dismiss('cancel');
    };
  }
]);
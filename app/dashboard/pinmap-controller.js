angular.module('magna-app')

.controller('PinMapCtrl', ['$scope', '$modalInstance', 'map',
  function($scope, $modalInstance, map) {
    $scope.form = {};
    $scope.map = map;
    $scope.title = '';

    $scope.ok = function () {
      if ($scope.pinmapForm.$invalid) {
        return false;
      }
      var item = {
        'coords': $scope.map.coords,
        'zoom': $scope.map.zoom,
        'title': $scope.title
      };
      $modalInstance.close(item);
    };

    $scope.cancel = function () {
      $modalInstance.dismiss('cancel');
    };
  }
]);
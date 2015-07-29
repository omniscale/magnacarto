angular.module('magna-app')

.controller('EditLayerCtrl', ['$scope', '$modalInstance', 'layer',
  function($scope, $modalInstance, layer) {
    $scope.form = {};
    $scope.layer = angular.copy(layer);

    $scope.datasourceTemplates = {
      'raster': 'src/edit-layer/raster-datasource-template.html',
      'sqlite': 'src/edit-layer/sqlite-datasource-template.html',
      'postgis': 'src/edit-layer/postgis-datasource-template.html',
      'shape': 'src/edit-layer/shape-datasource-template.html'
    };

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
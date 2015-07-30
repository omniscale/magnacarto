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

    var cleanupDatasource = function(datasource) {
      switch(datasource.type) {
        case 'postgis':
          return {
            'type': datasource.type,
            'dbname': datasource.dbname,
            'user': datasource.user,
            'password': datasource.password,
            'extent': datasource.extent,
            'extent_cache': datasource.extent_cache,
            'geometry_field': datasource.geometry_field,
            'key_field': datasource.key_field,
            'srid': datasource.srid,
            'table': datasource.table
          };
        case 'sqlite':
          return {
            'type': datasource.type,
            'file': datasource.file,
            'attachdb' : datasource.attachdb,
            'extent': datasource.extent,
            'geometry_field': datasource.geometry_field,
            'key_field': datasource.key_field,
            'srid': datasource.srid,
            'table': datasource.table
          };
        case 'shape':
          return  {
            'type': datasource.type,
            'file': datasource.file
          };
        case 'gdal':
          return {
            'type': datasource.type,
            'file': datasource.file,
            'band': datasource.band
          };
        default:
          return datasource;
      }

    };

    $scope.ok = function () {
      if ($scope.layerForm.$invalid) {
        return false;
      }
      $scope.layer.Datasource = cleanupDatasource($scope.layer.Datasource);
      $modalInstance.close($scope.layer);
    };

    $scope.cancel = function () {
      $modalInstance.dismiss('cancel');
    };
  }
]);
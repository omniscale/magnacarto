angular.module('magna-app')

.controller('EditLayerCtrl', ['$scope', 'magnaConfig', 'LayerService', 'EditLayerFormStatusService', '$modal', '$modalInstance', 'layer',
  function($scope, magnaConfig, LayerService, EditLayerFormStatusService, $modal, $modalInstance, layer) {
    $scope.layer = angular.copy(layer);
    $scope.layers = LayerService.layers;

    var uniquePush = function(array, value) {
      if(angular.isUndefined(value)) {
        return;
      }
      if(angular.isString(value)) {
        value = value.trim();
        if(value === '') {
          return;
        }
      }
      if(array.indexOf(value) === -1) {
        array.push(value);
      }
    };

    var prepareSuggestions = function(defaults) {
      if(angular.isUndefined(defaults)) {
        return [];
      }
      var target = angular.copy(defaults);
      if(target.length > 0) {
        target.push('_separator_');
      }
      return target;
    };

    $scope.srsSuggestions = prepareSuggestions(magnaConfig.defaultSuggestions.srs);
    $scope.extentSuggestions = prepareSuggestions(magnaConfig.defaultSuggestions.extent);
    $scope.geometryFieldSuggestions = prepareSuggestions(magnaConfig.defaultSuggestions.geometry_field);
    $scope.sridSuggestions = prepareSuggestions(magnaConfig.defaultSuggestions.srid);
    $scope.dbHostSuggestions = prepareSuggestions(magnaConfig.defaultSuggestions.host);
    $scope.dbPortSuggestions = prepareSuggestions(magnaConfig.defaultSuggestions.port);
    $scope.dbNameSuggestions = prepareSuggestions(magnaConfig.defaultSuggestions.dbname);
    $scope.dbUserSuggestions = prepareSuggestions(magnaConfig.defaultSuggestions.user);
    $scope.dbPasswordSuggestions = prepareSuggestions(magnaConfig.defaultSuggestions.password);

    angular.forEach($scope.layers, function(_layer) {
      if(_layer.srs !== undefined && _layer.srs !== '') {
        uniquePush($scope.srsSuggestions, _layer.srs);
        if(_layer.Datasource !== undefined && _layer.Datasource.type !== undefined) {
          if(_layer.Datasource.type === 'postgis' || _layer.Datasource.type === 'sqlite') {
              uniquePush($scope.extentSuggestions, _layer.Datasource.extent);
              uniquePush($scope.geometryFieldSuggestions, _layer.Datasource.geometry_field);
              uniquePush($scope.sridSuggestions, _layer.Datasource.srid);
          }
          if(_layer.Datasource.type === 'postgis') {
              uniquePush($scope.dbHostSuggestions, _layer.Datasource.host);
              uniquePush($scope.dbPortSuggestions, _layer.Datasource.port);
              uniquePush($scope.dbNameSuggestions, _layer.Datasource.dbname);
              uniquePush($scope.dbUserSuggestions, _layer.Datasource.user);
              uniquePush($scope.dbPasswordSuggestions, _layer.Datasource.password);
          }
        }
      }
    });
    $scope.isNewLayer = LayerService.isDefaultLayer(layer);

    $scope.datasourceTemplates = {
      'postgis': 'src/layer/datasource-postgis-template.html',
      'shape': 'src/layer/datasource-shape-template.html',
      'sqlite': 'src/layer/datasource-sqlite-template.html',
      'ogr': 'src/layer/datasource-ogr-template.html',
      'gdal': 'src/layer/datasource-gdal-template.html',
      'geojson': 'src/layer/datasource-geojson-template.html'
    };

    $scope.aceOptions = {
      mode: 'sql'
    };

    $scope.hideGeneral = EditLayerFormStatusService.hideGeneral();
    $scope.hideExtentSRS = EditLayerFormStatusService.hideExtentSRS();
    $scope.hideDatasource = EditLayerFormStatusService.hideDatasource();

    $scope.$watch('hideGeneral', function(hideGeneral) {
      EditLayerFormStatusService.hideGeneral(hideGeneral);
    });

    $scope.$watch('hideExtentSRS', function(hideExtentSRS) {
      EditLayerFormStatusService.hideExtentSRS(hideExtentSRS);
    });

    $scope.$watch('hideDatasource', function(hideDatasource) {
      EditLayerFormStatusService.hideDatasource(hideDatasource);
    });

    // calculate height of sql query input
    var text = '';
    if(angular.isDefined($scope.layer.Datasource)) {
      text = $scope.layer.Datasource.table || '';
    }
    var lines = Math.min(text.split(/\r*\n/).length + 4, 35);
    $scope.tableInputHeight = Math.ceil(lines * 14);

    var cleanupDatasource = function(datasource) {
      switch(datasource.type) {
        case 'postgis':
          return {
            'type': datasource.type,
            'host': datasource.host,
            'port': datasource.port,
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
            'file': datasource.file,
            'srid': datasource.srid
          };
        case 'gdal':
          return {
            'type': datasource.type,
            'file': datasource.file,
            'band': datasource.band
          };
        case 'ogr':
          return {
            'type': datasource.type,
            'file': datasource.file,
            'srid': datasource.srid,
            'layer': datasource.layer,
            'layer_by_sql': datasource.layer_by_sql,
            'extent': datasource.extent
          };
        case 'geojson':
          return {
            'type': datasource.type,
            'file': datasource.file
          };
        default:
          return datasource;
      }

    };

    $scope.setId = function() {
      if($scope.layer.name !== undefined && ($scope.layer.id === undefined || $scope.layer.id === '')) {
        $scope.layer.id = $scope.layer.name.toLowerCase().replace(new RegExp(/[^\w\d_]/g), '_');
      }
    };

    $scope.copyLayer = function(layer) {
      $scope.layer = angular.copy(layer);
      $scope.layer.name += '-copy';
      $scope.layer.id += '-copy';
      $scope.isNewLayer = false;
    };

    $scope.ok = function () {
      if ($scope.layerForm.$invalid) {
        return false;
      }
      $scope.layer.Datasource = cleanupDatasource($scope.layer.Datasource);
      $modalInstance.close($scope.layer);
    };

    $scope.openRemoveModal = function() {
      var modalInstance = $modal.open({
        templateUrl: 'src/layer/remove-layer-template.html',
        controller: 'RemoveLayerCtrl',
        backdrop: 'static',
        resolve: {
          layer: function () {
            return layer;
          }
        }
      });

      modalInstance.result.then(function () {
        $modalInstance.close('remove');
      });
    };

    $scope.cancel = function () {
      $modalInstance.dismiss('cancel');
    };
  }
]);
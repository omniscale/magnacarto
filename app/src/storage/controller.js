angular.module('magna-app')


.controller('StorageCtrl', ['$scope', '$location', 'MMLService', 'DashboardService', 'StyleService',
  function($scope, $location, MMLService, DashboardService, StyleService) {
    if(!MMLService.loaded()) {
      $location.path('/');
    }
    $scope.navItemName = 'storage';

    $scope.styles = StyleService.activeStyles;

    $scope.maps = MMLService.storedMaps;

    $scope.gridsterStorageOptions = {
      margins: [5, 5],
      resizable: {
        enabled: false
      },
      draggable: {
        enabled: false
      }
    };

    $scope.restore = function(map) {
      DashboardService.addMap({
        coords: map.coords,
        zoom: map.zoom
      });
    };

    $scope.remove = function(map) {
      angular.forEach($scope.maps, function(value, key) {
        if (angular.equals(map.id, value.id)) {
          $scope.maps.splice(key, 1);
        }
      });
    };
  }
]);

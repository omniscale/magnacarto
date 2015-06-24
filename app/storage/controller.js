angular.module('magna-app')


.controller('StorageCtrl', ['$scope', '$timeout', '$cookieStore', 'DashboardService',
  function($scope, $timeout, $cookieStore, DashboardService) {
    $scope.navItemName = 'storage';
    // TODO JSON
    var savedMaps = $cookieStore.get('savedMaps');

    $scope.layers = DashboardService.layers;

    $scope.maps = savedMaps;
    // $scope.dashboard
    $scope.gridsterStorageOptions = {
      margins: [5, 5],
      resizable: {
        enabled: false
      },
      draggable: {
        enabled: false
      }
    };

    // TODO JSON
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

      // TODO JSON
      $cookieStore.put('savedMaps', $scope.maps);
    };

    $scope.$on('gridster-item-initialized', function(){
      $timeout(function(){
        $scope.$broadcast('gridUpdate');
      });
    });

  }
]);

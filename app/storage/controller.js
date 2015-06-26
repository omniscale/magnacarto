angular.module('magna-app')


.controller('StorageCtrl', ['$scope', '$timeout', '$cookieStore', 'DashboardService', 'StyleService',
  function($scope, $timeout, $cookieStore, DashboardService, StyleService) {
    $scope.navItemName = 'storage';
    // TODO JSON
    var savedMaps = $cookieStore.get('savedMaps');

    $scope.styles = StyleService.activeStyles;

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

    $scope.$watch(function() {
      return angular.element(document.querySelector('.gridster-element')).attr('class');
    }, function(classes){
      if ((classes.indexOf('gridster-loaded')) > -1) {
        // trigger updateSize in ol3-directive
        $scope.$broadcast('gridInit');

        $scope.$on('gridster-item-initialized', function(){
          $timeout(function(){
            $scope.$broadcast('gridUpdate');
          }, 0);
        });
      }
    });

  }
]);

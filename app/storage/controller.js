angular.module('magna-app')


.controller('StorageCtrl', ['$scope', '$timeout', '$cookieStore',
  function($scope, $timeout, $cookieStore) {
    var savedMaps = $cookieStore.get('savedMaps');

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

    $scope.$watch(function() {
      return angular.element(document.querySelector('.gridster-element')).attr('class');
      }, function(classes){
        if ((classes.indexOf('gridster-loaded')) > -1) {
          $scope.$broadcast('gridUpdate');
        }
    });

    $scope.$on('gridster-item-transition-end', function(){
      $scope.$broadcast('gridUpdate');
    });

    $scope.$on('gridster-item-size-changed', function(){
      $scope.$broadcast('gridUpdate');
    });

    // TODO JSON
    $scope.restore = function(map) {
      var dashboardItems = $cookieStore.get('magnatorDashboard');
      if (dashboardItems === undefined) {
        dashboardItems = [];
      }

      dashboardItems.push({
        sizeX: 1,
        sizeY: 1,
        coords: map.coords,
        zoom: map.zoom
      });

      // TODO JSON show popup that dashboard item was saved
      $cookieStore.put('magnatorDashboard', dashboardItems);
      // TODO: check if there is a nicer way then $scope.$apply();
      $scope.$apply();
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

  }
]);

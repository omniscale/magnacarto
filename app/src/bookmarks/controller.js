angular.module('magna-app')


.controller('BookmarksCtrl', ['$scope', '$location', 'ProjectService', 'DashboardService', 'StyleService',
  function($scope, $location, ProjectService, DashboardService, StyleService) {
    if(ProjectService.projectLoaded() === undefined) {
      $location.path('/');
      return;
    }
    $scope.navItemName = 'storage';

    $scope.styles = StyleService.activeStyles;

    $scope.maps = ProjectService.storedMaps;

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

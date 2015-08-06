angular.module('magna-app')


.controller('BookmarksCtrl', ['$scope', 'ProjectService', 'DashboardService', 'StyleService',
  function($scope, ProjectService, DashboardService, StyleService) {
    $scope.maps = ProjectService.bookmarkedMaps;
    $scope.styles = StyleService.activeStyles;

    $scope.navItemName = 'bookmarks';

    $scope.styles = StyleService.activeStyles;

    $scope.maps = ProjectService.bookmarkedMaps;

    $scope.gridsterOptions = {
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

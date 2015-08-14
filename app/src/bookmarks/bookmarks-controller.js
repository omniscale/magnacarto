angular.module('magna-app')


.controller('BookmarksCtrl', ['$scope', 'ProjectService', 'DashboardService', 'StyleService', 'SideNavService',
  function($scope, ProjectService, DashboardService, StyleService, SideNavService) {
    $scope.maps = ProjectService.bookmarkedMaps;
    $scope.styles = StyleService.activeStyles;

    SideNavService.currentPage('bookmarks');

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

    $scope.$watchCollection(function() {
      return ProjectService.bookmarkedMaps;
    }, function() {
      $scope.maps = ProjectService.bookmarkedMaps;
    });
  }
]);

angular.module('magna-app')


.controller('BookmarksCtrl', ['$scope', '$location', 'magnaConfig', 'ProjectService', 'DashboardService', 'StyleService', 'SideNavService',
  function($scope, $location, magnaConfig, ProjectService, DashboardService, StyleService, SideNavService) {
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
        handle: '.move-bookmark'
      }
    };

    $scope.restore = function(event, map) {
      DashboardService.addMap({
        coords: map.coords,
        zoom: map.zoom
      });
      if(!event[magnaConfig.selectMultipleBookmarksKey]) {
        $location.path('dashboard/' + ProjectService.project.url);
      }
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

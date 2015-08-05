angular.module('magna-app')

.controller('DashboardCtrl', ['$scope', '$location', '$routeParams', 'ProjectsService', 'DashboardService', 'StyleService', 'ProjectService',
  function($scope, $location, $routeParams, ProjectsService, DashboardService, StyleService, ProjectService) {
    var project = ProjectsService.projectByRouteParams($routeParams);

    if(project === undefined) {
      $location.path('projects');
      return;
    }

    if(ProjectService.projectLoaded() === undefined || ProjectService.project !== project) {
      var loadedPromise = ProjectService.loadProject(project);
      loadedPromise.then(function() {
        $scope.maps = DashboardService.maps;
        $scope.styles = StyleService.activeStyles;
      });
    }

    $scope.navItemName = 'dashboard';
    $scope.gridsterOptions = {
      margins: [5, 5],
      columns: 8,
      swapping: true,
      floating: true,
      resizable: {
        stop: function(event, uiWidget) {
          // resizeMap added by ol3-directive to gridster item scope
          uiWidget.scope().resizeMap();
        }
      },
      draggable: {
        handle: '.move-map'
      }
    };

    $scope.maps = DashboardService.maps;
    $scope.styles = StyleService.activeStyles;

    // Need to watch otherwise $scope.maps and DashboardService.maps are
    // different objects after clear maps and changes not recognized by
    // andgular
    $scope.$watchCollection(function() {
      return DashboardService.maps;
    }, function() {
      $scope.maps = DashboardService.maps;
    });
  }
])

.controller('DashboardMapCtrl', ['$scope', '$modal', 'DashboardService', 'ProjectService',
  function($scope, $modal, DashboardService, ProjectService) {

    $scope.openBookmarkModal = function (map) {
      var modalInstance = $modal.open({
        templateUrl: 'src/dashboard/bookmark-map-template.html',
        controller: 'BookmarkMapCtrl',
        backdrop: 'static',
        resolve: {
          map: function () {
            return map;
          }
        }
      });

      modalInstance.result.then(function (item) {
        // TODO: add antoher function to create an unique id
        var id = item.coords[0];
        id = id.toString();
        item.id = id.replace(/\./g,'');

        ProjectService.bookmarkedMaps.push(item);
      });
    };

    $scope.remove = function(map) {
      DashboardService.removeMap(map);
    };
  }
]);

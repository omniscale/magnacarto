angular.module('magna-app')

.controller('DashboardCtrl', ['$scope', '$location', 'DashboardService', 'StyleService', 'MMLService',
  function($scope, $location, DashboardService, StyleService, MMLService) {
    if(MMLService.projectLoaded() === undefined) {
      $location.path('/');
      return;
    }
    $scope.navItemName = 'dashboard';
    $scope.gridsterOptions = {
      margins: [5, 5],
      columns: 4,
      swapping: true,
      floating: true,
      resizable: {
        stop: function(event, uiWidget) {
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

.controller('DashboardMapCtrl', ['$scope', '$modal', 'DashboardService', 'MMLService',
  function($scope, $modal, DashboardService, MMLService) {

    $scope.openSaveModal = function (map) {
      var modalInstance = $modal.open({
        templateUrl: 'src/dashboard/pinmap.template.html',
        controller: 'PinMapCtrl',
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

        MMLService.storedMaps.push(item);
      });
    };

    $scope.remove = function(map) {
      DashboardService.removeMap(map);
    };
  }
]);

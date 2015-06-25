angular.module('magna-app')

.controller('DashboardCtrl', ['$scope', '$timeout', '$cookieStore', 'DashboardService', 'StyleService',
  function($scope, $timeout, $cookieStore, DashboardService, StyleService) {
    $scope.navItemName = 'dashboard';
    $scope.gridsterOptions = {
      margins: [5, 5],
      columns: 4,
      swapping: true,
      floating: true,
      resizable: {
        stop: function() {
          $timeout(function() {
            $scope.$broadcast('gridUpdate');
          }, 0);
        }
      },
      draggable: {
        handle: '.move-map',
        stop: function() {
          $timeout(function() {
            $scope.$broadcast('gridUpdate');
          }, 0);
        }
      }
    };

    $scope.maps = DashboardService.maps;
    // Need to watch otherwise $scope.maps and DashboardService.maps are
    // different objects after clear maps and changes not recognized by
    // andgular
    $scope.$watchCollection(function() {
      return DashboardService.maps;
    }, function() {
      $scope.maps = DashboardService.maps;
    });

    $scope.styles = StyleService.activeStyles;

    $scope.$on('gridster-item-initialized', function(){
      $timeout(function(){
        $scope.$broadcast('gridUpdate');
      });
    });
  }
])

.controller('DashboardMapCtrl', ['$scope', '$cookieStore', '$modal', 'DashboardService',
  function($scope, $cookieStore, $modal, DashboardService) {

    $scope.openSaveModal = function (map) {
      var modalInstance = $modal.open({
        templateUrl: 'app/dashboard/pinmap.template.html',
        controller: 'PinMapCtrl',
        resolve: {
          map: function () {
            return map;
          }
        }
      });

      modalInstance.result.then(function (item) {
        var savedPlaces = [];
        var cookie = $cookieStore.get('savedMaps');
        if (cookie !== undefined && angular.isArray(cookie)) {
          savedPlaces = cookie;
        }

        // TODO: add antoher function to create an unique id
        var id = item.coords[0];
        id = id.toString();
        item.id = id.replace(/\./g,'');

        savedPlaces.push(item);
        // TODO: add Message that mat
        $cookieStore.put('savedMaps', savedPlaces);
      });
    };

    $scope.remove = function(map) {
      DashboardService.removeMap(map);
    };
  }
]);

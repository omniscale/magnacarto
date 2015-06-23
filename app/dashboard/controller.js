angular.module('magna-app')

.service('dashboard', function Dashboard($rootScope) {
  var dashboard = this;
  dashboard.mss = $rootScope.dashboard.mss;
  dashboard.layers = $rootScope.dashboard.layers;
  dashboard.maps = $rootScope.dashboard.maps;
})

.controller('DashboardCtrl', ['$scope', '$timeout', '$cookieStore', 'dashboard',
  function($scope, $timeout, $cookieStore, dashboard) {
    $scope.gridsterOptions = {
      margins: [5, 5],
      columns: 4,
      swapping: true,
      floating: true,
      draggable: {
        handle: '.move-map'
      }
    };
    $scope.dashboard = dashboard;

    // TODO JSON
    // LOAD
    var magnatorDashboardCookie = $cookieStore.get('magnatorDashboard');
    if (magnatorDashboardCookie !== undefined) {
      $scope.dashboard.maps = magnatorDashboardCookie;
    }

    // SAVE
    $scope.$watch('dashboard', function(items){
      $cookieStore.put('magnatorDashboard', items.maps);
    }, true);

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

    $scope.$on('gridster-item-resized', function(){
       $scope.$broadcast('gridUpdate');
    });

    $scope.$on(['gridster-item-initialized'], function(){
      $timeout(function(){
        $scope.$broadcast('gridUpdate');
      });
    });


    $scope.clearMaps = function() {
      $scope.dashboard.maps = [];
    };

    $scope.addMap = function() {
      var defaultCoords = [8,53];
      var defaultZoom = 12;
      var lastMap = dashboard.maps[dashboard.maps.length-1];
      if (lastMap !== undefined) {
        defaultCoords = lastMap.coords;
        defaultZoom = lastMap.zoom;
      }

      $scope.dashboard.maps.push({
        sizeX: 1,
        sizeY: 1,
        coords: defaultCoords,
        zoom: defaultZoom
      });
    };
  }
])

.controller('DashboardMapCtrl', ['$scope', '$cookieStore', '$modal',
  function($scope, $cookieStore, $modal) {

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
      $scope.dashboard.maps.splice($scope.dashboard.maps.indexOf(map), 1);
    };
  }
])


.controller('PinMapCtrl', ['$scope', '$modalInstance', 'map',
  function($scope, $modalInstance, map) {
    $scope.form = {};
    $scope.map = map;
    $scope.title = '';

    $scope.ok = function () {
      if ($scope.pinmapForm.$invalid) {
        return false;
      }
      var item = {
        'coords': $scope.map.coords,
        'zoom': $scope.map.zoom,
        'title': $scope.title
      };
      $modalInstance.close(item);
    };

    $scope.cancel = function () {
      $modalInstance.dismiss('cancel');
    };
    }
]);


angular.module('magna-app')

.provider('DashboardService', [function() {
  this.$get = ['$rootScope', '$timeout', '$cookieStore', 'magnaConfig',
    function($rootScope, $timeout, $cookieStore, magnaConfig) {
      var DashboardServiceInstance = function() {
        var self = this;
        self.mss = [];
        self.activeMss = [];
        // stores ol3 layers
        self.layers = [];
        self.maps = [];

        // dashboard cookie block. Replace by JSON
        var magnatorDashboardCookie = $cookieStore.get('magnatorDashboard');
        if (magnatorDashboardCookie !== undefined) {
          self.maps = magnatorDashboardCookie;
        }
        $rootScope.$watch(function() {
          return self.maps;
        }, function() {
          $cookieStore.put('magnatorDashboard', self.maps);
        }, true);
      };

      DashboardServiceInstance.prototype.setMss = function(mss) {
        var self = this;
        self.mss = mss;
        // need to copy otherwise mss and activeMss are the same
        // list object
        self.activeMss = angular.copy(mss);
      };

      DashboardServiceInstance.prototype.toggleMss = function(mss) {
        var self = this;
        var idx = self.activeMss.indexOf(mss);
        if (idx > -1) {
          self.activeMss.splice(idx, 1);
        } else {
          self.activeMss.push(mss);
        }
      };


      DashboardServiceInstance.prototype.clearMaps = function() {
        var self = this;
        self.maps = [];
      };

      // TODO add map param to add map from store
      DashboardServiceInstance.prototype.addMap = function(map) {
        var self = this;
        var coords = magnaConfig.defaultCenter;
        var zoom = magnaConfig.defaultZoom;

        if(map !== undefined) {
          coords = map.coords;
          zoom = map.zoom;
        } else if(self.maps.length > 0) {
          var lastMap = self.maps[self.maps.length - 1];
          coords = lastMap.coords;
          zoom = lastMap.zoom;
        }

        self.maps.push({
          sizeX: 1,
          sizeY: 1,
          coords: coords,
          zoom: zoom
        });
      };

      DashboardServiceInstance.prototype.removeMap = function(map) {
        var self = this;
        var idx = self.maps.indexOf(map);
        if (idx > -1) {
          self.maps.splice(idx, 1);
        }
      };

      return new DashboardServiceInstance();
    }];
}]);

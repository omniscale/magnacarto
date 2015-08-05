angular.module('magna-app')

.provider('DashboardService', [function() {
  this.$get = ['magnaConfig',
    function(magnaConfig) {
      var DashboardServiceInstance = function() {
        var self = this;
        self.maps = [];
      };

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
          sizeX: 4,
          sizeY: 3,
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

angular.module('magna-app')

.provider('DashboardService', [function() {
  this.$get = ['$rootScope', '$timeout', '$cookieStore', 'magnaConfig',
    function($rootScope, $timeout, $cookieStore, magnaConfig) {
      var DashboardServiceInstance = function() {
        var self = this;
        self.styles = [];
        self.activeStyles = [];
        // TODO check if we can gat rid of layers, see ol3-direcitve
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

      DashboardServiceInstance.prototype.setStyles = function(styles) {
        var self = this;
        self.styles = styles;
        // need to copy otherwise styles and activeStyles are the same
        // list object
        self.activeStyles = angular.copy(styles);
      };

      DashboardServiceInstance.prototype.toggleStyle = function(style) {
        var self = this;
        var idx = self.activeStyles.indexOf(style);
        if (idx > -1) {
          self.activeStyles.splice(idx, 1);
        } else {
          self.activeStyles.push(style);
        }
      };

      DashboardServiceInstance.prototype.clearMaps = function() {
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

angular.module('magna-app')

.provider('MMLService', [function() {
  this.$get = ['$http', '$rootScope', '$websocket', 'magnaConfig', 'StyleService', 'DashboardService',
    function($http, $rootScope, $websocket, magnaConfig, StyleService, DashboardService) {
      var MMLServiceInstance = function() {
        this.mml = undefined;
        this.mmlData = undefined;
        this.styles = [];
        this.dashboardMaps = [];
        this.storedMaps = [];
        this.socketUrl = undefined;
        this.socket = undefined;
      };

      MMLServiceInstance.prototype.loadProject = function(project) {
        var self = this;

        self.disableWatchers();
        // unbind socket
        if(self.socket !== undefined) {
          self.socket.$close();
        }
        // clear project data
        self.dashboardMaps = [];
        self.storedMaps = [];

        self.basePath = project.base;
        self.mml = project.mml;
        self.availableMss = project.available_mss;

        self.loadPromise = $http.get(magnaConfig.projectBaseUrl + self.basePath + '/' + self.mml);
        // TODO add load project error handling
        self.loadPromise = self.loadPromise.then(function(response) {
          self.mmlData = response.data;
          self.bindSocket_();
          StyleService.setStyles(self.availableMss);
          StyleService.setProjectStyles(self.mmlData.Stylesheet);

          if(self.mmlData.magnacarto === undefined) {
            self.mmlData.magnacarto = {
              dashboardMaps: [],
              storedMaps: []
            };
          }

          // assign to object property for easy access from outside;
          self.storedMaps = self.mmlData.magnacarto.storedMaps;


          DashboardService.maps = self.mmlData.magnacarto.dashboardMaps;
          DashboardService.layers = [{
            styles: StyleService.activeStyles,
            mml: self.mml
          }];

          self.enableWatchers();

        });

        return self.loadPromise;
      };

      MMLServiceInstance.prototype.saveProject = function() {
        var self = this;
        $http.post(magnaConfig.projectBaseUrl + self.basePath + '/' + self.mml, self.mmlData);
      };

      MMLServiceInstance.prototype.bindSocket_ = function() {
        var self = this;
        self.socketUrl = magnaConfig.socketUrl;
        self.socketUrl += 'mml=' + self.mml;
        self.socketUrl += '&mss=' + self.availableMss;
        self.socketUrl += '&base=' + self.basePath;

        self.socket = $websocket.$new({
          url: self.socketUrl,
          reconnect: true,
          reconnectInterval: 100
        });
      };

      MMLServiceInstance.prototype.loaded = function() {
        var self = this;
        return self.loadPromise !== undefined;
      };

      MMLServiceInstance.prototype.getSocket = function() {
        return this.socket;
      };

      MMLServiceInstance.prototype.enableWatchers = function() {
        var self = this;

        // listen on changes in dashboardMaps
        // save if change occurs
        self.dashboardMapsWatcher = $rootScope.$watch(function() {
          return self.mmlData.magnacarto.dashboardMaps;
        }, function(n, o) {
          if(n === o) return;
          self.saveProject();
        }, true);

        // listen on changes in storedMaps
        // save if change occurs
        self.storedMapsWatcher = $rootScope.$watch(function() {
          return self.storedMaps;
        }, function(n, o) {
          if(n === o) return;
          self.saveProject();
        }, true);

        self.stylesWatcher = $rootScope.$watch(function() {
          return self.mmlData.Stylesheet;
        }, function(n, o) {
          if(n === o) return;
          self.saveProject();
        }, true);
      };

      MMLServiceInstance.prototype.disableWatchers = function() {
        var self = this;
        if(self.dashboardMapsWatcher !== undefined) {
          self.dashboardMapsWatcher();
          self.dashboardMapsWatcher = undefined;
        }

        if(self.storedMapsWatcher !== undefined) {
          self.storedMapsWatcher();
          self.storedMapsWatcher = undefined;
        }

        if(self.stylesWatcher !== undefined) {
          self.stylesWatcher();
          self.stylesWatcher = undefined;
        }
      };

      return new MMLServiceInstance();
  }];
}]);

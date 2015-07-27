angular.module('magna-app')

.provider('MMLService', [function() {

  var baseUrl;

  var fakePost = function(url, data) {
    console.log(url, data);
  };

  this.setBaseUrl = function(url) {
    baseUrl = url;
  };

  this.$get = ['$http', '$rootScope', '$websocket', 'magnaConfig', 'StyleService', 'DashboardService',
    function($http, $rootScope, $websocket, magnaConfig, StyleService, DashboardService) {
      var MMLServiceInstance = function(baseUrl) {
        this.mml = undefined;
        this.mmlData = undefined;
        this.styles = [];
        this.baseUrl = baseUrl;
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
        self.styles = [];
        self.dashboardMaps = [];
        self.storedMaps = [];

        self.basePath = project.base;
        self.mml = project.mml;
        self.availableMss = project.available_mss;

        self.loadPromise = $http.get(self.baseUrl + self.basePath + '/' + self.mml);
        // TODO add load project error handling
        self.loadPromise = self.loadPromise.then(function(response) {
          var data = response.data;
          self.bindSocket_();
          self.styles = data.Stylesheet;
          StyleService.setStyles(self.availableMss);
          StyleService.setProjectStyles(self.styles);

          if(data.magnacarto !== undefined) {
            self.dashboardMaps = data.magnacarto.dashboardMaps;
            self.storedMaps = data.magnacarto.storedMaps;
          }

          DashboardService.maps = self.dashboardMaps;
          DashboardService.layers = [{
            styles: StyleService.activeStyles,
            mml: self.mml
          }];

          self.enableWatchers();

        });

        return self.loadPromise;
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

      MMLServiceInstance.prototype.saveActiveStyles = function() {
        var self = this;
        fakePost(self.saveUrl, {
          'type': 'styles',
          'styles': self.styles
        });
      };

      MMLServiceInstance.prototype.saveDashboardMaps = function() {
        var self = this;
        fakePost(self.saveUrl, {
          'type': 'dashboardMaps',
          'maps': self.dashboardMaps
        });
        // TODO readd when we have a real endpoint
        // self.saveDashboardMapsPromise = $http.post(self.saveDashboardMapsUrl, self.dashboardMaps);
        // return self.saveDashboardMapsPromise;
      };

      MMLServiceInstance.prototype.saveStoredMaps = function() {
        var self = this;
        fakePost(self.saveUrl, {
          'type': 'storedMaps',
          'maps': self.storedMaps
        });
        // TODO readd when we have a real endpoint
        // self.saveStoredMapsPromise = $http.post(self.saveStoredMapsUrl, self.storedMaps);
        // return self.saveStoredMapsPromise;
      };

      MMLServiceInstance.prototype.enableWatchers = function() {
        var self = this;

        // listen on changes in dashboardMaps
        // save if change occurs
        self.dashboardMapsWatcher = $rootScope.$watch(function() {
          return self.dashboardMaps;
        }, function(o, n) {
          if(n === o) return;
          self.saveDashboardMaps();
        }, true);

        // listen on changes in storedMaps
        // save if change occurs
        self.storedMapsWatcher = $rootScope.$watch(function() {
          return self.storedMaps;
        }, function(o, n) {
          if(n === o) return;
          self.saveStoredMaps();
        }, true);

        self.stylesWatcher = $rootScope.$watch(function() {
          return self.styles;
        }, function(o, n) {
          if(n === o) return;
          self.saveActiveStyles();
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

      return new MMLServiceInstance(baseUrl);
  }];
}]);

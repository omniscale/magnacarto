angular.module('magna-app')
/* Todo rename to ProjectServicev */
.provider('MMLService', [function() {
  this.$get = ['$http', '$rootScope', '$q', '$websocket', 'magnaConfig', 'StyleService', 'LayerService', 'DashboardService',
    function($http, $rootScope, $q, $websocket, magnaConfig, StyleService, LayerService, DashboardService) {
      var MMLServiceInstance = function() {
        this.mml = undefined;
        this.mmlData = undefined;
        this.dashboardMaps = [];
        this.storedMaps = [];
        this.socketUrl = undefined;
        this.socket = undefined;
        this.projectLoadedPromise = undefined;
      };

      MMLServiceInstance.prototype.loadProject = function(project) {
        var self = this;

        self.unloadProject();

        self.basePath = project.base;
        self.mml = project.mml;
        self.mcp = project.mcp;
        self.availableMss = project.available_mss;

        var mmlLoadPromise = $http.get(magnaConfig.projectBaseUrl + self.basePath + '/' + self.mml);
        var mcpLoadPromise = $http.get(magnaConfig.projectBaseUrl + self.basePath + '/' + self.mcp);

        self.projectLoadedPromise = $q.all([mmlLoadPromise, mcpLoadPromise]).then(function(data){
          self.handleMMLResponse(data[0].data);
          self.handleMCPResponse(data[1].data);

          self.bindSocket();
          self.enableWatchers();
        });

        return self.projectLoadedPromise;
      };

      MMLServiceInstance.prototype.handleMMLResponse = function(response) {
        var self = this;
        self.mmlData = response;
        StyleService.setStyles(self.availableMss);
        StyleService.setProjectStyles(self.mmlData.Stylesheet);

        LayerService.setLayers(self.mmlData.Layer);
      };

      MMLServiceInstance.prototype.handleMCPResponse = function(response) {
        var self = this;
        response.dashboardMaps = response.dashboardMaps || [];
        response.storedMaps = response.storedMaps || [];
        self.mcpData = response;

        // assign to object property for easy access from outside;
        self.storedMaps = self.mcpData.storedMaps;
        DashboardService.maps = self.mcpData.dashboardMaps;
      };

      MMLServiceInstance.prototype.unloadProject = function() {
        var self = this;
        if(self.mmlData === undefined) {
          return;
        }

        self.disableWatchers();
        if(self.socket !== undefined) {
          self.socket.$close();
        }

        self.mml = undefined;
        self.mmlData = undefined;
        self.dashboardMaps = [];
        self.storedMaps = [];
        self.socketUrl = undefined;
        self.socket = undefined;
        self.mmlLoadPromise = undefined;

        DashboardService.maps = [];
        StyleService.setStyles([]);
        StyleService.setProjectStyles([]);
        LayerService.setLayers([]);
      };

      MMLServiceInstance.prototype.saveMML = function() {
        var self = this;
        $http.post(magnaConfig.projectBaseUrl + self.basePath + '/' + self.mml, angular.toJson(self.mmlData, true));
      };

      MMLServiceInstance.prototype.saveMCP = function() {
        var self = this;
        $http.post(magnaConfig.projectBaseUrl + self.basePath + '/' + self.mcp, angular.toJson(self.mcpData, true));
      };

      MMLServiceInstance.prototype.bindSocket = function() {
        var self = this;
        self.socketUrl = angular.copy(magnaConfig.socketUrl);
        self.socketUrl += 'mml=' + self.mml;
        self.socketUrl += '&mss=' + self.availableMss;
        self.socketUrl += '&base=' + self.basePath;

        self.socket = $websocket.$new({
          url: self.socketUrl,
          reconnect: true,
          reconnectInterval: 100
        });
      };

      MMLServiceInstance.prototype.projectLoaded = function() {
        var self = this;
        return self.projectLoadedPromise;
      };

      MMLServiceInstance.prototype.getSocket = function() {
        return this.socket;
      };

      MMLServiceInstance.prototype.enableWatchers = function() {
        var self = this;

        // listen on changes in dashboardMaps
        // save if change occurs
        self.dashboardMapsWatcher = $rootScope.$watch(function() {
          return self.mcpData.dashboardMaps;
        }, function(n, o) {
          if(n === o) return;
          self.saveMCP();
        }, true);

        // listen on changes in storedMaps
        // save if change occurs
        self.storedMapsWatcher = $rootScope.$watch(function() {
          return self.storedMaps;
        }, function(n, o) {
          if(n === o) return;
          self.saveMCP();
        }, true);

        self.stylesWatcher = $rootScope.$watch(function() {
          return self.mmlData.Stylesheet;
        }, function(n, o) {
          if(n === o) return;
          self.saveMML();
        }, true);

        self.layersWatcher = $rootScope.$watch(function() {
          return self.mmlData.Layer;
        }, function(n, o) {
          if(n === o) return;
          self.saveMML();
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

        if(self.layersWatcher !== undefined) {
          self.layersWatcher();
          self.layersWatcher = undefined;
        }
      };

      return new MMLServiceInstance();
  }];
}]);

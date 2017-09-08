angular.module('magna-app')
/* Todo rename to ProjectServicev */
.provider('ProjectService', [function() {
  this.$get = ['$http', '$rootScope', '$q', '$timeout', '$websocket', 'magnaConfig', 'StyleService', 'LayerService', 'DashboardService',
    function($http, $rootScope, $q, $timeout, $websocket, magnaConfig, StyleService, LayerService, DashboardService) {
      var ProjectServiceInstance = function() {
        this.project = undefined;
        this.mmlData = undefined;
        this.dashboardMaps = [];
        this.bookmarkedMaps = [];
        this.appSettings = {};
        this.socketUrl = undefined;
        this.socket = undefined;
        this.mmlLoadPromise = undefined;
        this.mcpLoadPromise = undefined;
        this.projectLoadedPromise = undefined;
        this.mcpSaveTimeout = undefined;
        this.mapOptions = undefined;
      };

      ProjectServiceInstance.prototype.loadProject = function(project) {
        var self = this;

        self.unloadProject();

        self.project = project;

        self.mmlLoadPromise = self.loadMML();
        self.mcpLoadPromise = self.loadMCP();

        self.projectLoadedPromise = $q.all([self.mmlLoadPromise, self.mcpLoadPromise]).then(function(data){
          self.handleMMLResponse(data[0].data);
          self.handleMCPResponse(data[1].data);

          self.bindSocket();
          self.enableWatchers();
        });

        return self.projectLoadedPromise;
      };

      ProjectServiceInstance.prototype.loadMML = function() {
        var self = this;
        return $http.get(magnaConfig.projectBaseUrl + self.project.base + '/' + self.project.mml);
      };

      ProjectServiceInstance.prototype.loadMCP = function() {
        var self = this;
        return $http.get(magnaConfig.projectBaseUrl + self.project.base + '/' + self.project.mcp);
      };

      ProjectServiceInstance.prototype.handleMMLResponse = function(response) {
        var self = this;
        if(self.mmlData !== undefined) {
          // Clear array but keep reference to it.
          // If a = [] is used instead of a.length = 0, reference changes
          self.mmlData.Stylesheet.length = 0;
          angular.forEach(response.Stylesheet, function(style) {
            self.mmlData.Stylesheet.push(style);
          });
          self.mmlData.Layer.length = 0;
          angular.forEach(response.Layer, function(layer) {
            self.mmlData.Layer.push(layer);
          });
        } else {
          self.mmlData = response;
        }

        self.handleMapOptions(self.mmlData.Map);

        StyleService.setStyles(self.project.available_mss);
        StyleService.setProjectStyles(self.mmlData.Stylesheet);

        LayerService.setLayers(self.mmlData.Layer);
      };

      ProjectServiceInstance.prototype.handleMapOptions = function(mapOptions) {
        var self = this;
        if (mapOptions === undefined) {
          mapOptions = {
            'SRS': 'EPSG:3857',
            'BBOX': [-20026376.39,-20048966.10, 20026376.39, 20048966.10],
            'Resolutions': undefined,
            'ZoomScales': undefined
          };
        }

        // calculate resolution for map
        function getResolutionForScale(scale) {
          // TODO use 72 dpi for mapserver
          var dpi = 72;
          var inchesPerMeter = 100 / 2.54;
          return parseFloat(scale) / (inchesPerMeter * dpi);
        }

        if (mapOptions.ZoomScales !== undefined) {
          mapOptions.Resolutions = [];
          angular.forEach(mapOptions.ZoomScales, function(scale, key) {
            var resolution;
            if (key === 0) {
              resolution = getResolutionForScale(scale);
              resolution = resolution * Math.sqrt(2);
            } else {
              var scaleAverage = (scale + mapOptions.ZoomScales[key-1]) / 2;
              resolution = getResolutionForScale(scaleAverage);
            }
            mapOptions.Resolutions.push(resolution);
          });
        }

        if (mapOptions.SRS === undefined) {
          mapOptions.SRS = 'EPSG:3857';
        }
        if (mapOptions.DefaultCenter === undefined) {
          mapOptions.DefaultCenter = [
            mapOptions.BBOX[0] + (mapOptions.BBOX[2] - mapOptions.BBOX[0]),
            mapOptions.BBOX[1] + (mapOptions.BBOX[3] - mapOptions.BBOX[1])
          ];
        }

        if (mapOptions.DefaultZoom === undefined) {
          mapOptions.DefaultZoom = 2;
        }
        self.mapOptions = mapOptions;
        DashboardService.mapOptions = mapOptions;
      };

      ProjectServiceInstance.prototype.handleMCPResponse = function(response) {
        var self = this;
        response.dashboardMaps = response.dashboardMaps || [];
        response.bookmarkedMaps = response.bookmarkedMaps || [];
        response.appSettings = response.appSettings || {};
        // set to default -1 because resizer needs a defined value in order to update initial size.
        // -1 is matched to default value by resizer
        response.appSettings.sidebarWidth = response.appSettings.sidebarWidth || -1;
        response.appSettings.loggingHeight = response.appSettings.loggingHeight || -1;
        response.appSettings.sidebarCollapsed = response.appSettings.sidebarCollapsed || false;
        self.mcpData = response;

        // assign to object property for easy access from outside;
        self.bookmarkedMaps = self.mcpData.bookmarkedMaps;
        DashboardService.maps = self.mcpData.dashboardMaps;
        self.appSettings = self.mcpData.appSettings;
      };

      ProjectServiceInstance.prototype.unloadProject = function() {
        var self = this;
        if(self.mmlData === undefined) {
          return;
        }

        self.disableWatchers();
        if(self.socket !== undefined) {
          self.socket.$close();
        }

        self.project = undefined;
        self.mmlData = undefined;
        self.dashboardMaps = [];
        self.bookmarkedMaps = [];
        self.appSettings = {};
        self.socketUrl = undefined;
        self.socket = undefined;
        self.mmlLoadPromise = undefined;

        DashboardService.maps = [];
        StyleService.setStyles([]);
        StyleService.setProjectStyles([]);
        LayerService.setLayers([]);
      };

      ProjectServiceInstance.prototype.saveMML = function() {
        var self = this;
        $http.post(magnaConfig.projectBaseUrl + self.project.base + '/' + self.project.mml, angular.toJson(self.mmlData, true));
      };

      ProjectServiceInstance.prototype.saveMCP = function() {
        var self = this;
        if(self.mcpSaveTimeout !== undefined) {
          $timeout.cancel(self.mcpSaveTimeout);
        }
        // prevent too often safe. mostly triggered by gridster when resize or dragging a map
        self.mcpSaveTimeout = $timeout(function() {
          var sendMcpData = angular.copy(self.mcpData);
          if(sendMcpData.appSettings.sidebarWidth === -1) {
            delete sendMcpData.appSettings.sidebarWidth;
          }
          if(sendMcpData.appSettings.loggingHeight === -1) {
            delete sendMcpData.appSettings.loggingHeight;
          }
          $http.post(magnaConfig.projectBaseUrl + self.project.base + '/' + self.project.mcp, angular.toJson(sendMcpData, true));
          self.mcpSaveTimeout = undefined;
        }, 1000);
      };

      ProjectServiceInstance.prototype.bindSocket = function() {
        var self = this;
        self.socketUrl = angular.copy(magnaConfig.socketUrl);
        self.socketUrl += 'mml=' + self.project.mml;
        self.socketUrl += '&mss=' + self.project.available_mss;
        self.socketUrl += '&base=' + self.project.base;
        self.socket = $websocket.$new({
          url: self.socketUrl,
          reconnect: true,
          reconnectInterval: 100
        });

        self.projectLoadedPromise = self.projectLoadedPromise.then(function() {
          self.socket.$on('$message', function (resp) {
            if(resp.updated_mml === true && LayerService.editModalOpen === false) {
              self.mmlLoadPromise = self.loadMML();
              self.projectLoadedPromise = $q.all([self.mmlLoadPromise, self.mcpLoadPromise]).then(function(data) {
                self.handleMMLResponse(data[0].data);
              });
            }
          });
        });
      };

      ProjectServiceInstance.prototype.projectLoaded = function() {
        var self = this;
        return self.projectLoadedPromise;
      };

      ProjectServiceInstance.prototype.getSocket = function() {
        return this.socket;
      };

      ProjectServiceInstance.prototype.enableWatchers = function() {
        var self = this;

        // listen on changes in dashboardMaps
        // save if change occurs
        self.dashboardMapsWatcher = $rootScope.$watch(function() {
          return self.mcpData.dashboardMaps;
        }, function(n, o) {
          if(n === o) return;
          self.saveMCP();
        }, true);

        // listen on changes in bookmarkedMaps
        // save if change occurs
        self.bookmarkedMapsWatcher = $rootScope.$watch(function() {
          return self.bookmarkedMaps;
        }, function(n, o) {
          if(n === o) return;
          self.saveMCP();
        }, true);

        self.appSettingsWatcher = $rootScope.$watch(function() {
          return self.appSettings;
        }, function(n, o) {
          if(angular.equals(n, o)) return;
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

      ProjectServiceInstance.prototype.disableWatchers = function() {
        var self = this;
        if(self.dashboardMapsWatcher !== undefined) {
          self.dashboardMapsWatcher();
          self.dashboardMapsWatcher = undefined;
        }

        if(self.bookmarkedMapsWatcher !== undefined) {
          self.bookmarkedMapsWatcher();
          self.bookmarkedMapsWatcher = undefined;
        }

        if(self.appSettingsWatcher !== undefined) {
          self.appSettingsWatcher();
          self.appSettingsWatcher = undefined;
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

      return new ProjectServiceInstance();
  }];
}]);

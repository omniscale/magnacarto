angular.module('magna-app')

.provider('MMLService', [function() {

  var loadUrl;
  var saveUrl;

  var fakePost = function(url, data) {
    console.log(url, data);
  };

  this.setLoadUrl = function(url) {
    loadUrl = url;
  };

  this.setSaveUrl = function(url) {
    saveUrl = url;
  };

  this.$get = ['$http', '$rootScope',
    function($http, $rootScope) {
      var MMLServiceInstance = function(loadUrl, saveUrl) {
        this.mml = undefined;
        this.styles = undefined;
        this.loadUrl = loadUrl;
        this.saveUrl = saveUrl;
      };

      MMLServiceInstance.prototype.load = function(mml) {
        var self = this;
        self.mml = mml;
        self.loadPromise = $http.get(self.loadUrl + self.mml);
        // TODO add on error
        self.loadPromise.success(function(data) {
          self.styles = data.Stylesheet;
          self.dashboardMaps = data.magnacarto.dashboardMaps;
          self.storedMaps = data.magnacarto.storedMaps;

          // listen on changes in dashboardMaps
          // save if change occurs
          $rootScope.$watch(function() {
            return self.dashboardMaps;
          }, function() {
            self.saveDashboardMaps();
          }, true);

          // listen on changes in storedMaps
          // save if change occurs
          $rootScope.$watch(function() {
            return self.storedMaps;
          }, function() {
            self.saveStoredMaps();
          }, true);

        });
        return self.loadPromise;
      };

      MMLServiceInstance.prototype.loaded = function() {
        var self = this;
        return self.loadPromise;
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

      return new MMLServiceInstance(loadUrl, saveUrl);
  }];
}]);

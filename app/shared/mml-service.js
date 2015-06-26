angular.module('magna-app')

.provider('MMLService', [function() {

  this.$get = ['$http',
    function($http) {
      var MMLServiceInstance = function() {
        this.mml = undefined;
        this.styles = undefined;
      };

      MMLServiceInstance.prototype.load = function(mml) {
        var self = this;
        self.loadPromise = $http.get(mml);
        // TODO add on error
        self.loadPromise.success(function(data) {
          self.styles = data.Stylesheet;
        });
        return self.loadPromise;
      };

      MMLServiceInstance.prototype.loaded = function() {
        var self = this;
        return self.loadPromise;
      };

      return new MMLServiceInstance();
  }];
}]);

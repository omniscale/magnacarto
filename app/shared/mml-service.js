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
        var loadPromise = $http.get(mml);
        // TODO add on error
        loadPromise.success(function(data) {
          self.styles = data.Stylesheet;
        });
        return loadPromise;
      };

      return new MMLServiceInstance();
  }];
}]);

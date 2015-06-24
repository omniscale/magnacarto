angular.module('magna-app')

.provider('MMLService', [function() {

  this.$get = ['$http',
    function($http) {
      var MMLServiceInstance = function() {
        this.mml = undefined;
        this.mss = undefined;
      };

      MMLServiceInstance.prototype.load = function(mml) {
        var self = this;
        var loadPromise = $http.get(mml);
        // TODO add on error
        loadPromise.success(function(data) {
          self.mss = data.Stylesheet;
        });
        return loadPromise;
      };

      return new MMLServiceInstance();
  }];
}]);

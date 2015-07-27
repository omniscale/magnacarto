angular.module('magna-app')

.provider('ProjectsService', ['magnaConfig', function(magnaConfig) {
  this.$get = ['$http',
    function($http) {
      var ProjectsServiceInstance = function() {
        this.projects = [];
      };

      ProjectsServiceInstance.prototype.load = function() {
        var self = this;

        self.loadPromise = $http.get(magnaConfig.projectsUrl);
        // TODO add on error
        self.loadPromise.success(function(data) {
          self.projects = data.projects;
        });
        return self.loadPromise;
      };

      ProjectsServiceInstance.prototype.loaded = function() {
        var self = this;
        return self.loadPromise;
      };

      return new ProjectsServiceInstance();
  }];
}]);

angular.module('magna-app')

.provider('ProjectsService', ['magnaConfig', function(magnaConfig) {
  this.$get = ['$http',
    function($http) {
      var ProjectsServiceInstance = function() {
        this.projects = {};
      };

      ProjectsServiceInstance.prototype.load = function() {
        var self = this;

        self.loadPromise = $http.get(magnaConfig.projectsUrl);
        // TODO add on error
        self.loadPromise.success(function(data) {
          angular.forEach(data.projects, function(project) {
            self.projects[project.base + '|' + project.mml] = project;
          });
        });
        return self.loadPromise;
      };

      ProjectsServiceInstance.prototype.loaded = function() {
        var self = this;
        return self.loadPromise;
      };

      ProjectsServiceInstance.prototype.projectByRouteParams = function(routeParams) {
        var self = this;
        return self.projects[routeParams.base + '|' + routeParams.mml];
      };

      return new ProjectsServiceInstance();
  }];
}]);

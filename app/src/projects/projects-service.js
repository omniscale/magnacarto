angular.module('magna-app')

.provider('ProjectsService', ['magnaConfig', function(magnaConfig) {
  this.$get = ['$http',
    function($http) {
      var ProjectsServiceInstance = function() {
        // items in projects and projectsList refer same objects
        this.projects = {};
        this.projectsList = [];
      };

      ProjectsServiceInstance.prototype.load = function() {
        var self = this;

        self.loadPromise = $http.get(magnaConfig.projectsUrl);
        // TODO add on error
        self.loadPromise.success(function(data) {
          angular.forEach(data.projects, function(project) {
            project.last_change = Date.parse(project.last_change);
            self.projects[project.base + '|' + project.mml] = project;
            self.projectsList.push(project);
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

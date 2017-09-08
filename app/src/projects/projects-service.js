angular.module('magna-app')

.provider('ProjectsService', ['magnaConfig', function(magnaConfig) {
  this.$get = ['$http', '$q',
    function($http, $q) {
      var ProjectsServiceInstance = function() {
        // items in projects and projectsList refer same objects
        this.projects = {};
        this.projectsList = [];
        this.loadDeferred = $q.defer();
        this.loadPromise = this.loadDeferred.promise;
      };

      ProjectsServiceInstance.prototype.load = function() {
        var self = this;

        $http.get(magnaConfig.projectsUrl).success(function(data) {
          angular.forEach(data.projects, function(project) {
            project.last_change = Date.parse(project.last_change);
            project.url = project.base === '.' ? '' : project.base + '/';
            project.url += project.mml;
            self.projects[project.url] = project;
            self.projectsList.push(project);
          });
          self.loadDeferred.resolve(data);
        });
        return self.loadPromise;
      };

      ProjectsServiceInstance.prototype.loaded = function() {
        var self = this;
        return self.loadPromise;
      };

      ProjectsServiceInstance.prototype.projectByUrl = function(url) {
        var self = this;
        return self.projects[url];
      };

      return new ProjectsServiceInstance();
  }];
}]);

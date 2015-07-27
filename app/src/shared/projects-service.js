angular.module('magna-app')

.provider('ProjectsService', [function() {

  var projectsUrl;

  this.setProjectsUrl = function(url) {
    projectsUrl = url;
  };

  this.$get = ['$http',
    function($http) {
      var ProjectsServiceInstance = function(url) {
        this.url = url;
        this.projects = [];
      };

      ProjectsServiceInstance.prototype.load = function() {
        var self = this;

        self.loadPromise = $http.get(self.url);
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

      return new ProjectsServiceInstance(projectsUrl);
  }];
}]);

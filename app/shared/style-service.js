angular.module('magna-app')

.provider('StyleService', [function() {
  this.$get = [
    function() {
      var StyleServiceInstance = function() {
        var self = this;
        self.styles = [];
        self.activeStyles = [];
      };

      StyleServiceInstance.prototype.setStyles = function(styles) {
        var self = this;
        self.styles = styles;
        // need to copy otherwise styles and activeStyles are the same
        // list object
        self.activeStyles = angular.copy(styles);
      };

      StyleServiceInstance.prototype.toggleStyle = function(style) {
        var self = this;
        var idx = self.activeStyles.indexOf(style);
        if (idx > -1) {
          self.activeStyles.splice(idx, 1);
        } else {
          self.activeStyles.push(style);
        }
      };

      return new StyleServiceInstance();
    }];
}]);

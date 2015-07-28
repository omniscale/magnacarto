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
        this.styles = styles;
      };

      StyleServiceInstance.prototype.setProjectStyles = function(styles) {
        this.activeStyles = styles;
      };

      StyleServiceInstance.prototype.toggleStyle = function(style) {
        var self = this;
        var idx = self.activeStyles.indexOf(style);
        if (idx > -1) {
          self.activeStyles.splice(idx, 1);
        } else {
          // TODO place style at right place
          self.activeStyles.push(style);
        }
      };

      return new StyleServiceInstance();
    }];
}]);

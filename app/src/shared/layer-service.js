angular.module('magna-app')

.provider('LayerService', [function() {
  this.$get = [
    function() {
      var LayerServiceInstance = function() {
        var self = this;
        self.layers = [];
      };

      LayerServiceInstance.prototype.setLayers = function(layers) {
        var self = this;
        self.layers = layers;
      };

      LayerServiceInstance.prototype.addLayer = function(layer) {
        var self = this;
        self.layers.push(layer);
      };

      return new LayerServiceInstance();
    }];
}]);

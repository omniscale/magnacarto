angular.module('magna-app')

.provider('LayerService', [function() {
  this.$get = ['$modal',
    function($modal) {
      var DEFAULT_LAYER = {
        'extent': [0, 0, 0, 0],
        'Datasource': {
          'type': 'postgis'
        },
        'advanced': {}
      };

      var LayerServiceInstance = function() {
        var self = this;
        self.layers = [];
      };

      LayerServiceInstance.prototype.setLayers = function(layers) {
        var self = this;
        self.layers = layers;
      };

      LayerServiceInstance.prototype.addLayer = function() {
        var self = this;
        var layer = angular.copy(DEFAULT_LAYER);
        self.editLayer(layer);
      };

      LayerServiceInstance.prototype.editLayer = function(layer) {
        var self = this;
        var modalInstance = $modal.open({
          templateUrl: 'src/layer/edit-layer-template.html',
          controller: 'EditLayerCtrl',
          backdrop: 'static',
          windowClass: 'layer-modal',
          resolve: {
            layer: function () {
              return layer;
            }
          }
        });

        modalInstance.result.then(function (item) {
          var layerIdx = self.layers.indexOf(layer);

          if(item === 'remove') {
            if(layerIdx !== -1) {
              self.layers.splice(layerIdx, 1);
            }
          } else if(layerIdx === -1) {
            self.layers.push(item);
          } else {
            self.layers[layerIdx] = item;
          }
        });
      };

      LayerServiceInstance.prototype.isDefaultLayer = function(layer) {
        return angular.equals(DEFAULT_LAYER, layer);
      };

      return new LayerServiceInstance();
    }];
}]);

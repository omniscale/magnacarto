angular.module('magna-app')

.provider('LayerService', [function() {
  this.$get = ['$modal',
    function($modal) {
      var DEFAULT_LAYER = {
        'extent': [0, 0, 0, 0],
        'Datasource': {},
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

      LayerServiceInstance.prototype.addLayer = function(layer) {
        var self = this;

        layer = layer === undefined ? angular.copy(DEFAULT_LAYER) : layer;

        var modalInstance = $modal.open({
          templateUrl: 'src/layer/edit-layer-template.html',
          controller: 'EditLayerCtrl',
          resolve: {
            layer: function () {
              return angular.copy(layer);
            }
          }
        });

        modalInstance.result.then(function (newLayer) {
          if(newLayer !== 'remove') {
            self.layers.push(newLayer);
          }
        });
      };

      LayerServiceInstance.prototype.copyLayer = function(layer) {
        var self = this;
        layer.name += '-copy';
        self.addLayer(layer);
      };

      LayerServiceInstance.prototype.editLayer = function(layer) {
        var self = this;
        var modalInstance = $modal.open({
          templateUrl: 'src/layer/edit-layer-template.html',
          controller: 'EditLayerCtrl',
          resolve: {
            layer: function () {
              return layer;
            }
          }
        });

        modalInstance.result.then(function (item) {
          var layerIdx = self.layers.indexOf(layer);

           if(item === 'remove') {
            self.layers.splice(layerIdx, 1);
          } else {
            self.layers[layerIdx] = item;
          }
        });
      };
      return new LayerServiceInstance();
    }];
}]);

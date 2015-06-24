angular.module('magna-app')

.directive('ol3Map',
  function() {
    return {
      restrict: 'A',
      scope: {
          controls: '=controls',
          layers: '=layers',
          settings: '=settings'
      },
      link: function(scope, element) {
        // intialize map with coords from settings
        var coords = scope.settings.coords;
        var zoom = scope.settings.zoom;
        var zoomControl = true;
        var controls = ol.control.defaults();
        var interactions = ol.interaction.defaults();
        var olLayers = [];
        var map;

        var updateSource = function(layer, olSource) {
          var params = olSource.getParams();
          params.mss = layer.mss.join(',');
          params.t = Date.now();
          olSource.updateParams(params);
        };

        if (scope.controls === 'false') {
          zoomControl = false;
          interactions = [];
          controls = [];
        }

        angular.forEach(scope.layers, function(layer) {
          var olSource = new ol.source.ImageWMS({
            url: layer.url,
            ratio: 1,
            params: {
              LAYERS: 'osm',
              TRANSPARENT: false,
              VERSION: '1.1.1',
              mml: layer.mml,
              mss: layer.mss.join(','),
              t: Date.now()
            }
          });
          olLayers.push(new ol.layer.Image({
            source: olSource
          }));

          // update source params when mss list changes
          scope.$watch(function() {
            return layer.mss;
          }, function() {
              updateSource(layer, olSource);
          }, true);

          scope.$on('socketUpdateImage', function () {
            updateSource(layer, olSource);
          });
        });
        map = new ol.Map({
          target: element[0],
          layers: olLayers,
          interactions: interactions,
          controls: controls,
          logo: false,
          view: new ol.View({
            center: ol.proj.transform(coords, 'EPSG:4326', 'EPSG:3857'),
            zoom: zoom
          })
        });
        map.on('moveend', function() {
          var center = map.getView().getCenter();
          center = ol.proj.transform(center, 'EPSG:3857', 'EPSG:4326');
          scope.$apply(function() {
            scope.settings.coords = center;
            scope.settings.zoom = map.getView().getZoom();
          });
        });
        // remove openlayers map
        scope.$on('$destroy', function () {
          map.setTarget(null);
          map = null;
        });

        // TODO: Find a solition to update map after loading dashboard
        scope.$on('gridUpdate', function () {
          map.updateSize();
        });
      }
    };
});

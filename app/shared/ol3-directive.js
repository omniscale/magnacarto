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
        var coords = (scope.settings.coords) ? scope.settings.coords: [8,52];
        var zoom = (scope.settings.zoom) ? scope.settings.zoom: 11;

        var zoomControl = true;
        var controls = ol.control.defaults();
        var interactions = ol.interaction.defaults();
        if (scope.controls === 'false') {
          zoomControl = false;
          interactions = [];
          controls = [];
        }

        // copy and add layers for each map
        var layersSettings = [];
        var layers = [];
        angular.copy(scope.layers, layersSettings);


        angular.forEach(layersSettings, function(layer) {
          var olLayer = new ol.layer.Image({
            source: new ol.source.ImageWMS({
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
            })
          });
          layers.push(olLayer);
        });

        var map = new ol.Map({
          target: element[0],
          layers: layers,
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
          scope.settings.coords = center;
          scope.settings.zoom = map.getView().getZoom();
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

        scope.$on('imageUpdateMss', function (scope, mss) {
          var mssList = mss.join(',');
          angular.forEach(layers, function(layer) {
            var params = layer.getSource().getParams();
            params.mss = mssList;
            layer.getSource().updateParams(params);
          });
        });


        scope.$on('socketUpdateImage', function () {
          // redraw all layers after socket update
          angular.forEach(layers, function(layer) {
            var params = layer.getSource().getParams();
            params.t = Date.now();
            layer.getSource().updateParams(params);
          });
        });
      }
    };
});

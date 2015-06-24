angular.module('magna-app')

.directive('ol3Map', ['magnaConfig',
  function(magnaConfig) {
    return {
      restrict: 'A',
      scope: {
          controls: '=controls',
          // TODO verify need of layers
          styles: '=styles',
          settings: '=settings'
      },
      link: function(scope, element) {
        // intialize with default values
        var coords = scope.settings.coords;
        var zoom = scope.settings.zoom;
        var zoomControl = true;
        var controls = ol.control.defaults();
        var interactions = ol.interaction.defaults();
        var olMap;
        var olSource;

        var updateSource = function() {
          params.mss = scope.styles.join(',');
          params.t = Date.now();
          olSource.updateParams(params);
        };

        var params = {
          LAYERS: magnaConfig.mapnikLayers,
          TRANSPARENT: false,
          VERSION: '1.1.1',
          mml: magnaConfig.mml,
          mss: scope.styles.join(','),
          t: Date.now()
        };

        if (scope.controls === 'false') {
          zoomControl = false;
          interactions = [];
          controls = [];
        }

        // init ol3 source
        olSource = new ol.source.ImageWMS({
          url: magnaConfig.mapnikUrl,
          ratio: 1,
          params: params
        });

        // update source params when style list changes
        scope.$watch('styles', function() {
            updateSource();
        }, true);

        scope.$on('socketUpdateImage', function () {
          updateSource();
        });

        // init map
        olMap = new ol.Map({
          target: element[0],
          layers: [new ol.layer.Image({
            source: olSource
          })],
          interactions: interactions,
          controls: controls,
          logo: false,
          view: new ol.View({
            center: ol.proj.transform(coords, 'EPSG:4326', 'EPSG:3857'),
            zoom: zoom
          })
        });

        // update zoom and coords after map move ends
        olMap.on('moveend', function() {
          var center = olMap.getView().getCenter();
          center = ol.proj.transform(center, 'EPSG:3857', 'EPSG:4326');
          scope.$apply(function() {
            scope.settings.coords = center;
            scope.settings.zoom = olMap.getView().getZoom();
          });
        });

        // remove openlayers map
        scope.$on('$destroy', function () {
          olMap.setTarget(null);
          olMap = null;
        });

        // TODO: Find a solition to update map after loading dashboard
        scope.$on('gridUpdate', function () {
          olMap.updateSize();
        });
      }
    };
}]);

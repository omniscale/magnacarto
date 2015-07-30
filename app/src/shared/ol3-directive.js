angular.module('magna-app')

.directive('ol3Map', ['$timeout', '$websocket', 'magnaConfig', 'MMLService',
  function($timeout, $websocket, magnaConfig, MMLService) {
    return {
      restrict: 'A',
      scope: {
          staticMap: '@',
          styles: '=',
          settings: '=ol3Map',
          gridsterItemElement: '='
      },
      link: {
        pre: function(scope) {
          scope.updateSource = function() {
            scope.params.mss = scope.styles.join(',');
            scope.params.t = Date.now();
            scope.olSource.updateParams(scope.params);
          };

          scope.socket = MMLService.getSocket();
          scope.staticMap = scope.staticMap === 'true' ? true : false;

          // intialize with default values
          scope.olControls = scope.staticMap ? [] : ol.control.defaults();
          scope.olInteractions = scope.staticMap ? [] : ol.interaction.defaults();
          scope.params = {
            LAYERS: magnaConfig.mapnikLayers,
            TRANSPARENT: false,
            VERSION: '1.1.1',
            mml: MMLService.mml,
            mss: scope.styles.join(','),
            base: MMLService.basePath,
            t: Date.now()
          };

          // init ol3 source
          scope.olSource = new ol.source.ImageWMS({
            url: magnaConfig.mapnikUrl,
            ratio: 1,
            params: scope.params
          });

          // init map
          scope.olMap = new ol.Map({
            layers: [],
            interactions: scope.olInteractions,
            controls: scope.olControls,
            logo: false,
            view: new ol.View({
              center: ol.proj.transform(scope.settings.coords, 'EPSG:4326', 'EPSG:3857'),
              zoom: scope.settings.zoom
            })
          });
        },
        post: function(scope, element) {
          scope.olMap.addLayer(new ol.layer.Image({
            source: scope.olSource
          }));
          scope.lastUpdate = new Date();

          if(!scope.staticMap) {
            var displayZoomLevel = angular.element('<span>' + scope.settings.zoom + '</span>');
            var zoomLevelContainer = angular.element('<div></div>');
            zoomLevelContainer.addClass('ol-control');
            zoomLevelContainer.addClass('display-zoom-level');
            zoomLevelContainer.append('Zoom level:');
            zoomLevelContainer.append(displayZoomLevel);

            var showZoomLevelControl = new ol.control.Control({element: zoomLevelContainer[0]});
            var view = scope.olMap.getView();
            view.on('change:resolution', function(e) {
              displayZoomLevel.text(view.getZoom());
            });
            scope.olMap.addControl(showZoomLevelControl);
          }

          scope.socket.$on('$message', function (resp) {
            // without updated_at do nothing
            if(resp.updated_at === undefined) {
              return;
            }
            var updatedAt = new Date(resp.updated_at);
            // do nothing if we up to date
            if(scope.lastUpdate !== undefined && scope.lastUpdate >= updatedAt) {
              return;
            }
            scope.updateSource();
            scope.lastUpdate = updatedAt;
          });

          // update zoom and coords after map move ends
          scope.olMap.on('moveend', function() {
            var center = scope.olMap.getView().getCenter();
            center = ol.proj.transform(center, 'EPSG:3857', 'EPSG:4326');
            scope.$apply(function() {
              scope.settings.coords = center;
              scope.settings.zoom = scope.olMap.getView().getZoom();
            });
          });

          // remove openlayers map
          scope.$on('$destroy', function () {
            scope.olMap.setTarget(null);
            scope.olMap = null;
            scope.updatedAt = undefined;
          });

          scope.$watch('gridsterItemElement.gridster.loaded', function(loaded) {
            // gridstar loaded complete
            if(loaded === true) {
              // add map to dom
              scope.olMap.setTarget(element[0]);
              // add function to gridsterItem scope
              // so it's callable in gridster resizeable stop callback
              // see controller which defines gridster options
              scope.gridsterItemElement.$element.scope().resizeMap = function() {
                $timeout(function() {
                  scope.olMap.updateSize();
                });
              };
              // update map size when gridster react on browser window size change
              scope.$on('gridster-resized', function() {
                $timeout(function() {
                  scope.olMap.updateSize();
                });
              });
            }
          });

          scope.$watch('styles', function() {
            scope.updateSource();
          }, true);
        }
      }
    };
}]);

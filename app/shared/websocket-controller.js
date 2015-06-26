angular.module('magna-app')

.controller('WebSocketCtrl', ['$scope', '$websocket', 'magnaConfig', 'MMLService',
  function($scope, $websocket, magnaConfig, MMLService) {
    $scope.alerts = [];

    // cause socket is set up after MMLService finished loading,
    // we have to bind our listeners after its loading promise
    // was successfull resolved
    MMLService.loaded().success(function() {
      var socket = $websocket.$get(magnaConfig.socketUrl);

      socket.$on('$open', function() {
        $scope.alerts.push({
           type: 'info',
            msg: 'Connect to the websocket Server'
          }
        );
      });

      socket.$on('$message', function (resp) {
        var type = 'success';
        if(resp.error !== undefined) {
          type = 'error';
        }
        $scope.alerts.push({
          type: 'success',
          msg: resp
        });
      });
    });
}]);

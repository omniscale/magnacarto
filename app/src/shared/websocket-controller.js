angular.module('magna-app')

.controller('WebSocketCtrl', ['$scope', '$websocket', 'magnaConfig', 'MMLService',
  function($scope, $websocket, magnaConfig, MMLService) {
    $scope.alerts = [];

    // Add messages handler when socket changes
    $scope.$watch(function() {
      return MMLService.getSocket();
    }, function(n, o) {
      if(n === o) return;
      var socket = n;
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

angular.module('magna-app')

.controller('WebSocketCtrl', ['$scope', '$websocket', 'magnaConfig', 'ProjectService',
  function($scope, $websocket, magnaConfig, ProjectService) {
    $scope.alerts = [];

    var appendMessage = function(type, msg) {
      $scope.$apply(function() {
        $scope.alerts.push({
          type: type,
          msg: msg
        });
      });
    };

    // Add messages handler when socket changes
    $scope.$watch(function() {
      return ProjectService.getSocket();
    }, function(n, o) {
      if(n === o || n === undefined) return;
      var socket = n;
      socket.$on('$open', function() {
        appendMessage('info', 'Connect to the websocket Server');
      });

      socket.$on('$message', function (resp) {
        var type = 'success';
        if(resp.error !== undefined) {
          type = 'error';
        }
        appendMessage(type, resp);
      });
    });
}]);

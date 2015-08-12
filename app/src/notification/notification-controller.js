angular.module('magna-app')

.controller('NotificationCtrl', ['$scope', '$websocket', 'magnaConfig', 'ProjectService',
  function($scope, $websocket, magnaConfig, ProjectService) {
    $scope.notifications = [];

    var appendMessage = function(type, msg) {
      $scope.$apply(function() {
        if(type === 'success') {
          angular.forEach($scope.notifications, function(notification) {
            notification.close(1000);
          });
        }
        $scope.notifications.push({
          type: type,
          msg: msg
        });
      });
    };

    var handleMessage = function(resp) {
      var type, msg;
      if(resp === undefined) {
        return;
      } else  if(resp.error !== undefined) {
        type = 'error';
        msg = [];
        if (resp.filename !== undefined) {
          msg.append('Error in ' + resp.filename + ':');
          msg.push(resp.error);
        } else {
          msg.push('Error: ' + resp.error);
        }
        if (resp.files !== undefined) {
          angular.forEach(resp.files, function(file) {
            msg.push('• ' + file);
          });
        }
      } else if(resp.updated_at !== undefined) {
        type = 'success';
        msg = ['Updated'];
      } else {
        return;
      }
      appendMessage(type, msg);
    };

    // Add messages handler when socket changes
    $scope.$watch(function() {
      return ProjectService.getSocket();
    }, function(n, o) {
      if(n === o || n === undefined) return;
      var socket = n;
      socket.$on('$open', function() {
        appendMessage('info', ['Connect to the websocket Server']);
      });

      socket.$on('$message', handleMessage);
    });
}]);

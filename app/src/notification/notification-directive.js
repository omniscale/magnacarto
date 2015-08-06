angular.module('magna-app')

.directive('notification', ['$timeout',
  function($timeout) {
    return {
    restrict: 'E',
    replace: true,
    controller: function NotificationCtrl ($scope) {
      $scope.close = function(index) {
        $scope.notifications.splice(index, 1);
      };
    },
    templateUrl: 'src/notification/notification-template.html',
    link: function(scope, element, attrs) {
      var notification = scope.notifications[attrs.index];

      notification.close = function(timeout) {
        timeout = timeout === undefined ? 0 : timeout;
        $timeout(function() {
          scope.close(attrs.index);
        }, timeout);
      };

      if(notification.type !== 'error') {
        $timeout(function() {
          scope.close(attrs.index);
        }, 2000);
      }
    }
  };
}]);

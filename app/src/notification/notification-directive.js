angular.module('magna-app')

.directive('notification', ['$timeout',
  function($timeout) {
    return {
    restrict: 'E',
    replace: true,
    scope: {
      notification: '=item',
      notifications: '=ngModel'
    },
    controller: function NotificationCtrl ($scope) {
      var close = function() {
        var idx = $scope.notifications.indexOf($scope.notification);
        if(idx > -1) {
          $scope.notifications.splice(idx, 1);
        }
      };

      $scope.closeTimeout = undefined;

      $scope.close = function(closeTime) {
        closeTime = closeTime === undefined ? 0 : closeTime;

        if($scope.notification.timeout !== undefined) {
          $timeout.cancel($scope.notification.timeout);
        }

        $scope.closeTimeout = $timeout(close, closeTime);
      };
    },
    templateUrl: 'src/notification/notification-template.html',
    link: function(scope) {
      scope.notification.close = scope.close;

      if(scope.notification.type !== 'error') {
        scope.notification.close(2000);
      }
    }
  };
}]);

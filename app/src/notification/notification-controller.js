angular.module('magna-app')

.controller('NotificationCtrl', ['$scope', 'LoggingService',
  function($scope, LoggingService) {
    $scope.notifications = [];

    var appendMessage = function(msg) {
      $scope.$apply(function() {
        angular.forEach($scope.notifications, function(notification) {
          notification.close(500);
        });
        $scope.notifications.push(msg);
      });
    };

    LoggingService.addNewMessageCallback(appendMessage);
}]);

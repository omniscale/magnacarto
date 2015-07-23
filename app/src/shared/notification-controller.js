angular.module('magna-app')

.directive('notification', ['$timeout',
  function($timeout) {
    return {
    restrict: 'E',
    replace: true,
    controller: function NotificationCtrl ($scope) {
      $scope.close = function(index) {
        $scope.alerts.splice(index, 1);
      };
    },
    templateUrl: '/src/shared/notification-template.html',
    link: function(scope, element, attrs) {
      $timeout(function(){
        scope.alerts.splice(attrs.index, 1);
      }, 2000);
    }
  };
}]);

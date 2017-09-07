angular.module('magna-app')

.controller('LoggingCtrl', ['$scope', 'LoggingService',
  function($scope, LoggingService) {
    $scope.messages = LoggingService.messages;
}]);

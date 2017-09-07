angular.module('magna-app')

.controller('LoggingCtrl', ['$scope', 'LoggingService',
  function($scope, LoggingService) {
    $scope.messages = LoggingService.messages;

    $scope.messageClass = function(idx, message) {
      if(idx <= LoggingService.lastSuccessfulUpdateIdx) {
        return message.type;
      }
      return 'outdated';
    };

    $scope.iconClass = function(idx, message) {
      var classes = [];
      switch(message.type) {
        case 'danger':
          classes.push('glyphicon-remove-sign');
          break;
        case 'info':
          classes.push('glyphicon-info-sign');
          break;
        case 'success':
          classes.push('glyphicon-ok-sign');
          break;
      }
      if(idx > LoggingService.lastSuccessfulUpdateIdx) {
        return classes;
      }
      classes.push('text-' + message.type);
      return classes;
    };
}]);

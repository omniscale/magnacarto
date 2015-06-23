angular.module('magna-app')

.controller('WebSocketCtrl', ['$scope',
  function($scope) {
    $scope.alerts = [];

    // TODO: hide after 1-2 secons
    $scope.$on('socketOpen', function () {
      $scope.alerts.push({
         type: 'info',
          msg: 'Connect to the websocket Server'
        }
      );
      // TODO: check if there is a nicer way then $scope.$apply();
      $scope.$apply();
    });

    $scope.$on('socketUpdateImage', function (evt, resp) {
      $scope.alerts.push({
          type: 'success',
          msg: resp
        }
      );
      // TODO: check if there is a nicer way then $scope.$apply();
      $scope.$apply();
    });

    $scope.$on('socketError', function (evt, resp) {
      $scope.alerts.push({
          type: 'error',
          msg: resp
        }
      );
      // TODO: check if there is a nicer way then $scope.$apply();
      $scope.$apply();
    });
}]);




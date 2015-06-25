angular.module('magna-app')

.controller('WebSocketCtrl', ['$scope',
  function($scope) {
    $scope.alerts = [];

    $scope.$on('socketOpen', function () {
      $scope.alerts.push({
         type: 'info',
          msg: 'Connect to the websocket Server'
        }
      );
    });

    $scope.$on('socketUpdateImage', function (evt, resp) {
      $scope.alerts.push({
          type: 'success',
          msg: resp
        }
      );
    });

    $scope.$on('socketError', function (evt, resp) {
      $scope.alerts.push({
          type: 'error',
          msg: resp
        }
      );
    });
}]);

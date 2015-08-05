angular.module('magna-app')

.controller('AboutCtrl', ['$scope', '$modalInstance', 'magnaConfig',
  function($scope, $modalInstance, magnaConfig) {
    $scope.version = magnaConfig.version;
    $scope.cancel = function () {
      $modalInstance.dismiss('cancel');
    };
  }
]);
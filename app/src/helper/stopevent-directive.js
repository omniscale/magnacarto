angular.module('magna-app')

.directive('stopEvent', function () {
  return {
    restrict: 'A',
    link: function (scope, element) {
      element.bind('click', function (e) {
        e.stopPropagation();
      });

      element.bind('mousedown', function (e) {
        e.stopPropagation();
      });
    }
  };
});

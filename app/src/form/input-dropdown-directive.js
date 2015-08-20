angular.module('magna-app')

.directive('inputDropdown', [function() {
  return {
    replace: true,
    transclude: true,
    require: '^form',
    scope: {
        suggestions: '=inputDropdown',
        target: '=ngModel'
    },
    templateUrl: 'src/form/input-dropdown.html',
    link: function(scope) {
      scope.select = function(value) {
        scope.target = value;
      };
    }
  };
}]);

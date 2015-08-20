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
    link: function(scope, element, attrs, formController) {
      scope.form = formController;

      scope.toggle = function() {
        scope.form.inputDropdownOpen = scope.form.inputDropdownOpen === scope.$id ? false : scope.$id;
      };

      scope.select = function(value) {
        scope.target = value;
        scope.form.inputDropdownOpen=false;
      };
    }
  };
}]);

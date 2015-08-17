angular.module('magna-app')

.directive('formGroup', ['$filter', function($filter) {
  return {
    replace: true,
    transclude: true,
    require: '^form',
    scope: {},
    templateUrl: 'src/form/form-group.html',
    link: function(scope, element, attrs, formController) {
      scope.form = formController;
      scope.formClass = {
        'has-error': undefined,
        'has-warning': undefined
      };
      scope.error = {};

      var items = element.find('input');
      if (items.length === 0) {
        items = element.find('select');
      }
      if (items.length === 0) {
        items = angular.element(element[0].getElementsByClassName('ace_editor'));
      }
      if(items.length === 0) {
        throw 'NoInputElementError';
      }

      var item = angular.element(items[0]);
      item.addClass('form-control');
      scope.name = item.attr('name');
      scope.title = item.attr('title') || $filter('titleCase')(scope.name);
      scope.id = item.attr('id') || scope.name;

      scope.itemScope = scope.form[scope.name];
      scope.$watchGroup(['itemScope.$dirty', 'itemScope.$invalid', 'itemScope.$error'],
        function (n, o, scope) {
          scope.error.required = n[0] && n[2].required;
          scope.formClass['has-error'] = n[0] && n[1];
          scope.formClass['has-warning'] = !n[0] && n[1];
      });
    }
  };
}]);

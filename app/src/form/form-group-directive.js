angular.module('magna-app')

.directive('formGroup', ['$filter', function($filter) {
  return {
    replace: true,
    transclude: true,
    require: '^form',
    scope: {},
    templateUrl: 'src/form/form-group.html',
    link: function(scope, element, attrs, formController) {
      scope.formClass = {
        'has-error': undefined,
        'has-warning': undefined
      };
      var isAce = false;

      var items = element.find('input');
      if (items.length === 0) {
        items = element.find('select');
      }
      if (items.length === 0) {
        items = angular.element(element[0].getElementsByClassName('ace_editor'));
        if(items.length > 0) {
          isAce = true;
        }
      }
      if(items.length === 0) {
        throw 'NoInputElementError';
      }

      var item = angular.element(items[0]);
      item.addClass('form-control');

      scope.name = item.attr('name');

      if(!item.attr('title')) {
        item.attr('title', $filter('titleCase')(scope.name));
      }
      scope.title = item.attr('title');

      if(!item.attr('id')) {
        item.attr('id', scope.name);
      }
      scope.id = item.attr('id');

      if(isAce) {
        item.removeAttr('id');
        item.find('textarea').attr('id', scope.id);
      }

      scope.itemScope = formController[scope.name];
      scope.$watchGroup(['itemScope.$dirty', 'itemScope.$invalid'],
        function (n) {
          scope.formClass['has-error'] = n[0] && n[1];
          scope.formClass['has-warning'] = !n[0] && n[1];
      });
    }
  };
}]);

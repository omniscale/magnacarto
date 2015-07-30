angular.module('magna-app')

.directive('float', function() {
  return {
    require: 'ngModel',
    link: function(scope, elm, attrs, ctrl) {
      ctrl.$parsers.unshift(function(viewValue) {
        if (!viewValue || viewValue.length === 0) {
          ctrl.$setValidity('float', true);
          return viewValue;
        } else if (/^\-?\d*((\.|\,)\d+)?$/.test(viewValue)) {
          //allow . and ,
          var replacement = viewValue.replace(',', '.');
          //allow - as single char
          if('-' === replacement) {
            ctrl.$setValidity('float', false);
            return viewValue;
          } else {
            ctrl.$setValidity('float', true);
            return parseFloat(replacement);
          }
        } else {
          ctrl.$setValidity('float', false);
          return undefined;
        }
      });
    }
  };
});

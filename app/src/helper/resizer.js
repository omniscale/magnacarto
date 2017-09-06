// inspirited by https://stackoverflow.com/a/22253161
angular.module('magna-app')
.directive('resizer', ['$document', '$window', function($document, $window) {
  return function($scope, $element, $attrs) {
    var resizerMax = parseInt($attrs.resizerMax);
    var resizerMin = parseInt($attrs.resizerMin);
    var resizerWidth = parseInt($attrs.resizerWidth);
    var leftElement, rightElement;

    $element.on('mousedown', function(event) {
      event.preventDefault();

      leftElement = angular.element(document.getElementById($attrs.resizerLeftId));
      rightElement = angular.element(document.getElementById($attrs.resizerRightId));

      $document.on('mousemove', mousemove);
      $document.on('mouseup', mouseup);

      $element.addClass('active');
    });

    function mousemove(event) {
      var x = event.pageX;
      if (resizerMax && x > resizerMax) {
        x = resizerMax;
      } else if (resizerMin && x < resizerMin) {
        x = resizerMin;
      }

      $element.css({
        left: x + 'px'
      });

      leftElement.css({
        width: x + 'px',
        marginLeft: (-x) + 'px',
        left: x + 'px'
      });

      rightElement.css({
        paddingLeft: (x + resizerWidth) + 'px'
      });
    }

    function mouseup() {
      $element.removeClass('active');

      angular.element($window).triggerHandler('resize');

      $document.unbind('mousemove', mousemove);
      $document.unbind('mouseup', mouseup);

      leftElement = undefined;
      rightElement = undefined;
    }
  };
}]);

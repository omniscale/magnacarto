// inspirited by https://stackoverflow.com/a/22253161
angular.module('magna-app')
.directive('resizer', ['$document', '$window', function($document, $window) {
  return function($scope, $element, $attrs) {
    var resizerMax = parseInt($attrs.resizerMax);
    var resizerMin = parseInt($attrs.resizerMin);
    var resizerWidth = parseInt($attrs.resizerWidth);
    var leftElements, rightElements;

    var findElements = function(idsString) {
      var elements = [];
      var elementIds = idsString.split(',');
      angular.forEach(elementIds, function(id) {
        var domElement = document.getElementById(id);
        if(domElement !== null) {
          elements.push(domElement);
        }
      });
      return angular.element(elements);
    };

    $element.on('mousedown', function(event) {
      event.preventDefault();

      leftElements = findElements($attrs.resizerLeftIds);
      rightElements = findElements($attrs.resizerRightIds);

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

      leftElements.css({
        width: x + 'px',
        marginLeft: (-x) + 'px',
        left: x + 'px'
      });

      rightElements.css({
        paddingLeft: (x + resizerWidth) + 'px'
      });
    }

    function mouseup() {
      $element.removeClass('active');

      $document.unbind('mousemove', mousemove);
      $document.unbind('mouseup', mouseup);

      leftElements = undefined;
      rightElements = undefined;
    }
  };
}]);

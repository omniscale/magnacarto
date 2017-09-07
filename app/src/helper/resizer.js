// inspirited by https://stackoverflow.com/a/22253161
angular.module('magna-app')
.directive('resizer', ['$document', '$timeout', function($document, $timeout) {
  return function($scope, $element, $attrs) {
    var resizerMax = parseInt($attrs.resizerMax);
    var resizerMin = parseInt($attrs.resizerMin);
    var resizerWidth = parseInt($attrs.resizerWidth);
    var leftElements, rightElements;
    var disabled = false;
    var storedX;

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

    var updateElements = function(x) {
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

      storedX = x;
    };

    var mousemove = function(event) {
      updateElements(event.pageX);
    };

    var mouseup = function() {
      $element.removeClass('active');

      $document.unbind('mousemove', mousemove);
      $document.unbind('mouseup', mouseup);
    };

    $element.on('mousedown', function(event) {
      event.preventDefault();

      $document.on('mousemove', mousemove);
      $document.on('mouseup', mouseup);

      $element.addClass('active');
    });

    var enableResizer = function() {
      $element.removeClass('hide');
      if(angular.isDefined(storedX)) {
        updateElements(storedX);
      }
      disabled = false;
    };

    var disableResizer = function() {
      $element.addClass('hide');
      leftElements.css({
        width: '',
        marginLeft: '',
        left: ''
      });
      rightElements.css({
        paddingLeft: ''
      });
      disabled = true;
    };

    $attrs.$observe('resizerDisabled', function(n, o) {
      if(n === 'false' && disabled) {
        enableResizer();
        return;
      }
      if(n === 'true' && !disabled) {
        disableResizer();
        return;
      }
    });

    // get left-/rightElements after dom rendered
    $timeout(function() {
      leftElements = findElements($attrs.resizerLeftIds);
      rightElements = findElements($attrs.resizerRightIds);
    });
  };
}]);

// inspirited by https://stackoverflow.com/a/22253161
angular.module('magna-app')
.directive('resizer', ['$document', '$timeout', function($document, $timeout) {
  return {
    scope: {
      resizerActualSize: '='
    },
    link: function($scope, $element, $attrs) {
      var resizerMax = parseInt($attrs.resizerMax);
      var resizerMin = parseInt($attrs.resizerMin);
      var resizerWidth = parseInt($attrs.resizerWidth);
      var leftElements, rightElements, bottomElements, topElements;
      var disabled = false;

      var horizontal = $attrs.resizer === 'horizontal';

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

      var _updateElementsVertical = function(x) {
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
          left: (x + resizerWidth) + 'px'
        });

        $scope.resizerActualSize = x;
      };

      var _updateElementsHorizontal = function(y) {
        if (resizerMax && y > resizerMax) {
          y = resizerMax;
        } else if (resizerMin && y < resizerMin) {
          y = resizerMin;
        }

        $element.css({
          bottom: (y - resizerWidth) + 'px'
        });
        topElements.css({
          bottom: y + 'px'
        });
        bottomElements.css({
          height: y + 'px'
        });
        $scope.resizerActualSize = y;
      };

      var updateElements = horizontal ? _updateElementsHorizontal : _updateElementsVertical;

      var _mousemoveVertical = function(event) {
        updateElements(event.pageX);
      };

      var _mousemoveHorizontal = function(event) {
        updateElements(window.innerHeight - event.pageY);
      };

      var mousemove = horizontal ? _mousemoveHorizontal : _mousemoveVertical;

      var mouseup = function() {
        $element.removeClass('active');

        $document.unbind('mousemove', mousemove);
        $document.unbind('mouseup', mouseup);
        $scope.$apply();
      };

      $element.on('mousedown', function(event) {
        event.preventDefault();

        $document.on('mousemove', mousemove);
        $document.on('mouseup', mouseup);

        $element.addClass('active');
      });

      var enableResizer = function() {
        $element.removeClass('hide');
        if(angular.isDefined($scope.resizerActualSize)) {
          updateElements($scope.resizerActualSize);
        }
        disabled = false;
      };

      var _disableResizerVertical = function() {
        $element.addClass('hide');
        leftElements.css({
          width: '',
          marginLeft: '',
          left: ''
        });
        rightElements.css({
          left: ''
        });
        disabled = true;
      };

      var _disableResizerHorizontal = function() {
        $element.addClass('hide');
        leftElements.css({
          bottom: ''
        });
        rightElements.css({
          height: ''
        });
        disabled = true;
      };

      var disableResizer = horizontal ? _disableResizerHorizontal : _disableResizerVertical;

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

      // get elements after dom rendered
      $timeout(function() {
        if(angular.isDefined($attrs.resizerLeftIds)) {
          leftElements = findElements($attrs.resizerLeftIds);
        }
        if(angular.isDefined($attrs.resizerRightIds)) {
          rightElements = findElements($attrs.resizerRightIds);
        }
        if(angular.isDefined($attrs.resizerBottomIds)) {
          bottomElements = findElements($attrs.resizerBottomIds);
        }
        if(angular.isDefined($attrs.resizerTopIds)) {
          topElements = findElements($attrs.resizerTopIds);
        }
      });

      $scope.$watch('resizerActualSize', function(n, o) {
        if(disabled === true) {
          return;
        }
        if(angular.isUndefined(o) && angular.isDefined(n) && n > -1) {
          updateElements(n);
        }
      });
    }
  };
}]);

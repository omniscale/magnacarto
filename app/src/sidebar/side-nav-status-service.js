angular.module('magna-app')

.factory('SideNavStatusService', [function() {
  var hideLayers, hideStyles;

  var reset = function() {
    hideLayers = false;
    hideStyles = true;
  };
  reset();
  return {
    hideLayers: function(val) {
      if(val !== undefined) { hideLayers = val; }
      return hideLayers;
    },
    hideStyles: function(val) {
      if(val !== undefined) { hideStyles = val; }
      return hideStyles;
    },
    reset: reset
  };
}]);

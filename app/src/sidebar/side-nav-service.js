angular.module('magna-app')

.factory('SideNavService', [function() {
  var hideLayers, hideStyles, currentPage;

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
    currentPage: function(val) {
      if(val !== undefined) { currentPage = val; }
      return currentPage;
    },
    reset: reset
  };
}]);

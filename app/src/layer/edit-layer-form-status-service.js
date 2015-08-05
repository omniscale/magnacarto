angular.module('magna-app')

.factory('EditLayerFormStatusService', [function() {
  var hideGeneral, hideExtentSRS, hideDatasource;

  var reset = function() {
    hideGeneral = false;
    hideExtentSRS = true;
    hideDatasource = false;
  };
  reset();
  return {
    hideGeneral: function(val) {
      if(val !== undefined) { hideGeneral = val; }
      return hideGeneral;
    },
    hideExtentSRS: function(val) {
      if(val !== undefined) { hideExtentSRS = val; }
      return hideExtentSRS;
    },
    hideDatasource: function(val) {
      if(val !== undefined) { hideDatasource = val; }
      return hideDatasource;
    },
    reset: reset
  };
}]);

angular.module('magna-app')

.directive('mss', ['$timeout', 'dashboard',
  function() {
    return {
    restrict: 'E',
    replace: true,
    controller: function($scope, $rootScope, dashboard) {
	    $scope.mss = dashboard.mss;
	    var selection = [];
	    angular.copy(dashboard.mss, selection);

	    $scope.selection = selection;

	    $scope.toggleSelection = function toggleSelection(style) {
	      var idx = $scope.selection.indexOf(style);
	      if (idx > -1) {
	        $scope.selection.splice(idx, 1);
	      }
	      else {
	        $scope.selection.push(style);
	      }
	      $rootScope.$broadcast('imageUpdateMss', $scope.selection);
	    };

    },
    templateUrl: '/app/dashboard/mss-template.html'
    // link: function(scope, element, attrs) {
   	// 	console.log(element)
    // }
  };
}]);
angular.module('magna-app')

// found at http://stackoverflow.com/questions/24039226/angularjs-format-text-return-from-json-to-camel-case
.filter('titleCase', function() {
  return function(input) {
    input = input || '';
    return input.replace(/\w\S*/g, function(txt){
      return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase();
    });
  };
});

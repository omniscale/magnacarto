angular.module('magna-app')

.provider('StyleService', [function() {
  this.$get = ['$rootScope', '$timeout',
    function($rootScope, $timeout) {
      var StyleServiceInstance = function() {
        var self = this;
        // stores all styles from project folder
        self.styles = [];
        // stores active project styles
        self.projectStyles = [];
        // stores styles in active list (for easy check only (see inActiveStyles))
        self.knownStyles = [];
        // the active style list
        self.activeStyles = [];

        $rootScope.$watchCollection(function() {
          return self.activeStyles;
        }, function() {
          self.updateProjectStyles();
        });
      };

      StyleServiceInstance.prototype.setStyles = function(styles) {
        this.styles = styles;
      };

      StyleServiceInstance.prototype.setProjectStyles = function(styles) {
        var self = this;
        var done = false;
        self.projectStyles = styles;

        angular.forEach(self.projectStyles, function(style, idx) {
          // handle new mss styles
          if(!self.inActiveStyles(style)) {
            // initial push into activeStyles
            if(self.activeStyles.length === 0) {
              self.activeStyles.push({
                active: true,
                style: style
              });
            } else {
              // look for last submitted style
              var prevStyle = idx > 0 ? styles[idx -1] : false;
              done = false;
              angular.forEach(self.activeStyles, function(activeStyle, activeStyleIdx) {
                if(!done) {
                  // add style after last added style
                  if(!prevStyle || activeStyle.style === prevStyle) {
                    self.activeStyles.splice(activeStyleIdx + 1, 0, {
                      active: true,
                      style: style
                    });
                    done = true;
                  }
                }
              });
            }
            // add style to known styles
            self.knownStyles.push(style);
          } else {
            done = false;
            // look for style and activate it
            angular.forEach(self.activeStyles, function(activeStyle) {
              if(!done) {
                if(style === activeStyle.style) {
                  activeStyle.active = true;
                  done = true;
                }
              }
            });
          }
        });
        // deactivate all styles not longer in projectStyles
        angular.forEach(self.activeStyles, function(styleObj) {
          if(self.projectStyles.indexOf(styleObj.style) === -1) {
            styleObj.active = false;
          }
        });
      };

      StyleServiceInstance.prototype.toggleStyle = function(style) {
        var self = this;
        // toggle status of style
        if(self.inActiveStyles(style)) {
          var done = false;
          angular.forEach(self.activeStyles, function(styleObj) {
            if(!done && styleObj.style === style) {
              styleObj.active = !styleObj.active;
              done = true;
            }
          });
        // or add it
        } else {
          self.activeStyles.push({
            style: style,
            active: true
          });
          self.knownStyles.push(style);
        }
        self.updateProjectStyles();
      };

      StyleServiceInstance.prototype.updateProjectStyles = function() {
        var self = this;
        self.projectStyles.length = 0;
        angular.forEach(self.activeStyles, function(styleObj) {
          if(styleObj.active) {
            self.projectStyles.push(styleObj.style);
          }
        });
      };

      StyleServiceInstance.prototype.inActiveStyles = function(style) {
        return this.knownStyles.indexOf(style) > -1;
      };

      return new StyleServiceInstance();
    }];
}]);

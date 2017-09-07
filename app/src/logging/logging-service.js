angular.module('magna-app')
/* Todo rename to ProjectServicev */
.provider('LoggingService', [function() {
  this.$get = ['$rootScope', 'ProjectService', function($rootScope, ProjectService) {
    var LoggingInstance = function() {
        var self = this;
        this.messages = [];
        this.lastSuccessfulUpdate = undefined;
        this.lastSuccessfulUpdateIdx = undefined;
        this.newMessageCallbacks = [];
        // Add messages handler when socket changes
        $rootScope.$watch(function() {
          return ProjectService.getSocket();
        }, function(n, o) {
          if(n === o || n === undefined) return;
          var socket = n;
          socket.$on('$open', function() {
            self.appendMessage('info', ['Connect to the websocket Server']);
          });

          socket.$on('$message', self.handleMessage.bind(self));
        });
    };

    LoggingInstance.prototype.addNewMessageCallback = function(callback) {
        this.newMessageCallbacks.push(callback);
    };

    LoggingInstance.prototype.appendMessage = function(type, msg) {
        var newMsg = {
          type: type,
          msg: msg,
          time: new Date()
        };
        if(newMsg.type === 'success') {
            this.lastSuccessfulUpdate = newMsg;
        }
        this.messages.unshift(newMsg);
        this.lastSuccessfulUpdateIdx = this.messages.indexOf(this.lastSuccessfulUpdate);
        angular.forEach(this.newMessageCallbacks, function(callback) {
            callback(newMsg);
        });
    };

    LoggingInstance.prototype.handleMessage = function(resp) {
        var type, msg;
        if(resp === undefined) {
            return;
        } else if(resp.error !== undefined) {
            type = 'danger';
            msg = [];

            if (resp.warnings !== undefined) {
                type = 'warning';
                if (resp.error !== '') {
                    type = 'danger';
                    msg.push('Error: ' + resp.error);
                }
                angular.forEach(resp.warnings, function(warning) {
                    if (warning) {
                        msg.push(warning);
                    }
                });

                if (msg.length === 0) {
                    return;
                }

            } else if (resp.filename !== undefined) {
                msg.push('Error in ' + resp.filename + ':');
                msg.push(resp.error);
            } else {
                msg.push('Error: ' + resp.error);
            }
            if (resp.files !== undefined) {
                angular.forEach(resp.files, function(file) {
                    msg.push('• ' + file);
                });
            }
        } else if(resp.updated_at !== undefined) {
            type = 'success';
            msg = ['Updated'];
        } else {
            return;
        }
        this.appendMessage(type, msg);
    };

    return new LoggingInstance();
  }];
}]);

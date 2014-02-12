var pollApp = angular.module('pollApp', [
  'ngRoute',
  'pollControllers'
]);
 
pollApp.config(['$routeProvider', '$locationProvider',
  function($routeProvider, $locationProvider) {
    $locationProvider.html5Mode(true);
    $routeProvider.
      when('/', {
        templateUrl: '/static/partials/index.html',
        controller: 'IndexCtrl'
      }).
      when('/:pollId', {
        templateUrl: '/static/partials/poll.html',
        controller: 'PollCtrl'
      }).
      otherwise({
        redirectTo: '/'
      });
  }]);

pollApp.factory('VoteStreamService', function() {
  var service = {};
 
  service.connect = function(pollId) {
    if(service.ws) { return; }
 
    var l = window.location;
    var url = ((l.protocol === "https:") ? "wss://" : "ws://") +
      l.hostname + (((l.port != 80) && (l.port != 443)) ? ":" + l.port : "") +
      "/api/poll/" + pollId + "/stream";
    var ws = new WebSocket(url);
 
    ws.onopen = function() {
    };
 
    ws.onerror = function() {
    }
 
    ws.onmessage = function(message) {
      service.callback(message.data);
    };
 
    service.ws = ws;
  }
 
  service.send = function(message) {
    service.ws.send(message);
  }
 
  service.subscribe = function(callback) {
    service.callback = callback;
  }
 
  return service;
});

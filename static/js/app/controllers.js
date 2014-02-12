var pollControllers = angular.module('pollControllers', []);
 
pollControllers.controller('IndexCtrl', ['$scope', 
  function ($scope) {
  }]);
 
pollControllers.controller('PollCtrl', ['$scope', '$routeParams', 'VoteStreamService',
  function($scope, $routeParams, VoteStreamService) {
    $scope.pollId = $routeParams.pollId;

    VoteStreamService.subscribe(function(message) {
    	$scope.counts = message;
    	$scope.$apply();
    });

    VoteStreamService.connect($routeParams.pollId);
  }]);
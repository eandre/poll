var pollControllers = angular.module('pollControllers', []);
 
pollControllers.controller('IndexCtrl', ['$scope', '$route', '$http', '$location',
  function ($scope, $route, $http, $location) {
    $scope.$route = $route;
    $scope.answers = [];
    $scope.numAnswers = 0;
    $scope.minAnswers = 2;
    $scope.maxAnswers = 9;

    for ( var i = 0; i < $scope.maxAnswers; i++ ) {
      $scope.answers.push({});
    }

    $scope.updateAnswers = function() {
      for ( var i = 0; i < $scope.maxAnswers; i++ ) {
        var answer = $scope.answers[i];
        var hasText = (typeof answer.text !== "undefined" && answer.text !== "");
        if ( i < ($scope.maxAnswers-1) && (hasText || i < $scope.minAnswers) ) {
          $scope.answers[i].visible = true;
          $scope.answers[i+1].visible = true;

          if ( hasText ) {
            $scope.numAnswers = i+1;
          }
        }
      }
    }

    $scope.create = function() {
      var answers = [];
      for ( var i = 0; i < $scope.answers.length; i++ ) {
        var answer = $scope.answers[i];
        if ( answer.visible && typeof answer.text !== "undefined" && answer.text !== "" )  {
          answers.push(answer.text);
        }
      }

      var data = $.param({
        question: $scope.question,
        answer: answers,
        multipleChoice: $scope.multipleChoice
      });
      $http({
        method: 'POST',
        url: "/api/poll/",
        data: data,
        headers: {'Content-Type': 'application/x-www-form-urlencoded'},
      }).then(function(data) {
        $location.path("/" + data.data);
      }, function(error) {
        $scope.submitError = error;
      });
    }

    $scope.updateAnswers();
  }]);

pollControllers.controller('AboutCtrl', ['$scope', '$route', 
  function ($scope, $route) {
    $scope.$route = $route;
  }]);

pollControllers.controller('PollCtrl', ['$scope', '$routeParams', '$http', '$location', '$filter',
  function($scope, $routeParams, $http, $location, $filter) {
    // Get poll info
    $scope.pollId = $routeParams.pollId;
    $scope.selectedAnswers = [];

    $http({method: 'GET', url: "/api/poll/" + $scope.pollId + "/"})
      .then(function(result) {
        $scope.poll = processPoll(result['data']);
        $scope.poll.selectedAnswer = -1;
      }, function(error) {
        $location.path("/");
      });

      $scope.updateSelection = function() {
        $scope.selectedAnswers = $filter('filter')($scope.poll.answers, {checked: true});
      }

      $scope.vote = function() {
        var answers = [];
        if ( $scope.poll.multipleChoice ) {
          for ( var i = 0; i < $scope.selectedAnswers.length; i++ ) {
            answers.push($scope.selectedAnswers[i].index + 1);
          }
        } else {
          answers.push($scope.poll.selectedAnswer);
        }

        var data = $.param({answer: answers});
        $http({
          method: 'POST',
          url: "/api/poll/" + $scope.pollId + "/",
          data: data,
          headers: {'Content-Type': 'application/x-www-form-urlencoded'},
        }).then(function(data) {
          $scope.showResults();
        }, function(error) {
          $scope.voteError = error;
        });
      }

      $scope.showResults = function() {
        $location.path("/" + $scope.pollId + "/results");
      }
  }]);
 
pollControllers.controller('ResultsCtrl', ['$scope', '$routeParams', '$http',
  '$location', 'VoteStreamService',
  function($scope, $routeParams, $http, $location, VoteStreamService) {
    // Get poll info
    $scope.pollId = $routeParams.pollId;
    $http({method: 'GET', url: "/api/poll/" + $scope.pollId + "/"})
      .then(function(result) {
        $scope.poll = result['data'];
        updateCounts($scope.poll.counts);
      }, function(error) {
        $location.path("/");
      });

    // Vote ratio
    $scope.voteRatio = function(index) {
      var sum = $scope.poll.countSum;
      if ( sum == 0 ) {
        return 0;
      }
      return $scope.poll.counts[index] / sum;
    }

    function updateCounts(counts) {
      var winningCount = -1;
      var sum = 0;
      for ( var i = 0; i < counts.length; i++ ) {
        var count = counts[i];
        if ( count > winningCount ) {
          $scope.poll.winningIndex = i;
          winningCount = count;
        } else if ( count === winningCount ) {
          $scope.poll.winningIndex = -1;
        }
        sum += count;
      }

      $scope.poll.countSum = sum;
      $scope.poll.counts = counts;
    }

    // Vote streaming
    VoteStreamService.subscribe(function(message) {
      updateCounts(message.data);
    	$scope.$apply();
    });

    VoteStreamService.connect($scope.pollId);
  }]);

function processPoll(data) {
  var poll = {
    question: data.question,
    answers: [],
    multipleChoice: data.multipleChoice
  };
  for ( var i = 0; i < data.answers.length; i++ ) {
    poll.answers.push({text: data.answers[i], count: data.counts[i], index: i});
  }
  return poll;
}

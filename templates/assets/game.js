var gameFetchInterval;

function getGame() {
  fetch('/api/games/' + gameId)
    .then(function(response) {
      return response.json();
    })
    .then((game) => {
      if (game.gameState.ended === true) {
        clearInterval(gameFetchInterval);
      }

      blueGoals = game.gameState.blueGoals;
      redGoals = game.gameState.redGoals;

      document.getElementById('blue-goals').innerText = blueGoals;
      document.getElementById('red-goals').innerText = redGoals;
    });
}

gameFetchInterval = setInterval(getGame, 2500);

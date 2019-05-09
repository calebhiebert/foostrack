// Some variables in this file are declared in ../game.html

var gameFetchInterval;

function saveScroll() {
  localStorage.setItem('scroll', document.scrollingElement.scrollTop);
}

function restoreScroll() {
  if (localStorage.getItem('scroll') !== null) {
    document.scrollingElement.scrollTo(0, localStorage.getItem('scroll'));
    localStorage.removeItem('scroll');
  }
}

function getGame() {
  fetch('/api/games/' + gameId + '/eventcount')
    .then(function(response) {
      return response.json();
    })
    .then((ec) => {
      if (eventCount !== ec.eventCount) {
        saveScroll();
        location.reload();
      }
    });
}

restoreScroll();

if (!gameEnded) {
  gameFetchInterval = setInterval(getGame, 2000);
  getGame();
}

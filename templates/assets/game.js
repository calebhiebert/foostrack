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
        .then(function (response) {
            return response.json();
        })
        .then((ec) => {
            if (eventCount !== ec.eventCount) {
                saveScroll();
                location.reload();
            }
        });
}

function initCharts() {
    // Setup chart stuff
    var goalsChartCTX = document.getElementById('goals-chart').getContext('2d');
    var timingChartCTX = document.getElementById('timing-chart').getContext('2d');

    userGoals = userGoals.sort((a, b) => {
        if (a.user.id > b.user.id) {
            return 1;
        } else if (a.user.id < b.user.id) {
            return -1;
        } else {
            return 0;
        }
    });

    var goalsChart = new Chart(goalsChartCTX, {
        type: 'bar',
        data: {
            labels: userGoals.map((ug) => ug.user.username),
            datasets: [
                {
                    data: userGoals.map((ug) => ug.antigoals),
                    backgroundColor: 'hsl(348, 100%, 61%)',
                    label: 'Antigoals',
                },
                {
                    data: userGoals.map((ug) => ug.goals),
                    backgroundColor: 'hsl(141, 71%, 48%)',
                    label: 'Goals',
                },
            ],
        },
        options: {
            animation: false,
            title: {
                display: true,
                text: 'Goals',
            },
            responsive: true,
            scales: {
                xAxes: [
                    {
                        stacked: true,
                    },
                ],
                yAxes: [
                    {
                        stacked: true,
                    },
                ],
            },
        },
    });

    var timingChart = new Chart(timingChartCTX, {
        type: 'line',
        data: {
            labels: events.slice(4, -1).map((e, idx) => ''),
            datasets: [
                {
                    data: events.slice(4, -1).map((e) => e.elapsed / 1000000000),
                    label: 'Goal Delay',
                    backgroundColor: '#209cee',
                    pointRadius: 0,
                    borderWidth: 0,
                },
            ],
        },
        options: {
            animation: false,
        },
    });
}

restoreScroll();

window.addEventListener('scroll', () => {
    saveScroll();
});

if (!gameEnded) {
    gameFetchInterval = setInterval(getGame, 2000);
    getGame();
}

if (eventCount > 4) {
    initCharts();
}
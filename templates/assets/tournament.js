// Some variables in this file are declared in ../tournament.html

// How "Dark" a color is before white text is displayed over top instead of black
const LUMINANCE_TOLERANCE = 0.43;

// When making team gradients, how much the hue is shifted to make the second color
const GRADIENT_CHANGE = -30;

(function() {
  teams.forEach(function(team) {
    var card = document.getElementById(`tm-${team.id}`);

    if (card) {
      var color = chroma(team.color);
      var useWhite = color.luminance() < LUMINANCE_TOLERANCE;

      card.style.background = `linear-gradient(90deg, ${color.css()} 0%, ${getSecondaryColor(color).css()} 100%)`;

      if (useWhite) {
        // Change Card Title
        var title = card.querySelector('.card-header > .card-header-title');
        title.classList.add('has-text-white');

        // Change card actions
        var actions = card.querySelectorAll('.card-footer > .card-footer-item');

        actions.forEach((a) => {
          a.classList.add('has-text-white');
        });
      } else {
        // Change card actions
        var actions = card.querySelectorAll('.card-footer > .card-footer-item');

        actions.forEach((a) => {
          a.classList.add('has-text-dark');
        });
      }
    }
  });

  var bracketLevels = splitBracketLevels();

  var bracketChartCTX = document.getElementById('bracket-chart').getContext('2d');

  var bracketChart = new Chart(bracketChartCTX, {
    type: 'doughnut',
    data: {
      datasets: bracketLevels.map(bl => {
        return {
          data: bl.map(() => 1),
          backgroundColor: bl.map((bt) => chroma(bt.team.color).css()),
          hoverBackgroundColor: bl.map((bt) => getSecondaryColor(bt.team.color).css()),
        };
      }),
      labels: (bracketLevels[0] || []).map(bt => bt.team.name),
    },
    options: {
      cutoutPercentage: 10,
      gridLines: {
        display: false,
      },
      responsive: true,
      tooltips: {
        callbacks: {
          label: function(tti, data) {
            var bl = bracketLevels[tti.datasetIndex][tti.index];

            if (bl.gameId) {
              return 'Game ' + bl.gameId;
            } else {
              return 'Game Not Created';
            }
          },

          title: function(tti, data) {
            var bl = bracketLevels[tti[0].datasetIndex][tti[0].index];

            return bl.team.name;
          }
        }
      }
    },
  });

  document.getElementById('bracket-chart').onclick = (e) => {
    var ele = bracketChart.getElementAtEvent(e)[0];

    if (ele) {
      var bl = bracketLevels[ele._datasetIndex][ele._index];

      if (bl.gameId) {
        window.location = `/game/${bl.gameId}`
      }
    }
  }
})();

function splitBracketLevels() {
  var bracketLevels = [];

  bracketPositions.forEach(bp => {
    if (!bracketLevels[bp.bracketLevel]) {
      bracketLevels[bp.bracketLevel] = [];
    }

    bracketLevels[bp.bracketLevel].push(bp);
  });

  return bracketLevels;
}

function getSecondaryColor(primary) {
  var secondary = chroma(primary).hsl();

  if (isNaN(secondary[0])) {
    secondary[0] = 0;
  }

  secondary[0] += GRADIENT_CHANGE;

  return chroma.hsl(secondary[0], secondary[1], secondary[2])
}
// Some variables in this file are declared in ../tournament.html

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

  var cutoutPercentage = 0;

  switch (bracketLevels.length) {
    case 1:
      cutoutPercentage = 80;
      break;
    case 2:
      cutoutPercentage = 60;
      break;
    case 3:
      cutoutPercentage = 40;
      break;
    case 4:
      cutoutPercentage = 20;
      break;
    default:
      cutoutPercentage = 0;
  }

  var bracketChart = new Chart(bracketChartCTX, {
    type: 'doughnut',
    data: {
      datasets: bracketLevels.map((bl, idx) => {

        // Make the last circle in the set have no white border
        // this gets around some ugly artifacts that chart.js creates
        var borderColor = bl.map(bt => {
          if (idx == bracketLevels.length - 1) {
            var w = bracketChartCTX.canvas.width;
            var h = bracketChartCTX.canvas.height;
  
            var graident = bracketChartCTX.createLinearGradient(0.5 * w, 0, 0.5 * w, h);
  
            graident.addColorStop(0, chroma(bt.team.color).css());
            graident.addColorStop(1, getSecondaryColor(bt.team.color).css())
  
            return graident;
          } else {
            return '#FFFFFF'
          }
        });

        var backgroundColor = bl.map(bt => {
          var w = bracketChartCTX.canvas.width;
          var h = bracketChartCTX.canvas.height;

          var graident = bracketChartCTX.createLinearGradient(0.5 * w, 0, 0.5 * w, h);

          graident.addColorStop(0, chroma(bt.team.color).css());
          graident.addColorStop(1, getSecondaryColor(bt.team.color).css())

          return graident;
        })

        var weight = 0.5;

        // If the bar is the last in the circle, make it larger
        if (idx == bracketLevels.length - 1) {
          weight = 1;
        }

        return {
          data: bl.map(() => 1),
          backgroundColor: backgroundColor,
          borderColor: borderColor,
          weight: weight,
        };
      }),
      labels: (bracketLevels[0] || []).map((bt) => bt.team.name),
    },
    options: {
      cutoutPercentage: bracketLevels[bracketLevels.length - 1].length == 1 ? 0 : cutoutPercentage,
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
          },
        },
      },
    },
  });

  document.getElementById('bracket-chart').onclick = (e) => {
    var ele = bracketChart.getElementAtEvent(e)[0];

    if (ele) {
      var bl = bracketLevels[ele._datasetIndex][ele._index];

      if (bl.gameId) {
        window.location = `/game/${bl.gameId}`;
      }
    }
  };
})();

function splitBracketLevels() {
  var bracketLevels = [];

  bracketPositions.forEach((bp) => {
    if (!bracketLevels[bp.bracketLevel]) {
      bracketLevels[bp.bracketLevel] = [];
    }

    bracketLevels[bp.bracketLevel].push(bp);
  });

  return bracketLevels;
}

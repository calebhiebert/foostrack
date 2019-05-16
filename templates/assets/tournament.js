// Some variables in this file are declared in ../tournament.html

(function() {
  console.log(teams);

  teams.forEach(function(team) {
    var card = document.getElementById(`tm-${team.id}`);

    if (card) {
      var color = chroma(team.color);
      var useWhite = color.luminance() < 0.4;

      var color2 = color.hsl();

      if (isNaN(color2[0])) {
        color2[0] = 0;
      }

      color2[0] -= 50;

      card.style.background = `linear-gradient(90deg, ${color.css()} 0%, ${chroma.hsl(color2[0], color2[1], color2[2]).css()} 100%)`;

      if (useWhite) {

        // Change Card Title
        var title = card.querySelector('.card-header > .card-header-title');
        title.classList.add('has-text-white');

        // Change card actions
        var actions = card.querySelectorAll('.card-footer > .card-footer-item');

        actions.forEach(a => {
          a.classList.add('has-text-white');
        })
      }
    }
  });
})();

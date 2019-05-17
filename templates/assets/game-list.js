
(function() {
  teams.forEach(function(team) {
    var boxes = document.querySelectorAll(`.tm-${team.ID}`);

    boxes.forEach(b => {
      var color = chroma(team.color);
      var useWhite = color.luminance() < LUMINANCE_TOLERANCE;
  
      b.style.background = `linear-gradient(90deg, ${color.css()} 0%, ${getSecondaryColor(color).css()} 100%)`;

      if (useWhite) {
        b.classList.add('has-text-white');
      }
    });
  });
})();
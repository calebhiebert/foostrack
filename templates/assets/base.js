// How "Dark" a color is before white text is displayed over top instead of black
const LUMINANCE_TOLERANCE = 0.43;

// When making team gradients, how much the hue is shifted to make the second color
const GRADIENT_CHANGE = -30;

function getSecondaryColor(primary) {
  var secondary = chroma(primary).hsl();

  if (isNaN(secondary[0])) {
    secondary[0] = 0;
  }

  secondary[0] += GRADIENT_CHANGE;

  return chroma.hsl(secondary[0], secondary[1], secondary[2])
}
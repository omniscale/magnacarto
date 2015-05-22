#test::line {
    line-width: 1;
}

#test::arrow {
  marker-type: arrow;
  marker-line-width: 1;
  marker-line-color: blue;
  marker-fill: red;
  marker-placement: line;
  marker-spacing: 100;
  [id=1] {
    marker-opacity: 0.5;
    marker-spacing: 200;
    marker-transform: "scale(0.8) rotate(180)";
  }
}

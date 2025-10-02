// only zoom=15 has a line-width

#roads[zoom>=14] {
  line-color: white;
  [type='primary'] {
    // this should create a rule for level 15 with line-width: 5
    line-color: red;
  }
}

#roads[zoom=15] {
  line-width: 5;
}


#roads[zoom>=14] {
  line-cap: round;
}

#roads {
  line-join: bevel;
}
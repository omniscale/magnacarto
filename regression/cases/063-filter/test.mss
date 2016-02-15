Map { background-color: white; }

#test[name =~ '.*match inside.*'] {
  marker-type: ellipse;
  marker-line-width: 1;
  marker-line-color: blue;
  marker-fill: blue;
}

#test[name=~"Starts with.*"] {
  marker-type: ellipse;
  marker-line-width: 1;
  marker-line-color: green;
  marker-fill: green;
}

#test[name =~ 'Point [0-9]$'] {
  marker-type: ellipse;
  marker-line-width: 1;
  marker-line-color: red;
  marker-fill: red;
}

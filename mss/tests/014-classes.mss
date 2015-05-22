/* Applies to all layers with .land class */

#lakes {
  line-color: #f00;
  line-width: 0.5;
  polygon-fill: #0f0;
  // Applies to #lakes.land
  .land {
    polygon-fill: #00f;
  }
}



.water {
  polygon-fill: #0f0;
  line-width: 1;
  .basin {
    polygon-fill: #fff;
    polygon-opacity: 0.5;
  }
}

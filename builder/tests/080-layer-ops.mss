#labels {
  text-name: [name];
  text-size: 8;

  // test layer based properties, only work for Mapnik
  opacity: 50%;
  comp-op: screen;
  

  image-filters: "invert(), color-blind-deuteranope()";
  direct-image-filters: "invert()";
}

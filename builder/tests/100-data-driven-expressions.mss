#labels_field {
    text-name: [name];
    text-size: "2+6";
}

#labels_field2 {
    text-name: [name];
    text-size: "[size]";
}

#labels_roads_refs {
  [type='motorway'] {
    shield-name: "[ref]";
    shield-size: "[size]";
    shield-face-name: "DejaVu Sans Book";
    shield-fill: "[color]";
    shield-file: url("img/rail-[width].svg");
    shield-placement: line;
  }
}

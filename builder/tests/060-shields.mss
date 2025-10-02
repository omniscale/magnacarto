Map {
    background-color: #fff;
}

#roads[type='motorway'] {
    line-width: 1;
}

#labels_roads_refs {
  [type='motorway'] {
    shield-name: "[ref]";
    shield-size: 8;
    shield-face-name: "DejaVu Sans Book";
    shield-fill: #eee;
    shield-file: url(img/rail-24.svg);
    shield-avoid-edges: true;
    shield-clip: false;
    shield-placement: line;
    /*shield-allow-overlap: true;*/
    shield-min-distance: 250;
    shield-min-padding: 50;
    shield-spacing: 250;
    [reflen=5]  { shield-file: url(img/rail-24.svg); }
    [reflen>=6] { shield-file: url(img/rail-24.svg); }
  }
}

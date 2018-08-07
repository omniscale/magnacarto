Map { background-color: white; }

#test::line {
    line-width: 1;
}
#test::shield{
    shield-file: url(shield-19x11-bbb-f0c900.svg);
    shield-name: "[id]";
    shield-size: 11;
    shield-face-name: "Noto Sans Regular";
    shield-fill: black;
    // shield-avoid-edges: true;
    // shield-clip: false;
    shield-placement: line;
    // shield-min-distance: 5;
    // shield-min-padding: 50;
    // [id=1] { marker-opacity: .5; }

    [id=2] { shield-repeat-distance: 20; shield-spacing: 20; }
    [id=3] { shield-repeat-distance: 30; shield-spacing: 30; }
    [id=4] { shield-repeat-distance: 40; shield-spacing: 40; }

}

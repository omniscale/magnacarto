Map { background-color: white; }

#test{
    marker-width: 20;
    marker-height: 20;
    marker-allow-overlap: true;
    // top left feature
    [id=1] {
        marker-file: url(cross.svg);
    }
    // top right feature
    [id=2] {
        marker-file: url(cross.svg);
        marker-transform: "rotate(45.0, 0, 0)";
    }
    // bottom left feature
    [id=3] {
        marker-file: url(cross.svg);
        marker-transform: "translate(0.000, -10.000) scale(1.000) rotate(45.000, 0.000, 10.000)";
    }
    // bottom right feature
    [id=4] {
        marker-file: url(cross.svg);
        marker-transform: "translate(0.000000, -10.000000) scale(1.00000) rotate(180.000000, 0.000000, 10.000000)";
    }
}

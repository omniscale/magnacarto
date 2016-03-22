Map { background-color: white; }

#test{
    marker-width: 23;
    marker-height: 40;
    [id=1] {
        marker-file: url(arrow-down.svg);
        marker-transform: "translate(0.000000, -20.000000) scale(1.000000) rotate(45.000000, 0.000000, 20.000000)";
    }

    [id=2] {
        marker-file: url(arrow-down.svg);
        marker-transform: "translate(0.000000, -30.000000) scale(1.500000) rotate(270.000000, 0.000000, 20.000000)";
    }
    [id=3] {
        marker-file: url(arrow-up.svg);
        marker-transform: "translate(0.000000, 0.000000) scale(1.000000) rotate(45.000000, 0.000000, 0.000000)";
    }
    [id=4] {
        marker-file: url(arrow-up.svg);
        marker-transform: "translate(0.000000, 0.000000) scale(1.500000) rotate(-90.000000, 0.000000, 0.000000)";
    }
}

Map { background-color: white; }

#test::outline{
    line-width: 16;
    line-color: #555;
    line-dasharray: 10, 20, 5, 30;

    [id=1] {
        line-cap: round;
        line-join: round;
    }
    [id=2] {
        line-cap: square;
        line-join: miter;
    }
    [id=3] {
        line-cap: butt;
        line-join: bevel;
    }
}


#test::inline {
    line-width: 12;
    line-color: #f0b300;
    line-dasharray: 10, 20, 5, 30;

    [id=1] {
        line-cap: round;
        line-join: round;
    }
    [id=2] {
        line-cap: square;
        line-join: miter;
    }
    [id=3] {
        line-cap: butt;
        line-join: bevel;
    }
}
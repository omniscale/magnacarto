#test[type="cemetery"] {
    line-color: #26c600;
    line-width: 8;
    polygon-fill: rgba(185, 222, 0, 0.6);

}

#test[type="residential"] {
    polygon-fill: rgba(255, 133, 86, 0.6000);
    line-color: #d53427;
    line-width: 8;

    [id=2] {
        line-cap: square;
    }
}

#test[type="building"] {
    polygon-fill: rgba(136, 153, 136, 0.6000);
    line-color: #444;
    line-width: 0.5;
    line-join: round;
}


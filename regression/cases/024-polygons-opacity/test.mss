Map { background-color: white; }

#test[type="cemetery"] {
    line-color: #26c600;
    line-width: 8;
    polygon-fill: rgba(185, 222, 0, 0.6);

}

#test[type="residential"] {
    polygon-fill: rgba(255, 133, 86, 0.6);
    line-color: #d53427;
    line-width: 8;
    line-opacity: 0.3;

    [id=2] {
        polygon-fill: rgb(255, 133, 86);
        polygon-opacity: 0.6;
        line-cap: square;
    }
}

#test[type="building"] {
    polygon-fill: rgba(136, 153, 136, 0.6000);
    line-color: #444;
    line-opacity: 0.3;
    line-width: 0.5;
    line-join: round;
}


Map { background-color: #c2dae6;}

@land: #ede8c8;

#world {
    line-color: black;
    line-width: 1;
    polygon-fill: @land;
    [ISO_A2='DE'] {
        polygon-fill: darken(@land, 10%);
    }
}

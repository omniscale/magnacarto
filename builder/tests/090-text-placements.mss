@foo: 2;

#labels {
    text-name: "[name]";
    text-fill: black;
    text-size: 12;

    text-placement-list: {
        text-dx: 6+3;
    },{
        text-size: @foo * 3;
        text-name: "[nm]"
    };
}

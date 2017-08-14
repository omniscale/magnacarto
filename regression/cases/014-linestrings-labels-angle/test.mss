Map { background-color: white; }

#test::line{
    line-width: 1;
}

#test::label{
    text-name: [name];
    text-size: 14;
    text-halo-radius: 2;
    text-halo-fill: #fff;

    text-face-name: "Noto Sans Regular";
    text-max-char-angle-delta: 43; // only the most curved line should not be labeled
    text-placement: line;
}

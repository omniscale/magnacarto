Map { background-color: white; }

#test::label{
    text-name: '';
    text-size: 14;
    text-halo-radius: 2;
    text-halo-fill: #fff;

    text-face-name: "Noto Sans Regular";
    text-placement: point;

    [id=1] { text-name: [name] + "-" + [id]; }
    [id=2] { text-name: '|' + [name] + '|'; } // TODO | not supported by mapserver
    [id=3] { text-name: "{" + [name] + "}"; }
}

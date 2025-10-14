@foo: 2;

#test::label{
    text-name: [name];
    text-size: 14;
    text-halo-radius: 2;
    text-halo-fill: #fff;

    text-justify-alignment: center;

    text-face-name: "Noto Sans Regular";

    text-placement: point;

    [elev != 0] {
        text-name: [name] + "\n" <Format size="10">  "(" + [elev] + ")" </Format>;
        [id=1] {
            text-name: <Format fill="#ff0000">[name]</Format> "\n" <Format size="10">  "(" + [elev] + ")" </Format>;
        }
    }
}

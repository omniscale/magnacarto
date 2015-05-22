@foo: 1;
@bar: 12+12;
@bar: -12;
@bar: -12*12;
@bar: 12/2;
@bar: @foo * 12;
@baz: 3 * ( 1 + 2 );
@foo: #ddd;
@bar: fadein(@foo, 12/2);
@bar: fadein(@foo, 12%);

#class {
    line-join: miter;
    text-face-name: "foo", "bar";
}

@fonts: "foo", "bar";
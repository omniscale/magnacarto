@foo: #a03;
@bar: lighten(rgba(100, 50, 75, 150), 20);

#func        { line-width: 1; line-color: fadeout(@foo, 10%); }
#funcnested  { line-width: 1; line-color: fadeout(lighten(@foo, 10), 20); }
#funcfunc    { line-width: 1; line-color: fadeout(@bar, 12); }


@num: 12;
@hash: #66cCFf;
@hash2: #66C;
@rgb: rgb(102, 0, 255);
@rgbpercent: rgb(40%, 0%, 100%);
@rgba: rgba(0, 255, 102, 0.4);
@rgbacompat: rgba(0, 255, 102, 102);
@rgbapercent: rgba(0, 100%, 40%, 40%);
@list: "Foo", "Bar", "Baz";
@listnum: 2, 3, 4;

#num            { line-width: @num; }
#hash           { line-color: @hash; line-width: 1; }
#hash2          { line-color: @hash2; line-width: 1; }
#rgb            { line-color: @rgb; line-width: 1; }
#rgbpercent     { line-color: @rgbpercent; line-width: 1; }
#rgba           { line-color: @rgba; line-width: 1; }
#rgbacompat     { line-color: @rgbacompat ; line-width: 1; }
#rgbapercent    { line-color: @rgbapercent; line-width: 1; }
#list           { text-face-name: @list; text-size: 12; text-name: "foo"; }
#listnum        { line-dasharray: @listnum; line-width: 1;}


@motorway_width: 3;
@trunk_width: 3;
@primary_width: 3;
@outline_width: 1;

#roads::outline[zoom>=12] {
    line-color: black;
}

#roads::inline[zoom>=12] {
    line-color: rgba(0, 0, 0, 0.5);
    [type='motorway'] { line-color: red; }
    [type='trunk'] { line-color: yellow; }
    [type='primary'] { line-color: beige; }
    [type='secondary'] { line-color: white; }
}

#roads::outline[zoom>=12] {
    [type='motorway'] { line-width: @motorway_width + @outline_width; }
}

#roads::inline[zoom>=12] {
    [type='motorway'] { line-width: @motorway_width; }
}

#roads::outline[zoom>=14] {
    [type='motorway'] { line-width: @motorway_width + @outline_width; }
    [type='trunk'] { line-width: @trunk_width + @outline_width; }
    [type='primary'],
    [type='secondary'] { line-width: @primary_width + @outline_width; }
}

#roads::inline[zoom>=14] {
    line-width: 0.5;
    [type='motorway'] { line-width: @motorway_width; }
    [type='trunk'] { line-width: @trunk_width; }
    [type='primary'],
    [type='secondary'] { line-width: @primary_width; }
}
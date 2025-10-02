#foo[zoom=11] {
    line-width: 11;
}


#bar[zoom>=11][zoom<=14] {
    line-width: 12;
    [zoom>=15] {
        line-width: 15;
    }
}

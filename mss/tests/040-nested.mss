#roads{
  [type='rail'] {
    [service='yard'] {
      line-color: red;
    }
  }
}

#roads[zoom=17] {
  line-width: 2;
  line-color: yellow;

  [type='rail'] {
    line-width: 5;
    // service='yard' should be red
  }
}

#roads {
  line-width: 1;
}
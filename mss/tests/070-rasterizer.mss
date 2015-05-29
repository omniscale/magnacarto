#hillshade {
  raster-opacity: 1;
  raster-scaling: lanczos;
  raster-comp-op: grain-merge;
}
#slope {
  raster-comp-op: grain-extract;
  raster-opacity: 1;
  raster-scaling: lanczos;
  raster-colorizer-default-mode: linear;
  raster-colorizer-default-color: transparent;
  raster-colorizer-stops:
    stop(0,#fff)
    stop(90, #000)
}
#dem {
  raster-opacity: 1;
  raster-scaling: lanczos;
  raster-colorizer-default-mode: linear;
  raster-colorizer-default-color: transparent;
  raster-colorizer-stops:
    stop(0,#47443e)
    stop(50, #77654a)
    stop(100, rgb(85,107,50))
    stop(200, rgb(187, 187, 120))
    stop(255, rgb(217,222,170));
  raster-comp-op: color-dodge;
}
Magnacarto Examples
===================

Simple Worldmap
---------------

Download the following zip file, unpack and move the `ne_10m_admin_0_countries` directory into this directory (`examples`).

* <http://www.naturalearthdata.com/http//www.naturalearthdata.com/download/10m/cultural/ne_10m_admin_0_countries.zip>


### world.mml

`world.mml` defines the datasources for all layers. In this case it only contains a single `world` layer that uses the NaturalEarth shapefile.

### world.mss

`world.mss` contains the actual map styling.


### Mapnik

Create a Mapnik XML:

    magnacarto -mml world.mml  > world.xml


Render map with nik2img:

    nik2img.py -d 1000 1000 world.xml mapnik.png


### MapServer

Create MapServer map file:

    magnacarto -mml world.mml -builder mapserver > world.map

Render map with shp2img:

    shp2img -m world.map -o mapserver.png -s 1000 1000


Use the -ms-no-map-block option to create a mapfile without a MAP block.
The output will only contain LAYERS and SYMBOLS, useful for `INCLUDE`ing into map files with custom metadata, image formats, etc.:

    magnacarto -mml world.mml -builder mapserver -ms-no-map-block > layers.map


OSM Roads
---------

Download the following zip file and unpack everything into `docs/examples/hamburg_germany`.

* <https://s3.amazonaws.com/metro-extracts.mapzen.com/hamburg_germany.imposm-shapefiles.zip>
    (OSM extract from https://mapzen.com/data/metro-extracts)


### roads.mml

`roads.mml` defines a single layer based on the roads shapefile.

### roads.mss

Styling that demonstrates a few CartoCSS features:

- variables to define road widths
- arithmetics to calculate the width of outlines (`@motorway_width + @outline_width`)
- attachments to render a layer multiple times (`::outline` first, then `::inline`)
- work with zoom levels to define scale dependent styles
- define default values (`line-width` for all roads from zoom level 14)

### Mapnik

Create a Mapnik XML:

    magnacarto -mml roads.mml > roads.xml

Render map with nik2img:

    nik2img.py -d 500 500 -s 3857 -e 1120261 7086018 1122707 7088464 roads.xml mapnik.png


### MapServer

Create MapServer map file:

    magnacarto -mml roads.mml -builder mapserver > roads.map

Render map with shp2img:

    shp2img -m roads.map -o mapserver.png -s 500 500  -e 1120261 7086018 1122707 7088464

Use the -ms-no-map-block option to create a mapfile without a MAP block.
The output will only contain LAYERS and SYMBOLS, useful for `INCLUDE`ing into map files with custom metadata, image formats, etc.:

    magnacarto -mml roads.mml -builder mapserver -ms-no-map-block > layers.map


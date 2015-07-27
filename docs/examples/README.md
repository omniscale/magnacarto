Magnacarto Examples
===================

Simple Worldmap
---------------

Download the following zip file, unpack and move the `ne_10m_admin_0_countries` directory into this directory (`examples`).

* http://www.naturalearthdata.com/http//www.naturalearthdata.com/download/10m/cultural/ne_10m_admin_0_countries.zip


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
